package db

import (
	"gorm.io/gorm"
)

type Entry struct {
	gorm.Model
	ID       int    `db:"id" json:"id"`
	SourceIP string `db:"source_ip"`
	DstPort  int    `db:"dest_port"`
	Password string `db:"password"`
	Username string `db:"username"`
	UTCTime  string `db:"utc_time"`
}

func (u *Entry) Table() string {
	return "entry"
}

func (u *Entry) IDColumn() string {
	return "id"
}
