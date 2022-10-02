package server

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type ErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

func ErrorHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// There is surely a better way to do this
				errorMsg := fmt.Sprintf("%s", err)
				response := ErrorResp{Code: http.StatusInternalServerError, Message: errorMsg}
				json.NewEncoder(w).Encode(response)
			}
		}()
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func PingHandler(writer http.ResponseWriter, _ *http.Request) {

	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte(fmt.Sprintf("Pong\nTime now is %s", time.Now().Format("2006-01-02 15:04:05"))))
	if err != nil {
		log.Errorf("Failed to write ping response %s", err.Error())
	}
}
