package qbsql

import (
	"database/sql"
	"qqbot/lib/qblog"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Db    *sql.DB = nil
	dberr error
)

func InitDB() error {
	if Db != nil {
		if Db.Ping() != nil {
			Db.Close()
		} else {
			return nil
		}
	}
	dsn := "root:koishi@tcp(127.0.0.1:3306)/koishi?charset=utf8"
	Db, dberr = sql.Open("mysql", dsn)
	if dberr != nil {
		qblog.Log.Error("open mysql connect error:", dberr)
		return dberr
	}
	err := Db.Ping()
	if err != nil {
		qblog.Log.Error("ping mysql error:", err)
	}
	Db.SetConnMaxLifetime(time.Second * 60)
	Db.SetMaxOpenConns(20)
	Db.SetMaxIdleConns(10)
	return nil
}
