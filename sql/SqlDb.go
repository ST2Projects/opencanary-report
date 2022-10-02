package sql

import (
	log "github.com/sirupsen/logrus"
	"github.com/st2projects/opencanary-report/config"
	"github.com/st2projects/opencanary-report/model/api"
	"github.com/st2projects/opencanary-report/model/db"
	_ "gorm.io/driver/sqlite" // Import sqlite3 driver
	"gorm.io/gorm"
)

var dbConnection *gorm.DB

func Connect() {
	dbConfig := config.GetDBConfig()

	connectionString := dbConfig.Connection

	dbConnection, _ = gorm.Open(dbConfig.AsGormConnection(), &gorm.Config{})
	log.Infof("Created connection to %s", connectionString)

	initTables()
}

func AddEvent(event *api.Event) {

	entry := db.Entry{
		SourceIP: event.SrcHost,
		DstPort:  event.DstPort,
		Password: event.Logdata.Password,
		Username: event.Logdata.Username,
		UTCTime:  event.UtcTime,
	}

	dbConnection.Create(&entry)
}

func GetEvents(count int) []db.Entry {

	var entries []db.Entry

	dbConnection.Limit(count).Find(&entries)

	return entries
}

func initTables() {

	err := dbConnection.AutoMigrate(&db.Entry{})

	if err != nil {
		log.Fatalf("Failed to perform migration: [%s]", err.Error())
	}
}
