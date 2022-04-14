package controllers

import (
	"github.com/astaxie/beego"
)

type HealthzLiveController struct {
	beego.Controller
}

func (c *HealthzLiveController) RetSaData(resp map[string]interface{}) {
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Get liveness
// @Description get liveness
// @router / [get]
func (u *HealthzLiveController) Get() {
	resp := make(map[string]interface{})
	resp["code"] = 200
	resp["errmsg"] = "success"
	resp["body"] = ""
	defer u.RetSaData(resp)
	return
}

type HealthzReadController struct {
	beego.Controller
}

func (c *HealthzReadController) RetSaData(resp map[string]interface{}) {
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Get readlines
// @Description get readlines
// @router / [get]
func (u *HealthzReadController) Get() {
	resp := make(map[string]interface{})
	resp["code"] = 200
	resp["errmsg"] = "success"
	resp["body"] = ""
	defer u.RetSaData(resp)
	return
}

// @Get test
// @Description get test
// @router /test [get]
func (u *HealthzLiveController) GetTest() {
	resp := make(map[string]interface{})
	resp["code"] = 200
	resp["errmsg"] = "test"
	resp["body"] = "3"
	defer u.RetSaData(resp)
	return
}
