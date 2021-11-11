package models

import "github.com/astaxie/beego/orm"

func QueryResourceInfo(eoi *ResourceInfo, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertResourceInfo(eoi *ResourceInfo) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateResourceInfo(eoi *ResourceInfo, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryResourceConfigPath(eoi *ResourceConfigPath, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

func QueryUserResourceEnv(eoi *UserResourceEnv, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertUserResourceEnv(eoi *UserResourceEnv) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateUserResourceEnv(eoi *UserResourceEnv, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}
