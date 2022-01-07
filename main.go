package main

import (
	"github.com/astaxie/beego"
	"playground_backend/common"
	"playground_backend/controllers"
	"playground_backend/handler"
	"playground_backend/models"
	_ "playground_backend/routers"
	"playground_backend/task"
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
	// 1. Initialize memory resources
	handler.InitialMemoryRes()
	handler.InitialResourcePool()
	// Initialize a scheduled task
	taskOk := task.InitTask()
	if !taskOk {
		println("error: Timing task initialization failed, the program ends")
		task.StopTask()
		return
	}
	// single run
	task.StartTask()
	defer task.StopTask()
	//common.ReadFileToEntry()
	//common.DesString("")
	beego.ErrorController(&controllers.ErrorController{})
	beego.Run()
}

