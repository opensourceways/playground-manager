package models

import "github.com/astaxie/beego/orm"

func QueryGiteeUserInfo(eoi *GiteeUserInfo, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertGiteeUserInfo(eoi *GiteeUserInfo) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateGiteeUserInfo(eoi *GiteeUserInfo, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryGiteeTokenInfo(eoi *GiteeTokenInfo, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertGiteeTokenInfo(eoi *GiteeTokenInfo) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateGiteeTokenInfo(eoi *GiteeTokenInfo, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}
