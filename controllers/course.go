package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"playground_backend/handler"
	"playground_backend/models"
)

type CourseChapterControllers struct {
	beego.Controller
}

func (c *CourseChapterControllers) RetData(resp RescourseData) {
	logs.Info("Bind Course Response: ", resp)
	c.Data["json"] = resp
	c.ServeJSON()
}

type RescourseData struct {
	ResInfo handler.RspCourse `json:"courseInfo"`
	Mesg    string            `json:"message"`
	Code    int               `json:"code"`
}

// @Title CourseChapter
// @Description CourseChapter
// @Param	body		body 	models.CourseChapter	true		"body for user content"
// @Success 200 {int} models.CourseChapter
// @Failure 403 body is empty
// @router / [post]
func (u *CourseChapterControllers) Post() {
	var crp handler.CourseReqParameter
	var resData RescourseData
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	logs.Info("Bind Course parameters: ", string(u.Ctx.Input.RequestBody))
	err := json.Unmarshal(u.Ctx.Input.RequestBody, &crp)
	if err != nil {
		logs.Error("json.Unmarshal, err: ", err)
		resData.Code = 404
		resData.Mesg = "Parameter error"
		logs.Error("Bind Course parameters: ", crp)
		u.RetData(resData)
		return
	}
	resData.ResInfo.CId = 0
	resData.ResInfo.State = "error"
	if len(crp.CourseId) < 1 || crp.UserId < 1 {
		resData.Code = 400
		resData.Mesg = "Please check whether the request parameters are correct"
		logs.Error("Bind Course parameters: ", crp)
		u.RetData(resData)
		crd := models.Courses{CourseId: crp.CourseId}
		ccp := models.CoursesChapter{CourseId: crp.CourseId, ChapterId: ""}
		handler.WriteCourseData(crp.UserId, crp.CourseId, "",
			"User bound course", "", "failed",
			"Please check whether the request parameters are correct",
			crp.Status, crp.Status, &crd, &ccp)
		return
	}
	if len(crp.Token) < 1 {
		resData.Code = 401
		resData.Mesg = "Unauthorized authentication information"
		u.RetData(resData)
		crd := models.Courses{CourseId: crp.CourseId}
		ccp := models.CoursesChapter{CourseId: crp.CourseId, ChapterId: ""}
		handler.WriteCourseData(crp.UserId, crp.CourseId, "",
			"User bound course", "", "failed",
			"Unauthorized authentication information",
			crp.Status, crp.Status, &crd, &ccp)
		return
	} else {
		gui := models.AuthUserInfo{AccessToken: crp.Token, UserId: crp.UserId}
		ok := handler.CheckToken(&gui)
		if !ok {
			resData.Mesg = "Authority authentication failed"
			resData.Code = 403
			u.RetData(resData)
			crd := models.Courses{CourseId: crp.CourseId}
			ccp := models.CoursesChapter{CourseId: crp.CourseId, ChapterId: ""}
			handler.WriteCourseData(crp.UserId, crp.CourseId, "", "User bound course",
				"", "failed", "Authority authentication failed",
				crp.Status, crp.Status, &crd, &ccp)
			return
		}
	}
	// User bound course
	ucId := handler.UserBoundBourse(crp)
	if ucId > 0 {
		if len(crp.ChapterInfo) > 0 {
			for _, chInfo := range crp.ChapterInfo {
				handler.UserBoundBourseChapter(chInfo, ucId, crp.UserId, crp.CourseId, crp.Status)
			}
		}
		resData.ResInfo.CId = ucId
		resData.ResInfo.State = "success"
		resData.Code = 200
		resData.Mesg = "User binding course successfully"
	} else {
		resData.Mesg = "Service internal processing failed"
		resData.Code = 500
		u.RetData(resData)
		return
	}
	u.RetData(resData)
	return
}

type RescourseDetailData struct {
	CourseInfo []handler.RspCourseData `json:"courseInfo"`
	Mesg       string                  `json:"message"`
	Code       int                     `json:"code"`
	TotalCount int64                   `json:"totalCount"`
}

func (c *CourseChapterControllers) RetDetailData(resp RescourseDetailData) {
	logs.Info("For course information taken: ", resp)
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Get CourseChapter
// @Description get CourseChapter
// @Param	status	int	true (0,1,2)
// @Success 200 {object} CourseChapter
// @Failure 403 :status is err
// @router / [get]
func (u *CourseChapterControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	var resData RescourseDetailData
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	token := u.GetString("token")
	userId, _ := u.GetInt64("userId", 0)
	currentPage, _ := u.GetInt("currentPage", 1)
	pageSize, _ := u.GetInt("pageSize", 100)
	if userId == 0 {
		resData.Mesg = "User information error"
		resData.Code = 404
		u.RetDetailData(resData)
		return
	}
	if token == "" {
		resData.Mesg = "Unauthorized authentication information"
		resData.Code = 403
		u.RetDetailData(resData)
		crd := models.Courses{}
		ccp := models.CoursesChapter{}
		handler.WriteCourseData(userId, "", "", "Query user's courses",
			"", "failed", "Unauthorized authentication information",
			1, 1, &crd, &ccp)
		return
	} else {
		gui := models.AuthUserInfo{AccessToken: token, UserId: userId}
		ok := handler.CheckToken(&gui)
		if !ok {
			resData.Mesg = "Authority authentication failed"
			resData.Code = 403
			u.RetDetailData(resData)
			crd := models.Courses{}
			ccp := models.CoursesChapter{}
			handler.WriteCourseData(userId, "", "", "Query user's courses",
				"", "failed", "Authority authentication failed",
				1, 1, &crd, &ccp)
			return
		}
	}
	// Get the total number of courses the user has taken
	resData.TotalCount = models.QueryUserCourseCount(userId)
	if resData.TotalCount > 0 {
		// Get details of user's courses
		rcdList := handler.GetUserCourse(userId, currentPage, pageSize)
		resData.CourseInfo = rcdList
	}
	resData.Code = 200
	resData.Mesg = "success"
	u.RetDetailData(resData)
	return
}
