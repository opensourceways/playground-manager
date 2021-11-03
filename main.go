package main

import (
	"github.com/astaxie/beego"
	"playground_backend/common"
	"playground_backend/models"
	_ "playground_backend/routers"
)

func init() {
	// Initialization log
	common.LogInit()
}


func main() {
	// init db
	dbOk := models.Initdb()
	if !dbOk {
		println("error: Database initialization failed")
		return
	}
	//common.ReadFileToEntry()
	beego.Run()
}

