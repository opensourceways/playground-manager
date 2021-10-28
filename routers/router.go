package routers

import (
	"playground_backend/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
	// Get callback request link
	beego.Router("/playground/gitee/callbackurl", &controllers.GiteeCallBackUrlControllers{})
    // gitee callback request link
	beego.Router("/playground/gitee/callback", &controllers.GiteeCallBackControllers{})
    // User authorization and authentication, obtain user information
	beego.Router("/playground/gitee/oauth2", &controllers.GiteeOauth2Controllers{})
    // Get user information of gitee
	beego.Router("/playground/gitee/user/info", &controllers.GiteeUserInfoControllers{})
    // User create environment resource interface
	beego.Router("/playground/create/environment/resource", &controllers.CreateEnvResControllers{})
}
