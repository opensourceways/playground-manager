package routers

import (
	"playground_backend/controllers"

	"github.com/astaxie/beego"
)

func init() {
	// Get callback request link
	beego.Router("/playground/oauth2/callback/links", &controllers.Oauth2CallBackLinksControllers{})
	// authing callback request link,Get callback result
	beego.Router("/playground/oauth2/callback", &controllers.Oauth2CallBackControllers{})
	// User authorization and authentication, obtain user information
	beego.Router("/playground/oauth2/authentication", &controllers.Oauth2AuthenticationControllers{})
	// Get user information after successful login(Obtain user information after authorization)
	beego.Router("/playground/user/information", &controllers.UserInfoControllers{})
	// The user creates crd resources and returns the result of creating resources
	beego.Router("/playground/crd/resource", &controllers.CrdResourceControllers{})
	beego.Router("/playground/users/checkSubdomain", &controllers.CrdResourceControllers{}, "post:CheckSubdomain")
	// Bind the course/chapter selected by the user
	beego.Router("/playground/users/course/chapter", &controllers.CourseChapterControllers{})
	// Health check interface
	beego.Router("/healthz/readiness", &controllers.HealthzReadController{})
	beego.Router("/healthz/liveness", &controllers.HealthzLiveController{})
}
