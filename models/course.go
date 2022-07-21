package models

import (
	"fmt"
	"playground_backend/common"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
)

func QueryAllCourseData(status int) (cs []Courses) {
	o := orm.NewOrm()
	num := int64(0)
	err := errors.New("")
	if status > 0 {
		num, err = o.Raw("SELECT * FROM pg_courses where status = ?", status).QueryRows(&cs)
	} else {
		num, err = o.Raw("SELECT * FROM pg_courses").QueryRows(&cs)
	}

	if num > 0 {
		logs.Info("QueryUserCourseData, search num: ", num)
	}
	if err != nil {
		logs.Error("QueryUserCourseData, cur_time:",
			common.GetCurTime(), ",err: ", err.Error())
	}
	return
}

func QueryCourse(eoi *Courses, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertCourse(eoi *Courses) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateCourse(eoi *Courses, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func UpdateCourseByCId(courseId, eulerBranch string) error {
	o := orm.NewOrm()
	err := o.Raw("update pg_courses set euler_branch = ?,update_time = ? where course_id = ?",
		eulerBranch, common.GetCurTime(), courseId).QueryRow()
	return err
}

func UpdateCourseFlag(flag int) error {
	o := orm.NewOrm()
	err := o.Raw("update pg_courses set flag = ?",
		flag).QueryRow()
	return err
}

func QueryAllCourseChapterData(status int) (cs []CoursesChapter) {
	o := orm.NewOrm()
	num := int64(0)
	err := errors.New("")
	if status > 0 {
		num, err = o.Raw("SELECT * FROM pg_courses_chapter where status = ?", status).QueryRows(&cs)
	} else {
		num, err = o.Raw("SELECT * FROM pg_courses_chapter").QueryRows(&cs)
	}

	if err == nil && num > 0 {
		logs.Info("QueryAllCourseChapterData, search num: ", num)
	} else {
		logs.Error("QueryAllCourseChapterData, cur_time:",
			common.GetCurTime(), ",err: ", err)
	}
	return
}

func QueryAllCourseChapterById(courseId string) (cs []CoursesChapter) {
	o := orm.NewOrm()
	num := int64(0)
	err := errors.New("")
	num, err = o.Raw("SELECT * FROM pg_courses_chapter where course_id = ?", courseId).QueryRows(&cs)
	if err == nil && num > 0 {
		logs.Info("QueryAllCourseChapterById, search num: ", num)
	} else {
		logs.Error("QueryAllCourseChapterById, cur_time:",
			common.GetCurTime(), ",err: ", err)
	}
	return
}

func QueryCourseChapter(eoi *CoursesChapter, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertCourseChapter(eoi *CoursesChapter) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateCourseChapter(eoi *CoursesChapter, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

// Update all chapters under the course
func UpdateCourseAllChapter(status, flag int8, courseId string) error {
	o := orm.NewOrm()
	deletTime := ""
	if flag == 1 {
		deletTime = common.GetCurTime()
	}
	err := o.Raw("update pg_courses_chapter set status = ?,delete_time = ? where course_id = ?",
		status, deletTime, courseId).QueryRow()
	if err != nil && !strings.Contains(err.Error(), "no row found") {
		logs.Info("UpdateCourseAllChapter err:", err.Error())
	}
	return err
}

func QueryUserCourse(eoi *UserCourse, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertUserCourse(eoi *UserCourse) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateUserCourse(eoi *UserCourse, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryUserCourseChapter(eoi *UserCourseChapter, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertUserCourseChapter(eoi *UserCourseChapter) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateUserCourseChapter(eoi *UserCourseChapter, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func UpdateUserCourseCompleted(CompletedFlag int, courseId string, userId int64) error {
	o := orm.NewOrm()
	err := o.Raw("update pg_user_course set completed_flag = ?,update_time = ? where course_id = ? and user_id = ?",
		CompletedFlag, common.GetCurTime(), courseId, userId).QueryRow()
	logs.Info("UpdateUserCourseCompleted", err.Error())
	return err
}

func UpdateUserCourseByCourseId(status int, courseId string) error {
	o := orm.NewOrm()
	deletTime := common.GetCurTime()
	err := o.Raw("update pg_user_course set status = ?,delete_time = ? where course_id = ? and status=?",
		status, deletTime, courseId, 1).QueryRow()
	logs.Info("UpdateUserCourseByCourseId", err.Error())
	return err
}

func UpdateUserCourseChapterByCourseId(status int, courseId string) error {
	o := orm.NewOrm()
	deletTime := common.GetCurTime()
	err := o.Raw("update pg_user_course_chapter set status = ?,delete_time = ? where course_id = ? and status = ?",
		status, deletTime, courseId, 1).QueryRow()
	logs.Info("UpdateUserCourseChapterByCourseId", err.Error())
	return err
}

func UpdateUserCourseChapterByChapterId(status int, courseId, chapterId string) error {
	o := orm.NewOrm()
	deletTime := common.GetCurTime()
	err := o.Raw("update pg_user_course_chapter set status = ?,"+
		"delete_time = ? where course_id = ? and chapter_id = ? and status = ?",
		status, deletTime, courseId, chapterId, 1).QueryRow()
	//logs.Info("UpdateUserCourseChapterByChapterId", err.Error())
	return err
}

func QueryUserCourseCount(userId int64) (count int64) {
	sql := ""
	sql = fmt.Sprintf(`SELECT COUNT(id) total FROM pg_user_course where user_id = %d order by id desc`, userId)
	res := struct {
		Total int64
	}{}
	o := orm.NewOrm()
	err := o.Raw(sql).QueryRow(&res)
	if err != nil {
		logs.Error("QueryUserCourseCount, err: ", err.Error())
		return 0
	}
	return res.Total
}

func QueryUserCourseData(currentPage, pageSize int, UserId int64) (uc []UserCourse) {
	startSize := (currentPage - 1) * pageSize
	o := orm.NewOrm()
	num, err := o.Raw("SELECT * FROM pg_user_course where user_id = ? order by id desc limit ? offset ?",
		UserId, pageSize, startSize).QueryRows(&uc)
	if num > 0 {
		logs.Info("QueryUserCourseData, search num: ", num)
	}
	if err != nil {
		logs.Info("QueryUserCourseData, cur_time:",
			common.GetCurTime(), ",err: ", err.Error())
	}
	return
}

func QueryChapterByCourseId(courseId string, UserId int64) (ucp []UserCourseChapter) {
	o := orm.NewOrm()
	num, err := o.Raw("SELECT * FROM pg_user_course_chapter where "+
		"user_id = ? and course_id = ? order by id asc",
		UserId, courseId).QueryRows(&ucp)
	if num > 0 {
		logs.Info("QueryChapterByCourseId, search num: ", num)
	}
	if err != nil {
		logs.Info("QueryChapterByCourseId, cur_time:",
			common.GetCurTime(), ",err: ", err.Error())
	}
	return
}
