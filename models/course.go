package models

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"playground_backend/common"
)

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
	logs.Info("UpdateCourseByCId", err)
	return err
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
	logs.Info("UpdateCourseAllChapter", err)
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

func QueryUserCourseCount(userId int64) (count int64) {
	sql := ""
	sql = fmt.Sprintf(`SELECT COUNT(id) total FROM pg_user_course where user_id = %d order by id desc`, userId)
	res := struct {
		Total int64
	}{}
	o := orm.NewOrm()
	err := o.Raw(sql).QueryRow(&res)
	if err != nil {
		logs.Error("QueryUserCourseCount, err: ", err)
		return 0
	}
	return res.Total
}

func QueryUserCourseData(currentPage, pageSize int, UserId int64) (uc []UserCourse) {
	startSize := (currentPage - 1) * pageSize
	o := orm.NewOrm()
	num, err := o.Raw("SELECT * FROM pg_user_course where user_id = ? order by id desc limit ? offset ?",
		UserId, pageSize, startSize).QueryRows(&uc)
	if err == nil && num > 0 {
		logs.Info("QueryUserCourseData, search num: ", num)
	} else {
		logs.Info("QueryUserCourseData, cur_time:",
			common.GetCurTime(), ",err: ", err)
	}
	return
}

func QueryChapterByCourseId(courseId string, UserId int64) (ucp []UserCourseChapter) {
	o := orm.NewOrm()
	num, err := o.Raw("SELECT * FROM pg_user_course_chapter where "+
		"user_id = ? and course_id = ? order by id asc",
		UserId, courseId).QueryRows(&ucp)
	if err == nil && num > 0 {
		logs.Info("QueryChapterByCourseId, search num: ", num)
	} else {
		logs.Info("QueryChapterByCourseId, cur_time:",
			common.GetCurTime(), ",err: ", err)
	}
	return
}
