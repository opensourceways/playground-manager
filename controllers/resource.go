package controllers

import (
	"encoding/json"
	"playground_backend/handler"
	"playground_backend/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type CrdResourceControllers struct {
	beego.Controller
}

type RequestParameter struct {
	ResourceId   string `json:"resourceId"`
	CourseId     string `json:"courseId"`
	ChapterId    string `json:"chapterId"`
	Backend      string `json:"backend"`
	TemplatePath string `json:"templatePath"`
	UserId       int64  `json:"userId"`
	ContactEmail string `json:"contactEmail"`
	Token        string `json:"token"`
	ForceDelete  int    `json:"forceDelete"`
}

func (c *CrdResourceControllers) RetData(resp ResData) {
	logs.Info("Create Resource Response: ", resp)
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
	logs.Info("created crd parameters: ", string(u.Ctx.Input.RequestBody))
	jsErr := json.Unmarshal(u.Ctx.Input.RequestBody, &rp)
	if jsErr != nil {
		resData.Code = 404
		resData.Mesg = "Parameter error"
		logs.Error("Bind Course parameters: ", rp)
		u.RetData(resData)
	}
	if (len(rp.TemplatePath) < 1 && len(rp.Backend) < 1) ||
		rp.UserId < 1 || len(rp.CourseId) < 1 {
		resData.Code = 400
		resData.Mesg = "Please check whether the request parameters are correct"
		logs.Error("created crd parameters: ", rp)
		u.RetData(resData)
		crd := models.Courses{CourseId: rp.CourseId}
		ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		handler.WriteCourseData(rp.UserId, rp.ResourceId, rp.CourseId, rp.ChapterId, "Application Resources",
			"", "failed", "Please check whether the request parameters are correct",
			1, 1, &crd, &ccp)
		return
	}
	if len(rp.Token) < 1 {
		resData.Code = 401
		resData.Mesg = "Unauthorized authentication information"
		u.RetData(resData)
		crd := models.Courses{CourseId: rp.CourseId}
		ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		handler.WriteCourseData(rp.UserId, "0", rp.CourseId, rp.ChapterId,
			"Application Resources", "", "failed",
			"Unauthorized authentication information",
			1, 1, &crd, &ccp)
		return
	} else {
		gui := models.AuthUserInfo{AccessToken: rp.Token, UserId: rp.UserId}
		ok := handler.CheckToken(&gui)
		if !ok {
			logs.Error("CheckToken Error: ", gui)
			resData.Mesg = "Authority authentication failed"
			resData.Code = 403
			u.RetData(resData)
			crd := models.Courses{CourseId: rp.CourseId}
			ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
			handler.WriteCourseData(rp.UserId, rp.ResourceId, rp.CourseId, rp.ChapterId, "Application Resources",
				"", "filed", "Authority authentication failed",
				1, 1, &crd, &ccp)
			return
		}
	}
	if rp.ForceDelete == 0 {
		rp.ForceDelete = 1
	}
	// Query resource node information
	rcp := models.ResourceConfigPath{ResourcePath: rp.TemplatePath, EulerBranch: rp.Backend}
	var rri = new(handler.ResResourceInfo)
	rr := handler.ReqResource{EnvResource: rp.TemplatePath, UserId: rp.UserId,
		ContactEmail: rp.ContactEmail, ForceDelete: rp.ForceDelete,
		ResourceId: rcp.ResourceId, CourseId: rp.CourseId, ChapterId: rp.ChapterId}
	rri.CourseId = rp.CourseId
	rri.ChapterId = rp.ChapterId
	cs := models.Courses{CourseId: rp.CourseId}
	queryErr := models.QueryCourse(&cs, "CourseId")
	if queryErr != nil {
		logs.Error("QueryCourse Error: ", queryErr)
		resData.Mesg = "Retry later while course info is syncing"
		resData.Code = 404
		u.RetData(resData)
		crd := models.Courses{CourseId: rp.CourseId}
		ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		handler.WriteCourseData(rp.UserId, rp.ResourceId, rp.CourseId, rp.ChapterId, "Application Resources",
			"", "filed", "Retry later while course info is syncing",
			1, 1, &crd, &ccp)
		return
	}
	rcpErr := rr.SaveCourseAndResRel(&rcp, cs.Name)
	if rcpErr != nil {
		resData.Code = 403
		resData.Mesg = "The corresponding instance resource is not currently configured"
		logs.Error(rcp, "SaveCourseAndResRel crd parameters: ", rp)
		u.RetData(resData)
		crd := models.Courses{CourseId: rp.CourseId}
		ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		handler.WriteCourseData(rp.UserId, rp.ResourceId, rp.CourseId, rp.ChapterId,
			"Application Resources", "", "failed",
			"The corresponding instance resource is not currently configured",
			1, 1, &crd, &ccp)
		return
	}
	rp.TemplatePath = rcp.ResourcePath
	rp.ResourceId = rcp.ResourceId
	rp.Backend = rcp.EulerBranch
	rr.EnvResource = rcp.ResourcePath
	rr.ResourceId = rcp.ResourceId
	handler.CreateEnvResource(rr, rri)
	if rri.UserId > 0 {
		if rri.Status == 0 {
			resData.Code = 202
		} else {
			resData.Code = 200
		}
		crd := models.Courses{CourseId: rp.CourseId}
		ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		userResId := handler.CreateUserResourceEnv(rr)
		handler.WriteCourseData(rp.UserId, rp.ResourceId, rp.CourseId, rp.ChapterId, "Application Resources", rri.ResName,
			"success", "User learning courses apply for instance resources successfully",
			1, 1, &crd, &ccp)
		rri.UserResId = userResId
		resData.ResInfo = *rri
		resData.Mesg = "success"
	} else {
		resData.ResInfo = *rri
		resData.Code = 501
		resData.Mesg = "Failed to create resource, need to request resource again"
		crd := models.Courses{CourseId: rp.CourseId}
		ccp := models.CoursesChapter{CourseId: rp.CourseId, ChapterId: rp.ChapterId}
		handler.WriteCourseData(rp.UserId, rp.ResourceId, rp.CourseId, rp.ChapterId, "Application Resources", rri.ResName,
			"failed", "Failed to create resource, need to request resource again",
			1, 1, &crd, &ccp)
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
		handler.WriteCourseData(ure.UserId, ure.ResourceId, ure.CourseId, ure.ChapterId, "Query application resources", rri.ResName,
			"success", "Query application resource success",
			1, 1, &crd, &ccp)
	}
	return
}
