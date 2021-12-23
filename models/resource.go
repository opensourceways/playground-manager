package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

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

func QueryResourceTempathRel(eoi *ResourceTempathRel, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertResourceTempathRel(eoi *ResourceTempathRel) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateResourceTempathRel(eoi *ResourceTempathRel, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryResourceTempathRelAll() (ite []ResourceTempathRel, num int64, err error) {
	o := orm.NewOrm()
	num, err = o.Raw("select resource_id,resource_path"+
		" from pg_resource_tempath_rel").QueryRows(&ite)
	if err == nil && num > 0 {
		logs.Info("QueryResourceTempathRelAll, pg_resource_tempath_rel, search result: ", num)
	} else {
		logs.Error("QueryResourceTempathRelAll, err: ", err)
	}
	return
}
