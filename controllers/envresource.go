package controllers

import (
	"encoding/base64"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"playground_backend/handler"
	"playground_backend/models"
)

type CreateEnvResControllers struct {
	beego.Controller
}

type CreateResource struct {
	ResourceId string `json:"resourceId"`
	EnvResource  string `json:"envResource"`
	UserId       int64  `json:"userId"`
	ContactEmail string `json:"contactEmail"`
	Token        string `json:"token"`
}

func (c *CreateEnvResControllers) RetData(resp ResData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

type ResData struct {
	ResInfo handler.ResResourceInfo `json:"resInfo"`
	Mesg    string                  `json:"message"`
	Code    int                     `json:"code"`
}

// @Title GiteeOauth2
// @Description GiteeOauth2
// @Param	body		body 	models.GiteeOauth2	true		"body for user content"
// @Success 200 {int} models.GiteeOauth2
// @Failure 403 body is empty
// @router / [post]
func (u *CreateEnvResControllers) Post() {
	var cres CreateResource
	var resData ResData
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	logs.Info("gitee oauth2 request parameters: ", string(u.Ctx.Input.RequestBody))
	json.Unmarshal(u.Ctx.Input.RequestBody, &cres)
	if len(cres.EnvResource) < 1 || cres.UserId < 1 || len(cres.ResourceId) < 1 {
		resData.Code = 404
		resData.Mesg = "Request parameter is empty"
		u.RetData(resData)
		return
	}
	if len(cres.Token) < 1 {
		resData.Code = 403
		resData.Mesg = "Request parameter is empty"
		u.RetData(resData)
		return
	} else {
		gui := models.GiteeUserInfo{AccessToken: cres.Token, UserId: cres.UserId}
		ok := handler.CheckToken(&gui)
		if !ok {
			resData.Mesg = "Request parameter error"
			resData.Code = 405
			u.RetData(resData)
			return
		}
	}
	var rri = new(handler.ResResourceInfo)
	rr := handler.ReqResource{EnvResource: cres.EnvResource, UserId: cres.UserId, ContactEmail: cres.ContactEmail}
	handler.CreateEnvResourc(rr, rri, cres.ResourceId)
	resData.ResInfo = *rri
	resData.Code = 200
	resData.Mesg = "success"
	u.RetData(resData)
	return
}

// @Title Get CreateEnvResControllers
// @Description get CreateEnvResControllers
// @Param	status	int	true (0,1,2)
// @Success 200 {object} CreateEnvResControllers
// @Failure 403 :status is err
// @router / [get]
func (u *CreateEnvResControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	var resData ResData
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	token := u.GetString("token")
	userId, _ := u.GetInt64("userId", 0)
	if userId == 0 {
		resData.Mesg = "Request parameter error"
		resData.Code = 404
		u.RetData(resData)
		return
	}
	envResource := u.GetString("envResource")
	if len(envResource) == 0 {
		resData.Mesg = "Request parameter error"
		resData.Code = 405
		u.RetData(resData)
		return
	}
	resourceId := u.GetString("resourceId")
	if len(resourceId) == 0 {
		resData.Mesg = "Request parameter error"
		resData.Code = 405
		u.RetData(resData)
		return
	}
	envRes, err := base64.StdEncoding.DecodeString(envResource)
	if err != nil {
		resData.Mesg = "Request parameter error"
		resData.Code = 406
		u.RetData(resData)
		return
	}
	if token == "" {
		resData.Mesg = "Request parameter error"
		resData.Code = 403
		u.RetData(resData)
		return
	} else {
		var rri = new(handler.ResResourceInfo)
		rr := handler.ReqResource{EnvResource: string(envRes), UserId: userId}
		handler.GetEnvResourc(rr, rri, resourceId)
		resData.ResInfo = *rri
		resData.Code = 200
		resData.Mesg = "success"
		u.RetData(resData)
	}
	return
}
