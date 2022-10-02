package server

import (
	"encoding/json"
	"fmt"
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"github.com/st2projects/opencanary-report/config"
	"github.com/st2projects/opencanary-report/helper"
	cmdModel "github.com/st2projects/opencanary-report/model"
	"github.com/st2projects/opencanary-report/model/api"
	"github.com/st2projects/opencanary-report/model/db"
	"github.com/st2projects/opencanary-report/sql"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"os"
	"time"
)

const contentTypeKey = "content-type"
const jsonContentType = "application/json"

type Dashboard struct {
	Description string     `json:"description"`
	Events      []db.Entry `json:"events"`
}

func EventHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)

	if err != nil {
		panic(helper.NewError("Failed to read event %s", err))
	}

	event := &api.Event{}
	err = json.Unmarshal(body, event)

	if err != nil {
		panic(helper.NewError("Failed to unmarshall event %s", err))
	}

	sql.AddEvent(event)
	writer.WriteHeader(http.StatusOK)
}

func RootHandler(writer http.ResponseWriter, request *http.Request) {
	dash := Dashboard{Description: "Last 10 events", Events: sql.GetEvents(10)}
	writer.Header().Set(contentTypeKey, jsonContentType)
	writer.WriteHeader(http.StatusOK)
	responseEncoder := json.NewEncoder(writer)
	responseEncoder.Encode(dash)
}

func Serve(httpConfig *cmdModel.HTTPConfig) {

	var (
		certReloader *simplecert.CertReloader
		err          error
		numRenews    int
		ctx, cancel  = context.WithCancel(context.Background())

		// init strict tlsConfig (this will enforce the use of modern TLS configurations)
		// you could use a less strict configuration if you have a customer facing web application that has visitors with old browsers
		tlsConf = tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

		// a simple constructor for a http.Server with our Handler
		makeServer = func() *http.Server {
			return &http.Server{
				Addr:      fmt.Sprintf("0.0.0.0:%d", httpConfig.HttpsPort),
				Handler:   makeRouter(),
				TLSConfig: tlsConf,
			}
		}

		// init server
		srv = makeServer()

		// init simplecert configuration
		cfg = simplecert.Default
	)

	configuredTls := config.GetTLSConfig()
	cfg.Local = configuredTls.Local
	cfg.CacheDir = "./resources"
	cfg.Domains = configuredTls.CertDomains
	cfg.SSLEmail = configuredTls.CertEmail
	cfg.DNSProvider = configuredTls.DNSProvider
	cfg.HTTPAddress = ""
	cfg.TLSAddress = ""

	cfg.WillRenewCertificate = func() {
		cancel()
	}

	cfg.DidRenewCertificate = func() {
		numRenews++
		// Restart the server
		ctx, cancel = context.WithCancel(context.Background())
		srv = makeServer()

		// Force reload the cert
		certReloader.ReloadNow()

		go serve(ctx, srv)
	}

	certReloader, err = simplecert.Init(cfg, func() {
		os.Exit(0)
	})

	if err != nil {
		log.Fatalf("Simple cert init failed: %s\n", err)
	}

	// Redirect 80 -> 443
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", httpConfig.HttpPort), http.HandlerFunc(simplecert.Redirect))

	tlsConf.GetCertificate = certReloader.GetCertificateFunc()
	log.Infof("Serving at https://%s:%d", configuredTls.CertDomains[0], httpConfig.HttpsPort)
	serve(ctx, srv)
	<-make(chan bool)
}

func makeRouter() *mux.Router {
	commonHandlers := alice.New(LoggingHandler, ErrorHandler)

	router := mux.NewRouter()

	router.Handle("/", commonHandlers.ThenFunc(RootHandler))
	router.Handle("/ping", commonHandlers.ThenFunc(PingHandler))
	router.Handle("/event", commonHandlers.ThenFunc(EventHandler))
	return router
}

func serve(ctx context.Context, srv *http.Server) {
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen %s\n", err)
		}
	}()

	log.Info("Server started")
	<-ctx.Done()
	log.Info("Server stopped")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5+time.Second)
	defer func() {
		cancel()
	}()

	err := srv.Shutdown(ctxShutdown)
	if err == http.ErrServerClosed {
		log.Info("Server stopped correctly")
	} else {
		log.Errorf("Error when stopping server %s\n", err)
	}
}
