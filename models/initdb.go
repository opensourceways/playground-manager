package models

import (
	_ "database/sql"
	"os"

	_ "github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

//InitDb init database
func Initdb() bool {
	BConfig, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		logs.Error("config init error:", err.Error())
		return false
	}
	//ConnDb()
	dbhost := BConfig.String("mysql::dbhost")
	dbport := BConfig.String("mysql::dbport")
	dbuser := BConfig.String("mysql::dbuser")
	dbname := BConfig.String("mysql::dbname")
	dbpwd := BConfig.String("mysql::dbpwd")
	if os.Getenv("DB_NAME") != "" {
		dbname = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_PSWD") != "" {
		dbpwd = os.Getenv("DB_PSWD")
	}
	maxidle, lerr := BConfig.Int("mysql::maxidle")
	if lerr != nil {
		maxidle = 30
	}

	maxconn, lerr := BConfig.Int("mysql::maxconn")
	if lerr != nil {
		maxconn = 3000
	}
	dns := dbuser + ":" + dbpwd + "@tcp(" + dbhost + ":" + dbport + ")/" + dbname + "?charset=utf8"
	errx := orm.RegisterDriver("mysql", orm.DRMySQL)
	if errx != nil {
		logs.Error("RegisterDriver, orm err: ", errx)
		return false
	}
	errorm := orm.RegisterDataBase("default", "mysql", dns, maxidle, maxconn)
	if errorm != nil {
		logs.Error("RegisterDataBase failed", "errorm: ", errorm)
		return false
	}
	// orm.Debug = true
	logs.Info("Initdb, mysql connection is successful")
	res := CreateDb()
	if res {
		logs.Info("mysql table init success!")
	} else {
		logs.Error("mysql table init failed!")
		return false
	}
	return true
}
