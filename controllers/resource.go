package controllers

import (
	"encoding/base64"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"playground_backend/handler"
	"playground_backend/models"
)

type CrdResourceControllers struct {
	beego.Controller
}

type RequestParameter struct {
	ResourceId   string `json:"resourceId"`
	TemplatePath string `json:"templatePath"`
	UserId       int64  `json:"userId"`
	ContactEmail string `json:"contactEmail"`
	Token        string `json:"token"`
}

func (c *CrdResourceControllers) RetData(resp ResData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

type ResData struct {
	ResInfo handler.ResResourceInfo `json:"instanceInfo"`
	Mesg    string                  `json:"message"`
	Code    int                     `json:"code"`
}

// @Title CreateCrdResource
// @Description CreateCrdResource
// @Param	body		body 	models.CreateCrdResource	true		"body for user content"
// @Success 200 {int} models.CreateCrdResource
// @Failure 403 body is empty
// @router / [post]
func (u *CrdResourceControllers) Post() {
	var rp RequestParameter
	var resData ResData
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	logs.Info("gitee oauth2 request parameters: ", string(u.Ctx.Input.RequestBody))
	json.Unmarshal(u.Ctx.Input.RequestBody, &rp)
	if len(rp.TemplatePath) < 1 || rp.UserId < 1 || len(rp.ResourceId) < 1 {
		resData.Code = 400
		resData.Mesg = "Request parameter error"
		u.RetData(resData)
		return
	}
	if len(rp.Token) < 1 {
		resData.Code = 401
		resData.Mesg = "Request parameter error"
		u.RetData(resData)
		return
	} else {
		gui := models.GiteeUserInfo{AccessToken: rp.Token, UserId: rp.UserId}
		ok := handler.CheckToken(&gui)
		if !ok {
			resData.Mesg = "Request parameter error"
			resData.Code = 403
			u.RetData(resData)
			return
		}
	}
	var rri = new(handler.ResResourceInfo)
	rr := handler.ReqResource{EnvResource: rp.TemplatePath, UserId: rp.UserId, ContactEmail: rp.ContactEmail}
	handler.CreateEnvResourc(rr, rri, rp.ResourceId)
	if rri.UserId > 0 {
		if rri.Status == 0 {
			resData.Code = 202
		} else {
			resData.Code = 200
		}
		resData.ResInfo = *rri
		resData.Mesg = "success"
	} else {
		resData.ResInfo = *rri
		resData.Code = 501
		resData.Mesg = "Failed to create resource, need to request resource again"
	}
	u.RetData(resData)
	return
}

// @Title Get GetCrdResource
// @Description get GetCrdResource
// @Param	status	int	true (0,1,2)
// @Success 200 {object} GetCrdResource
// @Failure 403 :status is err
// @router / [get]
func (u *CrdResourceControllers) Get() {
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
	envResource := u.GetString("templatePath")
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
