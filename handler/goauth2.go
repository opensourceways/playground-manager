package handler

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"playground_backend/common"
	"playground_backend/http"
	"playground_backend/models"
	"time"
)

type GiteeTokenInfo struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

// Request code cloud access token
func GiteePostOauthToken(code, clientId, redirectUri, clientSecret string, giteeToken *GiteeTokenInfo) {
	url := fmt.Sprintf(`https://gitee.com/oauth/token?grant_type=authorization_code&code=%s&client_id=%s&redirect_uri=%s`,
		code, clientId, redirectUri)
	param := fmt.Sprintf(`{"client_secret": "%s"}`, clientSecret)
	res, err := http.HTTPPost(url, param)
	if err != nil {
		logs.Error(err)
		return
	}
	logs.Info("Access token result of request code cloud:", res)
	GiteeConstructor(res, giteeToken)
}

func GiteePostFreshToken(refreshToken string, giteeToken *GiteeTokenInfo) {
	url := fmt.Sprintf(`https://gitee.com/oauth/token?grant_type=refresh_token`)
	param := fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken)
	res, err := http.HTTPPost(url, param)
	if err != nil {
		logs.Error(err)
		return
	}
	logs.Info("Access token result of request code cloud:", res)
	GiteeConstructor(res, giteeToken)
}

func GiteeConstructor(res map[string]interface{}, giteeToken *GiteeTokenInfo) {
	if _, ok := res["access_token"]; ok {
		giteeToken.AccessToken = res["access_token"].(string)
		giteeToken.TokenType = res["token_type"].(string)
		giteeToken.RefreshToken = res["refresh_token"].(string)
		giteeToken.Scope = res["scope"].(string)
		giteeToken.ExpiresIn = int64(res["expires_in"].(float64))
		giteeToken.CreatedAt = int64(res["created_at"].(float64))
	}
}

type GiteeUserInfo struct {
	UserId    int64  `json:"id"`
	UserLogin string `json:"login"`
	UserName  string `json:"name"`
	Url       string `json:"url"`
	AvatarUrl string `json:"avatarUrl"`
}

// Obtain user information based on token
func GetGiteeUserInfoByToken(giteeToken string, giteeUser *GiteeUserInfo) {
	url := fmt.Sprintf(`https://gitee.com/api/v5/user?access_token=%s`, giteeToken)
	body, resErr := http.HTTPGitGet(url)
	if resErr != nil {
		logs.Error("resErr: ", resErr)
		return
	}
	GiteeUserConstructor(body, giteeUser)
}

func GiteeUserConstructor(res map[string]interface{}, giteeUser *GiteeUserInfo) {
	giteeUser.UserLogin = res["login"].(string)
	giteeUser.UserName = res["name"].(string)
	giteeUser.Url = res["url"].(string)
	giteeUser.AvatarUrl = res["avatar_url"].(string)
	giteeUser.UserId = int64(res["id"].(float64))
}

func ProcOauthData(giteeToken GiteeTokenInfo, giteeUser GiteeUserInfo, token, authCode string) int64 {
	userId := int64(0)
	gui := models.GiteeUserInfo{GitId: giteeUser.UserId}
	queryErr := models.QueryGiteeUserInfo(&gui, "GitId")
	if gui.UserId > 0 {
		userId = gui.UserId
		CreateGiteeUserInfo(&gui, giteeUser, 2, token)
		models.UpdateGiteeUserInfo(&gui, "UserLogin",
			"UserName", "UserUrl", "AccessToken", "ExpirationTime", "UpdateTime", "AvatarUrl")
	} else {
		logs.Info("queryErr: ", queryErr)
		CreateGiteeUserInfo(&gui, giteeUser, 1, token)
		id, inErr := models.InsertGiteeUserInfo(&gui)
		if id > 0 {
			userId = id
		} else {
			logs.Error("inErr: ", inErr)
		}
	}
	if userId > 0 {
		gti := models.GiteeTokenInfo{UserId: userId}
		getErr := models.QueryGiteeTokenInfo(&gti, "UserId")
		if gti.Id > 0 {
			CreateGiteeTokenInfo(&gti, giteeToken, 2, authCode)
			models.UpdateGiteeTokenInfo(&gti, "UpdateTime", "AccessToken",
				"ExpiresIn", "Scope", "CreatedAt", "RefreshToken", "TokenType", "authCode")
		} else {
			logs.Info("getErr: ", getErr)
			gti.UserId = userId
			CreateGiteeTokenInfo(&gti, giteeToken, 1, authCode)
			models.InsertGiteeTokenInfo(&gti)
		}
	}
	return userId
}

func CreateGiteeTokenInfo(gti *models.GiteeTokenInfo, giteeToken GiteeTokenInfo, flag int, authCode string) {
	if flag == 1 {
		gti.CreateTime = common.GetCurTime()
	} else {
		gti.UpdateTime = common.GetCurTime()
	}
	gti.AccessToken = giteeToken.AccessToken
	gti.ExpiresIn = giteeToken.ExpiresIn
	gti.Scope = giteeToken.Scope
	gti.CreatedAt = giteeToken.CreatedAt
	gti.RefreshToken = giteeToken.RefreshToken
	gti.TokenType = giteeToken.TokenType
	if len(authCode) > 0 {
		gti.AuthCode = authCode
	}
}

func CreateGiteeUserInfo(gui *models.GiteeUserInfo, giteeUser GiteeUserInfo, flag int, token string) {
	expirTime, _ := beego.AppConfig.Int("gitee::token_expir_time")
	newTime := time.Now().AddDate(0, 0, expirTime).Format(common.DATE_FORMAT)
	gui.UserLogin = giteeUser.UserLogin
	gui.UserName = giteeUser.UserName
	gui.UserUrl = giteeUser.Url
	gui.AccessToken = token
	gui.AvatarUrl = giteeUser.AvatarUrl
	gui.ExpirationTime = newTime
	if flag == 1 {
		gui.CreateTime = common.GetCurTime()
		gui.GitId = giteeUser.UserId
	} else {
		gui.UpdateTime = common.GetCurTime()
	}
}

type RespUserInfo struct {
	UserId     int64  `json:"userId"`
	UserLogin  string `json:"login"`
	UserName   string `json:"name"`
	Url        string `json:"url"`
	AvatarUrl  string `json:"avatarUrl"`
	GiteeToken string `json:"giteeToken"`
	UserToken  string `json:"userToken"`
}

func GetGiteeInfo(authCode string, rui *RespUserInfo) {
	redirectUri := beego.AppConfig.String("gitee::oauth2_callback_url")
	clientSecret := beego.AppConfig.String("gitee::client_secret")
	clientId := beego.AppConfig.String("gitee::client_id")
	var giteeToken GiteeTokenInfo
	GiteePostOauthToken(authCode, clientId, redirectUri, clientSecret, &giteeToken)
	if len(giteeToken.AccessToken) > 1 {
		var giteeUser GiteeUserInfo
		GetGiteeUserInfoByToken(giteeToken.AccessToken, &giteeUser)
		if giteeUser.UserId > 0 {
			token, terr := common.GenToken(giteeUser.UserName, giteeUser.UserLogin)
			if terr == nil {
				userId := ProcOauthData(giteeToken, giteeUser, token, authCode)
				rui.UserId = userId
				rui.UserToken = token
				CreateRespUserInfo(rui, giteeToken, giteeUser)
			}
		}
	} else {
		gti := models.GiteeTokenInfo{AuthCode: authCode}
		getErr := models.QueryGiteeTokenInfo(&gti, "AuthCode")
		if getErr == nil {
			gui := models.GiteeUserInfo{UserId: gti.UserId}
			queryErr := models.QueryGiteeUserInfo(&gui, "UserId")
			if queryErr == nil {
				rui.UserId = gui.UserId
				rui.UserToken = gui.AccessToken
				rui.Url = gui.UserUrl
				rui.UserName = gui.UserName
				rui.UserLogin = gui.UserLogin
				rui.GiteeToken = gti.AccessToken
				rui.AvatarUrl = gui.AvatarUrl
			}
		}
	}
}

func CreateRespUserInfo(rui *RespUserInfo, giteeToken GiteeTokenInfo, giteeUser GiteeUserInfo) {
	rui.GiteeToken = giteeToken.AccessToken
	rui.Url = giteeUser.Url
	rui.UserName = giteeUser.UserName
	rui.UserLogin = giteeUser.UserLogin
	rui.AvatarUrl = giteeUser.AvatarUrl
}

//CheckToken Check whether the token is legal
func GetGiteeUserData(gui *models.GiteeUserInfo, rui *RespUserInfo) bool {
	queryErr := models.QueryGiteeUserInfo(gui, "AccessToken", "UserId")
	if gui.UserId > 0 {
		now := time.Now().Format(common.DATE_FORMAT)
		logs.Info("token: now: ", now, ",expir: ", gui.ExpirationTime)
		if now > gui.ExpirationTime {
			return false
		}
	} else {
		logs.Error("queryErr: ", queryErr)
		return false
	}
	GetUserInfoByReshToken(gui.UserId, gui.AccessToken, rui)
	return true
}

func CheckToken(gui *models.GiteeUserInfo) bool {
	queryErr := models.QueryGiteeUserInfo(gui, "AccessToken", "UserId")
	if gui.UserId > 0 {
		now := time.Now().Format(common.DATE_FORMAT)
		logs.Info("token: now: ", now, ",expir: ", gui.ExpirationTime)
		if now > gui.ExpirationTime {
			return false
		}
	} else {
		logs.Error("queryErr: ", queryErr)
		return false
	}
	return true
}

func GetUserInfoByReshToken(userId int64, token string, rui *RespUserInfo) {
	gti := models.GiteeTokenInfo{UserId: userId}
	models.QueryGiteeTokenInfo(&gti, "UserId")
	if gti.Id > 0 {
		giteeToken := GiteeTokenInfo{}
		GiteePostFreshToken(gti.RefreshToken, &giteeToken)
		if len(giteeToken.AccessToken) > 1 {
			var giteeUser GiteeUserInfo
			GetGiteeUserInfoByToken(giteeToken.AccessToken, &giteeUser)
			if giteeUser.UserId > 0 {
				userId := ProcOauthData(giteeToken, giteeUser, token, "")
				rui.UserId = userId
				rui.UserToken = token
				CreateRespUserInfo(rui, giteeToken, giteeUser)
			}
		}
	}
}
