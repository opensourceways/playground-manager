package routers

import (
	"net/http"
	"playground_backend/controllers"
	"playground_backend/handler"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	//InsertFilter是提供一个过滤函数

	// Get callback request link
	beego.Router("/playground/oauth2/callback/links", &controllers.Oauth2CallBackLinksControllers{})
	// authing callback request link,Get callback result
	// beego.Router("/playground/oauth2/callback", &controllers.Oauth2CallBackControllers{})
	// User authorization and authentication, obtain user information
	beego.Router("/playground/oauth2/authentication", &controllers.Oauth2AuthenticationControllers{})
	// Get user information after successful login(Obtain user information after authorization)
	beego.Router("/playground/oauth2/checklogin", &controllers.UserInfoControllers{}, "get:CheckLogin")
	beego.Router("/playground/user/information", &controllers.UserInfoControllers{}, "get:GetCurrentUser")
	beego.Router("/playground/oauth2/callback", &controllers.UserInfoControllers{}, "get:AuthingCallback")
	// The user creates crd resources and returns the result of creating resources
	beego.Router("/playground/crd/resource", &controllers.CrdResourceControllers{})
	beego.Router("/playground/crd/makeKubeconfig", &controllers.CrdResourceControllers{}, "post:MakeKubeconfig")
	beego.Router("/playground/users/checkSubdomain", &controllers.CrdResourceControllers{}, "post:CheckSubdomain")
	// Bind the course/chapter selected by the user
	beego.Router("/playground/users/course/chapter", &controllers.CourseChapterControllers{})
	// Health check interface
	beego.Router("/healthz/readiness", &controllers.HealthzReadController{})
	beego.Router("/healthz/liveness", &controllers.HealthzLiveController{})
	beego.InsertFilter("/*", beego.BeforeRouter, corsFunc)
	beego.InsertFilter("/playground/crd/*", beego.BeforeRouter, handler.Authorize)
	beego.InsertFilter("/playground/users/*", beego.BeforeRouter, handler.Authorize)
	beego.InsertFilter("/playground/user/*", beego.BeforeRouter, handler.Authorize)
}

var success = []byte("SUPPORT OPTIONS")

var corsFunc = func(ctx *context.Context) {
	origin := ctx.Input.Header("Origin")
	ctx.Output.Header("Access-Control-Allow-Methods", "OPTIONS,DELETE,POST,GET,PUT,PATCH")
	ctx.Output.Header("Access-Control-Max-Age", "3600")
	ctx.Output.Header("Access-Control-Allow-Headers", "X-Custom-Header,accept,Content-Type,Access-Token,Authorization,token")
	ctx.Output.Header("Access-Control-Allow-Credentials", "true")
	ctx.Output.Header("Access-Control-Allow-Origin", origin)
	if ctx.Input.Method() == http.MethodOptions {
		// options请求，返回200
		ctx.Output.SetStatus(http.StatusOK)
		_ = ctx.Output.Body(success)
	}
}
