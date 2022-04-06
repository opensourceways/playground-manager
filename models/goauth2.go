package models

import "github.com/astaxie/beego/orm"

func QueryAuthUserInfo(eoi *AuthUserInfo, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertAuthUserInfo(eoi *AuthUserInfo) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateAuthUserInfo(eoi *AuthUserInfo, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryAuthUserDetail(eoi *AuthUserDetail, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertAuthUserDetail(eoi *AuthUserDetail) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateAuthUserDetail(eoi *AuthUserDetail, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryAuthTokenInfo(eoi *AuthTokenInfo, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertAuthTokenInfo(eoi *AuthTokenInfo) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateAuthTokenInfo(eoi *AuthTokenInfo, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}
