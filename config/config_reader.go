package config

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
)

const (
	cfEmailEnvName  = "CLOUDFLARE_EMAIL"
	cfAPIKeyEnvName = "CLOUDFLARE_DNS_API_TOKEN"
)

type Configtype struct {
	Db  DbType  `json:"db"`
	TLS TLSType `json:"tls"`
}

type DialectType string

func (d *DbType) AsGormConnection() gorm.Dialector {
	var dialect gorm.Dialector

	switch d.Dialect {
	case "sqlite3":
		dialect = sqlite.Open(d.Connection)
	default:
		log.Fatalf("Unkown dialect [%s]", d.Dialect)
	}

	return dialect
}

type DbType struct {
	Dialect    DialectType `json:"dialect"`
	Username   string      `json:"username"`
	Password   string      `json:"password"`
	Connection string      `json:"connection"`
	DBName     string      `json:"dbName"`
}

type TLSType struct {
	Local       bool     `json:"local"`
	CertDir     string   `json:"certDir"`
	CertDomains []string `json:"certDomains"`
	CertEmail   string   `json:"certEmail"`
	DNSProvider string   `json:"dnsProvider"`
	DNSAPIToken string   `json:"dnsApiToken"`
}

type DbDriver string

var Config *Configtype

func MakeConfig(configFile string) {
	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		panic("config file " + configFile + " does not exist")
	}
	configString, err := ioutil.ReadFile(configFile)

	if err != nil {
		panic(err)
	}

	appConfig := Configtype{}

	json.Unmarshal(configString, &appConfig)

	Config = &appConfig

	if !Config.TLS.Local {
		setEnvVars()
	}
}

func setEnvVars() {
	os.Setenv(cfEmailEnvName, Config.TLS.CertEmail)
	os.Setenv(cfAPIKeyEnvName, Config.TLS.DNSAPIToken)
}

func GetDBConfig() DbType {
	return Config.Db
}

func GetTLSConfig() TLSType {
	return Config.TLS
}
