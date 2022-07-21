/*
Description:Implementation of gitee authorization authentication login
Author: Zhang Jianjun
Date: 2021-10-12
*/
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"playground_backend/common"
	"playground_backend/handler"
	"playground_backend/models"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/coreos/go-oidc/v3/oidc"
)

type Oauth2CallBackLinksControllers struct {
	beego.Controller
}

type CallBackUrlData struct {
	ClientId    string `json:"appId"`
	CallbackUrl string `json:"callbackUrl"`
	AppHost     string `json:"appHost"`
}

type GetResData struct {
	CallbackInfo CallBackUrlData `json:"callbackInfo"`
	Mesg         string          `json:"message"`
	Code         int             `json:"code"`
}

func (c *Oauth2CallBackLinksControllers) RetData(resp GetResData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Get Oauth2CallBackLinks
// @Description get Oauth2CallBackLinks
// @Param	status	int	true (0,1,2)
// @Success 200 {object} Oauth2CallBackLinks
// @Failure 403 :status is err
// @router / [get]
func (u *Oauth2CallBackLinksControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	resp := GetResData{}
	var cbu CallBackUrlData
	clientId := beego.AppConfig.String("gitee::client_id")
	callbackUrl := beego.AppConfig.String("gitee::callback_url")
	resp.Code = 200
	resp.Mesg = "success"
	cbu.ClientId = clientId
	cbu.CallbackUrl = callbackUrl
	cbu.AppHost = beego.AppConfig.String("gitee::redirect_url")
	resp.CallbackInfo = cbu
	u.RetData(resp)
	return
}

type Oauth2CallBackControllers struct {
	beego.Controller
}

// @Title callback
// @Description callback
// @Param	body		body 	models.callback	true		"body for user content"
// @Success 200 {int} models.callback
// @Failure 403 body is empty
// @router / [post]
func (u *Oauth2CallBackControllers) Post() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	//json.Unmarshal(u.Ctx.Input.RequestBody, &tutorData)
	logs.Info("gitee token request parameters: ", string(u.Ctx.Input.RequestBody))
	u.Ctx.ResponseWriter.WriteHeader(200)
	u.Ctx.WriteString("success")
	return
}

func (c *Oauth2CallBackControllers) RetData(resp OauthInfoData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

type CodeResData struct {
	AuthCode string `json:"authCode"`
	Mesg     string `json:"message"`
	Code     int    `json:"code"`
}

// @Title Get Oauth2CallBack
// @Description get Oauth2CallBack
// @Param	status	int	true (0,1,2)
// @Success 200 {object} Oauth2CallBack
// @Failure 403 :status is err
// @router / [get]
func (u *Oauth2CallBackControllers) Get() {
	var oauthInfo OauthInfoData
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	code := u.GetString("code", "")
	logs.Info("code: ", code)
	authCode := handler.AuthCode{AuthCode: code}
	if len(code) < 1 {
		oauthInfo.Code = 400
		oauthInfo.Mesg = "Authorization code is empty"
		u.RetData(oauthInfo)
		return
	}
	rui := handler.RespUserInfo{}
	var gui handler.GiteeUserInfo
	authErr := handler.SaveAuthUserInfo(authCode, &rui, &gui)
	if rui.UserId == 0 {
		logs.Error("authErr: ", authErr)
		oauthInfo.Code = 401
		oauthInfo.Mesg = "Wrong authorization code"
		u.RetData(oauthInfo)
		return
	}
	oauthInfo.Code = 200
	oauthInfo.Mesg = "success"
	oauthInfo.UserInfo = rui
	u.RetData(oauthInfo)
	// 2. Save key information to file
	userStr := ""
	userJson, jsonErr := json.Marshal(gui)
	if jsonErr == nil {
		userStr = string(userJson)
	}
	sd := handler.StatisticsData{UserId: rui.UserId, UserName: rui.NickName,
		OperationTime: common.GetCurTime(), EventType: "Authorization callback", State: "success",
		StateMessage: "success", Body: userStr}
	sdErr := handler.StatisticsLog(sd)
	if sdErr != nil {
		logs.Error("StatisticsLog, sdErr: ", sdErr)
	}
	return
}

type Oauth2AuthenticationControllers struct {
	beego.Controller
}

func (c *Oauth2AuthenticationControllers) RetData(resp OauthInfoData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

type OauthInfoData struct {
	UserInfo handler.RespUserInfo  `json:"userInfo"`
	Other    handler.GiteeUserInfo `json:"other"`
	Mesg     string                `json:"message"`
	Code     int                   `json:"code"`
	Domain   string                `json:"domain"`
}

// @Title Oauth2Authentication
// @Description Oauth2Authentication
// @Param	body		body 	models.Oauth2Authentication	true		"body for user content"
// @Success 200 {int} models.Oauth2Authentication
// @Failure 403 body is empty
// @router / [post]
func (u *Oauth2AuthenticationControllers) Post() {

	var authToken handler.AuthToken
	var oauthInfo OauthInfoData
	var rip handler.ReqIdPrams
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	logs.Info("Login authentication request parameters: ", string(u.Ctx.Input.RequestBody))
	jsErr := json.Unmarshal(u.Ctx.Input.RequestBody, &authToken)
	if jsErr != nil {
		oauthInfo.Code = 404
		oauthInfo.Mesg = "Parameter error"
		logs.Error("Bind Course parameters: ", authToken)
		u.RetData(oauthInfo)
	}
	if len(authToken.Id) < 1 {
		oauthInfo.Code = 400
		oauthInfo.Mesg = "Id is empty"
		u.RetData(oauthInfo)
		return
	}
	// 0. Test get authorization code
	//handler.GetAuthCode()
	// 1. Obtain token information based on authorization code
	rip.Id = authToken.Id
	rip.IdentityId = authToken.IdentityId
	rui := handler.RespUserInfo{}
	var gui handler.GiteeUserInfo
	authErr := handler.SaveAuthUserByToken(rip, &rui, &gui, authToken)
	if rui.UserId == 0 {
		logs.Error("authErr: ", authErr)
		oauthInfo.Code = 401
		oauthInfo.Mesg = "Wrong token"
		u.RetData(oauthInfo)
		return
	}
	oauthInfo.Code = 200
	oauthInfo.Mesg = "success"
	oauthInfo.UserInfo = rui
	oauthInfo.Other = gui
	oauthInfo.Domain = os.Getenv("DOMAIN")
	u.RetData(oauthInfo)
	// 2. Save key information to file
	userStr := ""
	userJson, jsonErr := json.Marshal(gui)
	if jsonErr == nil {
		userStr = string(userJson)
	}
	sd := handler.StatisticsData{UserId: rui.UserId, UserName: rui.NickName,
		OperationTime: common.GetCurTime(), EventType: "login", State: "success",
		StateMessage: "success", Body: userStr}
	sdErr := handler.StatisticsLog(sd)
	if sdErr != nil {
		logs.Error("StatisticsLog, sdErr: ", sdErr)
	}

}

type UserInfoControllers struct {
	beego.Controller
}

type GetUserData struct {
	UserInfo handler.RespUserInfo `json:"userInfo"`
	Mesg     string               `json:"message"`
	Code     int                  `json:"code"`

	Domain string `json:"domain"`
}

func (c *UserInfoControllers) RetData(resp GetUserData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Get UserInfo
// @Description get UserInfo
// @Param	status	int	true (0,1,2)
// @Success 200 {object} UserInfo
// @Failure 403 :status is err
// @router / [get]
func (u *UserInfoControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	gud := GetUserData{}
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	token := u.GetString("token")
	userId, _ := u.GetInt64("userId", 0)
	if userId == 0 {
		gud.Mesg = "User information error"
		gud.Code = 400
		u.RetData(gud)
		return
	}
	if token == "" {
		gud.Mesg = "Authority authentication failed"
		gud.Code = 403
		u.RetData(gud)
		return
	} else {
		gui := models.AuthUserInfo{AccessToken: token, UserId: userId}
		rui := handler.RespUserInfo{}
		ok := handler.GetGiteeUserData(&gui, &rui)
		if rui.UserId == 0 {
			gud.Mesg = "Requested user information does not exist"
			gud.Code = 404
			u.RetData(gud)
			return
		} else if !ok {
			gud.Mesg = "Authorization authentication timeout"
			gud.Code = 403
			u.RetData(gud)
			return
		}
		gud.Mesg = "success"
		gud.Code = 200
		gud.UserInfo = rui
		gud.Domain = os.Getenv("DOMAIN")
		u.RetData(gud)
	}
	return
}

func (u *UserInfoControllers) CheckLogin() {
	state, _ := handler.RandString(16)
	nonce, _ := handler.RandString(16)
	handler.SetCallbackCookie(u.Ctx.ResponseWriter, u.Ctx.Request, "state", state)
	handler.SetCallbackCookie(u.Ctx.ResponseWriter, u.Ctx.Request, "nonce", nonce)
	u.Redirect(handler.OIDCConfig.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func (u *UserInfoControllers) AuthingCallback() {
	// state := u.Ctx.GetCookie("state")
	// if len(state) == 0 {
	// 	u.Data["json"] = "state is empty"
	// 	u.ServeJSON()
	// 	return
	// }
	// if u.GetString("state") != state {
	// 	u.Data["json"] = "state is bad"
	// 	u.ServeJSON()
	// 	return
	// }
	result, err := handler.GetTokenFromAuthing(u.GetString("code"))

	if err != nil {
		u.Data["json"] = fmt.Sprintf("code is bad, error:%v", err)
		u.ServeJSON()
		return
	}
	var authToken handler.AuthToken
	authToken.Id = result.User.Sub
	var rip handler.ReqIdPrams
	rip.Id = authToken.Id
	rui := handler.RespUserInfo{}
	var gui handler.GiteeUserInfo
	gui.AccessToken = result.Token
	handler.SaveAuthUserByToken(rip, &rui, &gui, authToken)
	u.Data["json"] = result
	u.ServeJSON()
}

// @Summary GetCurrentUser
// @Description 获取用户资料, 在请求的request的header中必须带有accessToken
// @Tags  User
// @Accept json
// @Produce json
// @Router /playground/user/information [get]
func (u *UserInfoControllers) GetCurrentUser() {
	userid := u.Data["me"].(string)
	userpool_id := beego.AppConfig.String("gitee::userpool_id")
	userpool_secret := beego.AppConfig.String("gitee::userpool_secret")
	gui := new(handler.GiteeUserInfo)
	handler.GetAuthUserBySub(userpool_id, userpool_secret, userid, gui)

	useridInt, _ := strconv.Atoi(userid)
	// aui := models.AuthUserInfo{AccessToken: u.GetString("token"), UserId: int64(useridInt)}
	// rui := handler.RespUserInfo{}
	// handler.GetGiteeUserData(&aui, &rui)

	// gui.Identity = nil

	rui := handler.RespUserInfo{}
	rui.AvatarUrl = gui.Picture
	rui.Email = gui.Email
	rui.NickName = gui.Nickname
	rui.UserId = int64(useridInt)
	rui.UserToken = u.GetString("token")
	gud := GetUserData{}
	gud.Mesg = "success"
	gud.Code = 200
	gud.UserInfo = rui
	gud.Domain = os.Getenv("DOMAIN")
	u.RetData(gud)
	// u.Data["json"] = gui
	// u.ServeJSON()
}
