package controllers

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"playground_backend/common"
	"playground_backend/handler"
	"playground_backend/models"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type CrdResourceControllers struct {
	beego.Controller
}

func (c *CrdResourceControllers) RetData(resp handler.ResData) {
	logs.Info("Create Resource Response: ", resp)
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title CreateCrdResource
// @Description CreateCrdResource
// @Param	body		body 	models.CreateCrdResource	true		"body for user content"
// @Success 200 {int} models.CreateCrdResource
// @Failure 403 body is empty
// @router / [post]
func (u *CrdResourceControllers) Post() {
	var rp handler.RequestParameter
	var resData handler.ResData
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Error("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	logs.Info("created crd parameters: ", string(u.Ctx.Input.RequestBody))
	jsErr := json.Unmarshal(u.Ctx.Input.RequestBody, &rp)
	if jsErr != nil {
		resData.Code = 404
		resData.Mesg = "Parameter error"
		logs.Error("Bind Course parameters: ", rp)
		u.RetData(resData)
		return
	}
	if (len(rp.TemplatePath) < 1 && len(rp.Backend) < 1) || len(rp.CourseId) < 1 {
		resData.Code = 400
		resData.Mesg = "Please check whether the request parameters are correct"
		logs.Error("created crd parameters: ", rp)
		u.RetData(resData)
		// crd := models.Courses{CourseId: rp.CourseId}
		// ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		// handler.WriteCourseData(rp.UserId, rp.CourseId, rp.ChapterId, "Application Resources",
		// 	"", "failed", "Please check whether the request parameters are correct",
		// 	1, 1, &crd, &ccp)
		return
	}
	// if len(rp.Token) < 1 {
	// 	resData.Code = 401
	// 	resData.Mesg = "Unauthorized authentication information"
	// 	u.RetData(resData)
	// 	// crd := models.Courses{CourseId: rp.CourseId}
	// 	// ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
	// 	// handler.WriteCourseData(rp.UserId, rp.CourseId, rp.ChapterId,
	// 	// 	"Application Resources", "", "failed",
	// 	// 	"Unauthorized authentication information",
	// 	// 	1, 1, &crd, &ccp)
	// 	return
	// } else {
	// 	gui := models.AuthUserInfo{AccessToken: rp.Token, UserId: rp.UserId}
	// 	ok := handler.CheckToken(&gui)
	// 	if !ok {
	// 		resData.Mesg = "Authority authentication failed:" + rp.Token + "----userid:" + strconv.Itoa(int(rp.UserId))
	// 		resData.Code = 403
	// 		u.RetData(resData)

	// 		// crd := models.Courses{CourseId: rp.CourseId}
	// 		// ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
	// 		// handler.WriteCourseData(rp.UserId, rp.CourseId, rp.ChapterId, "Application Resources",
	// 		// 	"", "filed", "Authority authentication failed",
	// 		// 	1, 1, &crd, &ccp)

	// 		return
	// 	}
	// }

	if rp.ForceDelete == 0 {
		rp.ForceDelete = 1
	}
	userid, _ := strconv.Atoi(rp.UserId)
	// Query resource node information
	rcp := models.ResourceConfigPath{ResourcePath: rp.TemplatePath, EulerBranch: rp.Backend}
	var rri = new(handler.ResResourceInfo)
	rr := handler.ReqResource{EnvResource: rp.TemplatePath, UserId: int64(userid),
		ContactEmail: rp.ContactEmail, ForceDelete: rp.ForceDelete,
		ResourceId: rcp.ResourceId, CourseId: rp.CourseId, ChapterId: rp.ChapterId}
	rri.CourseId = rp.CourseId
	rri.ChapterId = rp.ChapterId
	cs := models.Courses{CourseId: rp.CourseId}
	queryErr := models.QueryCourse(&cs, "CourseId")
	if queryErr != nil {
		logs.Error("createResource queryErr: ", queryErr)
		resData.Mesg = "Retry later while course info is syncing"
		resData.Code = 404
		u.RetData(resData)
		// crd := models.Courses{CourseId: rp.CourseId}
		// ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		// handler.WriteCourseData(rp.UserId, rp.CourseId, rp.ChapterId, "Application Resources",
		// 	"", "filed", "Retry later while course info is syncing",
		// 	1, 1, &crd, &ccp)
		return
	}
	rcpErr := rr.SaveCourseAndResRel(&rcp, cs.Name)
	if rcpErr != nil {
		resData.Code = 403
		resData.Mesg = "The corresponding instance resource is not currently configured"
		logs.Error("created crd parameters: ", rp)
		u.RetData(resData)
		// crd := models.Courses{CourseId: rp.CourseId}
		// ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		// handler.WriteCourseData(rp.UserId, rp.CourseId, rp.ChapterId,
		// 	"Application Resources", "", "failed",
		// 	"The corresponding instance resource is not currently configured",
		// 	1, 1, &crd, &ccp)
		return
	}
	rp.TemplatePath = rcp.ResourcePath
	rp.ResourceId = rcp.ResourceId
	rp.Backend = rcp.EulerBranch
	rr.EnvResource = rcp.ResourcePath
	rr.ResourceId = rcp.ResourceId
	err := handler.CreateEnvResource(rr, rri)
	if rri.UserId > 0 {
		if rri.Status == 0 {
			resData.Code = 202
		} else {
			resData.Code = 200
		}
		crd := models.Courses{CourseId: rp.CourseId}
		ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		handler.WriteCourseData(int64(userid), rp.CourseId, rp.ChapterId, "Application Resources", rri.ResName,
			"success", "User learning courses apply for instance resources successfully",
			1, 1, &crd, &ccp)
		userResId := handler.CreateUserResourceEnv(rr)
		rri.UserResId = userResId
		resData.ResInfo = *rri
		resData.Mesg = "success"
	} else {
		resData.ResInfo = err.Error()
		resData.Code = 501
		resData.Mesg = "Failed to create resource, need to request resource again"
		crd := models.Courses{CourseId: rp.CourseId}
		ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		handler.WriteCourseData(int64(userid), rp.CourseId, rp.ChapterId, "Application Resources", rri.ResName,
			"failed", "Failed to create resource, need to request resource again",
			1, 1, &crd, &ccp)
	}
	u.RetData(resData)

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
	var resData handler.ResData
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	token := u.GetString("token")
	userResId, _ := u.GetInt64("userResId", 0)
	if userResId == 0 {
		resData.Mesg = "Please check whether to upload user resource id information"
		resData.Code = 404
		u.RetData(resData)
		return
	}
	ure := models.UserResourceEnv{Id: userResId}
	handler.QueryUserResourceEnv(&ure)
	if ure.Id == 0 {
		resData.Mesg = "User resource id information is wrong"
		resData.Code = 405
		u.RetData(resData)
		return
	}
	if token == "" {
		resData.Mesg = "Unauthorized authentication information"
		resData.Code = 403
		u.RetData(resData)
		return
	} else {
		var rri = new(handler.ResResourceInfo)
		rr := handler.ReqResource{EnvResource: ure.TemplatePath, UserId: ure.UserId,
			ResourceId: ure.ResourceId, CourseId: ure.CourseId,
			ChapterId: ure.ChapterId, ContactEmail: ure.ContactEmail}
		rri.CourseId = ure.CourseId
		rri.ChapterId = ure.ChapterId
		handler.GetEnvResource(rr, rri)
		rri.UserResId = userResId
		resData.ResInfo = *rri
		resData.Code = 200
		resData.Mesg = "success"
		u.RetData(resData)
		crd := models.Courses{CourseId: ure.CourseId}
		ccp := models.CoursesChapter{CourseId: ure.CourseId, ChapterId: ure.ChapterId}
		handler.WriteCourseData(ure.UserId, ure.CourseId, ure.ChapterId, "Query application resources", rri.ResName,
			"success", "Query application resource success",
			1, 1, &crd, &ccp)
	}
	return
}

type CheckSubdomain struct {
	Token     string `json:"token"`
	Subdomain string `json:"subdomain"`
}

// @Title Get CheckSubdomain
// @Description 验证subdomain 和token
// @Param	status	int	true (0,1,2)
// @Success 200 {object} CheckPgweb
// @Failure 403 :status is err
// @router /playground/users/checkSubdomain [post]
func (u *CrdResourceControllers) CheckSubdomain() {

	var checkSubdomain CheckSubdomain
	err := json.Unmarshal(u.Ctx.Input.RequestBody, &checkSubdomain)
	if err != nil {
		u.Data["json"] = map[string]interface{}{
			"detail": "body参数异常 ",
			"error":  err.Error(),
			"code":   500,
		}
		u.ServeJSON()
		return
	}
	if len(checkSubdomain.Subdomain) == 0 || len(checkSubdomain.Token) == 0 {
		u.Data["json"] = map[string]interface{}{
			"checkSubdomain": checkSubdomain,
			"error":          "body参数异常",
			"code":           500,
		}
		u.ServeJSON()
		return

	}
	urlItem, err := url.Parse(checkSubdomain.Subdomain)
	if err != nil {
		u.Data["json"] = map[string]interface{}{
			"detail": "Subdomain参数异常 ",
			"error":  err.Error(),
			"code":   500,
		}
		u.ServeJSON()
		return
	}

	hostSplit := strings.Split(urlItem.Host, ".")
	if len(hostSplit) < 2 {
		u.Data["json"] = map[string]interface{}{
			"detail": "Subdomain参数异常 ",
			"host":   urlItem.Host,
			"code":   500,
		}
		u.ServeJSON()
		return
	}

	resourceInfo := models.ResourceInfo{Subdomain: hostSplit[0]}
	err = models.QueryResourceInfo(&resourceInfo, "Subdomain")
	if err != nil {
		u.Data["json"] = map[string]interface{}{
			"detail":    "查询Subdomain时候出错 ",
			"Subdomain": hostSplit[0],
			"error":     err.Error(),
			"code":      500,
		}
		u.ServeJSON()
		return
	}

	authUseraInfo := models.AuthUserInfo{AccessToken: checkSubdomain.Token}
	err = models.QueryAuthUserInfo(&authUseraInfo, "AccessToken")
	if err != nil {
		u.Data["json"] = map[string]interface{}{
			"detail":    "查询AccessToken时候出错 ",
			"Subdomain": checkSubdomain.Token,
			"error":     err.Error(),
			"code":      500,
		}
		u.ServeJSON()
		return
	}
	if int(authUseraInfo.UserId) != int(resourceInfo.UserId) {
		u.Data["json"] = map[string]interface{}{
			"detail": authUseraInfo.UserId,
			"body":   " 不等",
			"error":  resourceInfo.UserId,
			"code":   500,
		}
		u.ServeJSON()
		return
	}

	u.Data["json"] = map[string]interface{}{
		"code": 200,
		"data": resourceInfo,
	}
	u.ServeJSON()
}

// @Title Get MakeKubeconfig
// @Description 加密和编码kubeconfig内容
// @Param	kubeconfig		FormData 	file	true		"kubeconfig"
// @Success 200 {object} CheckPgweb
// @Failure 403 :status is err
// @router /playground/crd/makeKubeconfig [post]
func (u *CrdResourceControllers) MakeKubeconfig() {
	aesStr := common.AesString(u.Ctx.Input.RequestBody)
	encodeRes := base64.StdEncoding.EncodeToString([]byte(aesStr))
	u.Data["json"] = map[string]interface{}{
		"code": 200,
		"data": encodeRes,
	}
	u.ServeJSON()
}
