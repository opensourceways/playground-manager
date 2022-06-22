package controllers

import "github.com/astaxie/beego"

type ErrorController struct {
	beego.Controller
}

func (c *ErrorController) Error404() {
	c.Data["content"] = "您访问的地址或者方法不存在"
	c.TplName = "error/404.tpl"
}
