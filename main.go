package main

import (
	"fmt"
	"playground_backend/common"
	"playground_backend/controllers"
	"playground_backend/handler"
	"playground_backend/models"
	_ "playground_backend/routers"
	"playground_backend/task"

	"github.com/astaxie/beego"
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
	fmt.Println("--------------NewCoursePool------")
	handler.NewCoursePool(0)
	fmt.Println("----------------InitialResourcePool------")
	handler.InitialResourcePool()
	// Initialize a scheduled task
	fmt.Println("----------------InitTask------")

	taskOk := task.InitTask()
	if !taskOk {
		println("error: Timing task initialization failed, the program ends")
		task.StopTask()
		return
	}
	fmt.Println("----------------StartTask------")

	// single run
	task.StartTask()
	defer task.StopTask()
	beego.ErrorController(&controllers.ErrorController{})
	fmt.Println("----------------beego.Run------")
	beego.Run()
}
