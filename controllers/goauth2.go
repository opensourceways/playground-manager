/*
Description:Implementation of gitee authorization authentication login
Author: Zhang Jianjun
Date: 2021-10-12
*/
package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"playground_backend/handler"
	"playground_backend/models"
)

type GiteeCallBackControllers struct {
	beego.Controller
}

// @Title callback
// @Description callback
// @Param	body		body 	models.callback	true		"body for user content"
// @Success 200 {int} models.callback
// @Failure 403 body is empty
// @router / [post]
func (u *GiteeCallBackControllers) Post() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	//json.Unmarshal(u.Ctx.Input.RequestBody, &tutorData)
	logs.Info("gitee token request parameters: ", string(u.Ctx.Input.RequestBody))
	u.Ctx.ResponseWriter.WriteHeader(200)
	u.Ctx.WriteString("success")
	return
}

func (c *GiteeCallBackControllers) RetData(resp CodeResData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

type CodeResData struct {
	GiteeCode string `json:"giteeCode"`
	Mesg      string `json:"message"`
	Code      int    `json:"code"`
}

// @Title Get GiteeCallBack
// @Description get GiteeCallBack
// @Param	status	int	true (0,1,2)
// @Success 200 {object} GiteeCallBack
// @Failure 403 :status is err
// @router / [get]
func (u *GiteeCallBackControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	code := u.GetString("code", "")
	logs.Info("code: ", code)
	rui := handler.RespUserInfo{}
	handler.GetGiteeInfo(code, &rui)
	var crd CodeResData
	crd.GiteeCode = code
	crd.Code = 200
	crd.Mesg = "success"
	u.RetData(crd)
}

type GiteeCallBackUrlControllers struct {
	beego.Controller
}

type CallBackUrlData struct {
	CallBackUrl string `json:"callbackUrl"`
	ClientId    string `json:"clientId"`
}

type GetResData struct {
	CallBackUrl CallBackUrlData
	Mesg        string `json:"message"`
	Code        int    `json:"code"`
}

func (c *GiteeCallBackUrlControllers) RetData(resp GetResData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Get GiteeCallBackUrl
// @Description get GiteeCallBackUrl
// @Param	status	int	true (0,1,2)
// @Success 200 {object} GiteeCallBackUrl
// @Failure 403 :status is err
// @router / [get]
func (u *GiteeCallBackUrlControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	resp := GetResData{}
	var cbu CallBackUrlData
	callBackUrl := beego.AppConfig.String("gitee::oauth2_callback_url")
	clientId := beego.AppConfig.String("gitee::client_id")
	resp.Code = 200
	resp.Mesg = "success"
	cbu.CallBackUrl = callBackUrl
	cbu.ClientId = clientId
	resp.CallBackUrl = cbu
	u.RetData(resp)
	return
}

type GiteeOauth2Controllers struct {
	beego.Controller
}

type AuthCode struct {
	AuthCode string `json:"code"`
}

func (c *GiteeOauth2Controllers) RetData(resp OauthInfoData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

type OauthInfoData struct {
	UserInfo handler.RespUserInfo
	Mesg     string `json:"message"`
	Code     int    `json:"code"`
}

// @Title GiteeOauth2
// @Description GiteeOauth2
// @Param	body		body 	models.GiteeOauth2	true		"body for user content"
// @Success 200 {int} models.GiteeOauth2
// @Failure 403 body is empty
// @router / [post]
func (u *GiteeOauth2Controllers) Post() {
	var authCode AuthCode
	var oauthInfo OauthInfoData
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	logs.Info("gitee oauth2 request parameters: ", string(u.Ctx.Input.RequestBody))
	json.Unmarshal(u.Ctx.Input.RequestBody, &authCode)
	if len(authCode.AuthCode) < 1 {
		oauthInfo.Code = 404
		oauthInfo.Mesg = "Request parameter is empty"
		u.RetData(oauthInfo)
		return
	}
	rui := handler.RespUserInfo{}
	handler.GetGiteeInfo(authCode.AuthCode, &rui)
	logs.Info("rui: ", rui)
	oauthInfo.Code = 200
	oauthInfo.Mesg = "success"
	oauthInfo.UserInfo = rui
	u.RetData(oauthInfo)
	return
}

type GiteeUserInfoControllers struct {
	beego.Controller
}

type GetUserData struct {
	UserInfo handler.RespUserInfo
	Mesg     string `json:"message"`
	Code     int    `json:"code"`
}

func (c *GiteeUserInfoControllers) RetData(resp GetUserData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Get GiteeUserInfo
// @Description get GiteeUserInfo
// @Param	status	int	true (0,1,2)
// @Success 200 {object} GiteeUserInfo
// @Failure 403 :status is err
// @router / [get]
func (u *GiteeUserInfoControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	gud := GetUserData{}
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	token := u.GetString("token")
	userId, _ := u.GetInt64("userId", 0)
	if userId == 0 {
		gud.Mesg = "Request parameter error, no userId"
		gud.Code = 403
		u.RetData(gud)
		return
	}
	if token == "" {
		gud.Mesg = "Request parameter error"
		gud.Code = 403
		u.RetData(gud)
		return
	} else {
		gui := models.GiteeUserInfo{AccessToken: token, UserId: userId}
		rui := handler.RespUserInfo{}
		ok := handler.GetGiteeUserData(&gui, &rui)
		if !ok {
			gud.Mesg = "Request parameter error"
			gud.Code = 405
			u.RetData(gud)
			return
		}
		gud.UserInfo = rui
		u.RetData(gud)
	}
	return
}
