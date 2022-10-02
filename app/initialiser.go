package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/st2projects/opencanary-report/config"
	"github.com/st2projects/opencanary-report/model"
	"github.com/st2projects/opencanary-report/server"
	"github.com/st2projects/opencanary-report/sql"
)

func InitialiseApp(configPath string, httpConfig *model.HTTPConfig) {

	customLogFormat := new(log.TextFormatter)
	customLogFormat.TimestampFormat = "2022-01-01 01:01:01.123"
	customLogFormat.FullTimestamp = true

	log.SetFormatter(customLogFormat)

	log.Info("Starting Sentinel service")
	config.MakeConfig(configPath)
	sql.Connect()

	server.Serve(httpConfig)
}
