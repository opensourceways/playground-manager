package handler

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/pkg/errors"
	"playground_backend/common"
	"playground_backend/http"
	"playground_backend/models"
	"strings"
)

type CourseReqParameter struct {
	Token       string                `json:"token"`
	UserId      int64                 `json:"userId"`
	CourseId    string                `json:"courseId"`
	CourseName  string                `json:"courseName"`
	Status      int                   `json:"status"`
	ChapterInfo []ChapterReqParameter `json:"chapterInfo"`
}

type ChapterReqParameter struct {
	ChapterId   string `json:"chapterId"`
	ChapterName string `json:"chapterName"`
	Status      int    `json:"status"`
}

type RspCourse struct {
	CId   int64  `json:"cId"`
	State string `json:"state"`
}

type RspCourseData struct {
	CourseId    string                 `json:"courseId"`
	CourseName  string                 `json:"courseName"`
	Status      int                    `json:"status"`
	IsOnline    int8                   `json:"isOnline"`
	ChapterData []RspCourseChapterData `json:"chapterInfo"`
}

type RspCourseChapterData struct {
	ChapterId   string `json:"chapterId"`
	ChapterName string `json:"chapterName"`
	Status      int    `json:"status"`
	IsOnline    int8   `json:"isOnline"`
}

type EnvPrams struct {
	OnlineEnv        string
	OfflineEnv       string
	CourseUrl        string
	ChapterUrl       string
	ChapterDetailUrl string
}

type EulerBranch struct {
	imageid string
}

type ChapterDetailData struct {
	Title       string
	Description string
	Details     interface{}
	Environment interface{}
	Backend     EulerBranch
}

func addUserCourse(crp CourseReqParameter, uc *models.UserCourse, flag int) {
	uc.CourseId = crp.CourseId
	uc.UserId = crp.UserId
	uc.Status = 1
	uc.CourseName = crp.CourseName
	uc.CompletedFlag = crp.Status
	if flag == 1 {
		uc.CreateTime = common.GetCurTime()
	} else {
		uc.UpdateTime = common.GetCurTime()
	}
}

func UserBoundBourse(crp CourseReqParameter) int64 {
	crd := models.Courses{CourseId: crp.CourseId}
	ccp := models.CoursesChapter{CourseId: crp.CourseId, ChapterId: ""}
	WriteCourseData(crp.UserId, crp.CourseId, "", "User bound course",
		"", "Processing", "",
		crp.Status, crp.Status, &crd, &ccp)
	uc := models.UserCourse{UserId: crp.UserId, CourseId: crp.CourseId}
	ucId := int64(0)
	queryErr := models.QueryUserCourse(&uc, "UserId", "CourseId")
	if uc.Id > 0 || queryErr == nil {
		addUserCourse(crp, &uc, 2)
		uc.CId = crd.Id
		upErr := models.UpdateUserCourse(&uc, "CourseName", "CompletedFlag", "UpdateTime", "CId")
		if upErr != nil {
			logs.Error("UserBoundBourse, upErr: ", upErr)
			WriteCourseData(crp.UserId, crp.CourseId, "", "User bound course",
				"", "failed", "User binding course failed",
				crp.Status, crp.Status, &crd, &ccp)
			return 0
		}
		ucId = uc.Id
	} else {
		addUserCourse(crp, &uc, 1)
		uc.CId = crd.Id
		id, inErr := models.InsertUserCourse(&uc)
		if inErr != nil {
			logs.Error("UserBoundBourse, inErr: ", inErr, ",id:", id)
			WriteCourseData(crp.UserId, crp.CourseId, "", "User bound course",
				"", "failed", "User binding course failed",
				crp.Status, crp.Status, &crd, &ccp)
			return 0
		}
		ucId = id
	}
	WriteCourseData(crp.UserId, crp.CourseId, "", "User bound course",
		"", "success", "User binding course successfully",
		crp.Status, crp.Status, &crd, &ccp)
	return ucId
}

func addUserCourseChapter(crp ChapterReqParameter, uc *models.UserCourseChapter,
	ucId, userId int64, courseId string) {
	uc.UcId = ucId
	uc.CourseId = courseId
	uc.UserId = userId
	uc.ChapterId = crp.ChapterId
	uc.ChapterName = crp.ChapterName
	uc.CompletedFlag = crp.Status
	uc.Status = 1
	uc.CreateTime = common.GetCurTime()
	uc.UpdateTime = common.GetCurTime()
}

func UserBoundBourseChapter(crp ChapterReqParameter, ucId, userId int64, courseId string, courseStatus int) {
	crd := models.Courses{CourseId: courseId}
	ccp := models.CoursesChapter{CourseId: courseId, ChapterId: crp.ChapterId}
	WriteCourseData(userId, courseId, crp.ChapterId, "User bound chapter",
		"", "Processing", "",
		courseStatus, crp.Status, &crd, &ccp)
	uc := models.UserCourseChapter{UserId: userId, CourseId: courseId, ChapterId: crp.ChapterId}
	queryErr := models.QueryUserCourseChapter(&uc, "UserId", "CourseId", "ChapterId")
	if uc.Id > 0 || queryErr == nil {
		uc.CId = crd.Id
		uc.TId = ccp.Id
		uc.CourseName = crd.Title
		addUserCourseChapter(crp, &uc, ucId, userId, courseId)
		upErr := models.UpdateUserCourseChapter(&uc, "UcId", "ChapterName",
			"CompletedFlag", "UpdateTime", "CId", "TId", "CourseName")
		if upErr != nil {
			logs.Error("UserBoundBourseChapter, upErr: ", upErr)
			WriteCourseData(userId, courseId, crp.ChapterId, "User bound chapter",
				"", "failed", "User binding course chapter failed",
				courseStatus, crp.Status, &crd, &ccp)
			return
		}
	} else {
		uc.CId = crd.Id
		uc.TId = ccp.Id
		uc.CourseName = crd.Title
		addUserCourseChapter(crp, &uc, ucId, userId, courseId)
		id, inErr := models.InsertUserCourseChapter(&uc)
		if inErr != nil {
			logs.Error("UserBoundBourseChapter, inErr: ", inErr, ",id:", id)
			WriteCourseData(userId, courseId, crp.ChapterId, "User bound chapter",
				"", "failed", "User binding course chapter failed",
				courseStatus, crp.Status, &crd, &ccp)
			return
		}
	}
	WriteCourseData(userId, courseId, crp.ChapterId, "User bound chapter",
		"", "success", "User binding course chapter successfully",
		courseStatus, crp.Status, &crd, &ccp)
}

func (ep EnvPrams) AddCourseToDb(cr interface{}) {
	cor := cr.(map[string]interface{})
	statusList := cor["status"].([]interface{})
	onEnvList := strings.Split(ep.OnlineEnv, ",")
	if len(statusList) > 0 && len(onEnvList) > 0 {
		for _, st := range statusList {
			status := st.(string)
			status = strings.ToLower(status)
			if status == ep.OfflineEnv {
				cr := models.Courses{}
				cr.CourseId = cor["id"].(string)
				queryErr := models.QueryCourse(&cr, "CourseId")
				if cr.Id > 0 {
					cr.Status = 2
					cr.DeleteTime = common.GetCurTime()
					delErr := models.UpdateCourse(&cr, "Status", "DeleteTime")
					if delErr == nil {
						delChapterErr := models.UpdateCourseAllChapter(cr.Status, 1, cr.CourseId)
						if delChapterErr != nil {
							logs.Error("AddCourseToDb, delChapterErr: ", delChapterErr)
						}
					}
				} else {
					logs.Info("AddCourseToDb, The course does not exist, "+
						"no need to go offline, queryErr: ", queryErr)
				}
			}
			for _, env := range onEnvList {
				if status == env {
					imageid := ""
					coursePathName := cor["name"].(string)
					courseId := cor["id"].(string)
					chapterUrl := fmt.Sprintf(ep.ChapterUrl, coursePathName)
					body, resErr := http.HTTPGitGet(chapterUrl)
					if resErr != nil {
						logs.Error("AddCourseToDb, resErr: ", resErr, ",body: ", body)
						continue
					}
					cId := int64(0)
					cr := models.Courses{}
					cr.CourseId = courseId
					queryErr := models.QueryCourse(&cr, "CourseId")
					if cr.Id > 0 {
						cr.Name = cor["name"].(string)
						AddCourseData(body, &cr)
						upErr := models.UpdateCourse(&cr, "Status", "Name",
							"Title", "Description", "Icon", "Poster", "Banner", "Estimated", "UpdateTime")
						if upErr != nil {
							logs.Error("AddCourseToDb, upErr: ", upErr)
						}
						cId = cr.Id
					} else {
						cr.CourseId = courseId
						cr.Name = cor["name"].(string)
						AddCourseData(body, &cr)
						id, inErr := models.InsertCourse(&cr)
						if inErr != nil {
							logs.Error("AddCourseToDb, inErr: ", inErr, ",queryErr: ", queryErr)
						}
						cId = id
					}
					upChapterErr := models.UpdateCourseAllChapter(2, 2, cr.CourseId)
					if upChapterErr != nil {
						logs.Error("AddCourseToDb, delChapterErr: ", upChapterErr)
					}
					chapterList := body["chapters"].([]interface{})
					if len(chapterList) > 0 {
						for _, ct := range chapterList {
							chapter := ct.(map[string]interface{})
							cdd := ChapterDetailData{}
							chapterId := chapter["chapter_id"].(string)
							ep.GetChapterDetail(coursePathName, chapterId, &cdd)
							imageid = cdd.Backend.imageid
							cp := models.CoursesChapter{CourseId: cr.CourseId, ChapterId: chapterId, EulerBranch: cdd.Backend.imageid}
							querychapterErr := models.QueryCourseChapter(&cp, "CourseId", "ChapterId")
							if cp.Id > 0 {
								AddChapterData(chapter, &cp, cId)
								cp.EulerBranch = cdd.Backend.imageid
								upChapterErr := models.UpdateCourseChapter(&cp,
									"Status", "Title", "Description", "Estimated", "UpdateTime", "EulerBranch")
								if upChapterErr != nil {
									logs.Error("UpdateCourseChapter, upChapterErr: ", upChapterErr)
								}
							} else {
								cp.CourseId = cr.CourseId
								cp.EulerBranch = cdd.Backend.imageid
								AddChapterData(chapter, &cp, cId)
								_, inChapterErr := models.InsertCourseChapter(&cp)
								if inChapterErr != nil {
									logs.Error("InsertCourseChapter, inChapterErr: ",
										inChapterErr, ",querychapterErr: ", querychapterErr)
								}
							}
						}
					}
					if len(imageid) > 1 {
						models.UpdateCourseByCId(courseId, imageid)
						ProcCourseAndResRel(courseId, imageid)
					}
				}
			}
		}
	}
}

func ProcCourseAndResRel(courseId, eulerBranch string) {
	rcp := models.ResourceConfigPath{EulerBranch: eulerBranch}
	rr := ReqResource{CourseId: courseId}
	rcpErr := rr.SaveCourseAndResRel(&rcp)
	if rcpErr != nil {
		logs.Error("ProcCourseAndResRel, rcpErr: ", rcpErr)
		return
	}
}

func (rr *ReqResource) SaveCourseAndResRel(rcp *models.ResourceConfigPath) error {
	rcpErr := errors.New("")
	if len(rcp.EulerBranch) > 1 {
		rcpErr = models.QueryResourceConfigPath(rcp, "EulerBranch")
	} else {
		rcpErr = models.QueryResourceConfigPath(rcp, "ResourcePath")
	}
	if rcpErr != nil {
		logs.Error("QueryResourceConfigPath, rcpErr: ", rcpErr)
		return rcpErr
	}
	rr.EnvResource = rcp.ResourcePath
	rr.ResourceId = rcp.ResourceId
	saveErr := SaveResourceTemplate(rr)
	return saveErr
}

func (ep EnvPrams) GetChapterDetail(coursePathName, chapterId string, cdd *ChapterDetailData) {
	chapterDetailUrl := fmt.Sprintf(ep.ChapterDetailUrl, coursePathName, chapterId)
	body, resErr := http.HTTPGitGet(chapterDetailUrl)
	if resErr != nil {
		logs.Error("AddCourseToDb, resErr: ", resErr, ",body: ", body)
		return
	}
	cdd.Title = body["title"].(string)
	cdd.Description = body["description"].(string)
	eulerBranch := body["backend"].(map[string]interface{})
	imageid, ok := eulerBranch["imageid"]
	if ok {
		cdd.Backend.imageid = imageid.(string)
	}
}

func AddChapterData(cor map[string]interface{}, cr *models.CoursesChapter, cId int64) {
	cr.Status = 1
	cr.CId = cId
	cr.ChapterId = cor["chapter_id"].(string)
	cr.Title = cor["title"].(string)
	cr.Description = cor["description"].(string)
	cr.Estimated = cor["estimated_time"].(string)
	cr.UpdateTime = common.GetCurTime()
	cr.CreateTime = common.GetCurTime()
}

func AddCourseData(cor map[string]interface{}, cr *models.Courses) {
	cr.Status = 1
	cr.Title = cor["title"].(string)
	cr.Description = cor["description"].(string)
	cr.Icon = cor["logo"].(string)
	cr.Poster = cor["poster"].(string)
	cr.Banner = cor["banner"].(string)
	cr.Estimated = cor["estimated_time"].(string)
	cr.UpdateTime = common.GetCurTime()
	cr.CreateTime = common.GetCurTime()
}

func (ep EnvPrams) ParsingCourse(body map[string]interface{}) error {
	courses, ok := body["courses"]
	if !ok {
		logs.Error("The course list file is abnormal and cannot be parsed")
		return errors.New("The course list file is abnormal and cannot be parsed")
	}
	cousrseList := courses.([]interface{})
	if len(cousrseList) > 0 {
		for _, cousrse := range cousrseList {
			ep.AddCourseToDb(cousrse)
		}
		return nil
	}
	logs.Error("ParsingCourse, Course list is empty")
	return errors.New("ParsingCourse, Course list is empty")
}

func SyncCourse() error {
	onlineEnv := beego.AppConfig.String("courses::online_env")
	offlineEnv := beego.AppConfig.String("courses::offline_env")
	courseUrl := beego.AppConfig.String("courses::course_url")
	chapterUrl := beego.AppConfig.String("courses::chapter_url")
	chapterDetailUrl := beego.AppConfig.String("courses::chapter_detail_url")
	body, resErr := http.HTTPGitGet(courseUrl)
	if resErr != nil {
		logs.Error("SyncCourse, resErr: ", resErr)
		return resErr
	}
	ep := EnvPrams{OnlineEnv: onlineEnv, OfflineEnv: offlineEnv, CourseUrl: courseUrl,
		ChapterUrl: chapterUrl, ChapterDetailUrl: chapterDetailUrl}
	pErr := ep.ParsingCourse(body)
	if pErr != nil {
		logs.Error("pErr: ", pErr)
	}
	return pErr
}

func GetUserCourse(userId int64, currentPage, pageSize int) (rcdList []RspCourseData) {
	// Inquire about course details
	ucList := models.QueryUserCourseData(currentPage, pageSize, userId)
	if len(ucList) > 0 {
		for _, uc := range ucList {
			crd := models.Courses{CourseId: uc.CourseId}
			ccp := models.CoursesChapter{CourseId: uc.CourseId}
			WriteCourseData(userId, uc.CourseId, "", "Query user's courses",
				"", "success", "Query the user has taken courses successfully",
				uc.CompletedFlag, uc.CompletedFlag, &crd, &ccp)
			rcd := RspCourseData{}
			rccdList := []RspCourseChapterData{}
			cour := models.Courses{CourseId: uc.CourseId}
			quyCourErr := models.QueryCourse(&cour, "CourseId")
			if quyCourErr == nil {
				uc.CourseName = cour.Title
				uc.Status = cour.Status
				ucpList := models.QueryChapterByCourseId(uc.CourseId, userId)
				if len(ucpList) > 0 {
					for _, ucp := range ucpList {
						ccp.ChapterId = ucp.ChapterId
						WriteCourseData(userId, uc.CourseId, ucp.ChapterId, "Query user's courses chapters",
							"", "success", "Query the user's learned course chapters successfully",
							uc.CompletedFlag, ucp.CompletedFlag, &crd, &ccp)
						ccp := models.CoursesChapter{CourseId: ucp.CourseId, ChapterId: ucp.ChapterId}
						quyChapterErr := models.QueryCourseChapter(&ccp, "CourseId", "ChapterId")
						if quyChapterErr == nil {
							ucp.ChapterName = ccp.Title
							ucp.Status = ccp.Status
							rccd := RspCourseChapterData{}
							RspChapter(ucp, &rccd)
							rccdList = append(rccdList, rccd)
						}
					}
				}
			}
			AddRspCourse(uc, &rcd)
			rcd.ChapterData = rccdList
			rcdList = append(rcdList, rcd)
		}
	}
	return
}

func AddRspCourse(uc models.UserCourse, rcd *RspCourseData) {
	rcd.IsOnline = uc.Status
	rcd.Status = uc.CompletedFlag
	rcd.CourseId = uc.CourseId
	rcd.CourseName = uc.CourseName
}

func RspChapter(ucp models.UserCourseChapter, rccd *RspCourseChapterData) {
	rccd.Status = ucp.CompletedFlag
	rccd.ChapterName = ucp.ChapterName
	rccd.ChapterId = ucp.ChapterId
	rccd.IsOnline = ucp.Status
}
