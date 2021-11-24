package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type GiteeUserInfo struct {
	UserId         int64  `orm:"pk;auto;column(id)"`
	GitId          int64  `orm:"column(git_id);unique" description:"git内部id"`
	UserName       string `orm:"size(512);column(user_name)"`
	UserLogin      string `orm:"size(512);column(user_login)"`
	UserUrl        string `orm:"size(512);colnum(avatar_url)"`
	AvatarUrl      string `orm:"size(512);colnum(user_url)"`
	AccessToken    string `orm:"size(512);column(access_token)"`
	ExpirationTime string `orm:"size(32);column(expiration_time)" description:"token的过期时间"`
	CreateTime     string `orm:"size(32);column(create_time);"`
	UpdateTime     string `orm:"size(32);column(update_time);null"`
	DeleteTime     string `orm:"size(32);column(delete_time);null"`
}

type GiteeTokenInfo struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	UserId       int64  `orm:"column(user_id);unique" description:"用户id"`
	AccessToken  string `orm:"size(512);column(access_token)"`
	TokenType    string `orm:"size(128);column(token_type)"`
	ExpiresIn    int64  `orm:"colnum(expires_in)"`
	RefreshToken string `orm:"size(512);column(refresh_token)"`
	AuthCode     string `orm:"size(512);column(auth_code)"`
	Scope        string `orm:"type(text);column(scope)"`
	CreatedAt    int64  `orm:"colnum(created_at)"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

type ResourceInfo struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	UserId       int64  `orm:"column(user_id);index" description:"用户id"`
	ResourceName string `orm:"size(256);column(res_name);unique"`
	Subdomain    string `orm:"size(256);column(sub_domain)"`
	UserName     string `orm:"size(256);column(user_name)"`
	PassWord     string `orm:"size(256);column(pass_word)"`
	ResourId     string `orm:"size(256);column(res_id)"`
	KindName     string `orm:"size(256);column(kind_name)"`
	RemainTime   int64  `orm:"colnum(remain_time)"`
	CompleteTime int64  `orm:"colnum(complete_time)"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

type ResourceConfigPath struct {
	Id              int64  `orm:"pk;auto;column(id)"`
	ResourceId      string `orm:"size(256);column(resource_id);unique"`
	ResourcePath    string `orm:"size(512);column(resource_path)"`
	ResourceContent string `orm:"type(text);column(resource_content)"`
	EncryptionType  string `orm:"size(32);column(encrypt_type)"`
}

type UserResourceEnv struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	UserId       int64  `orm:"column(user_id);index" description:"用户id"`
	ResourceId   string `orm:"size(256);column(resource_id);unique"`
	TemplatePath string `orm:"size(512);column(template_path)"`
	ContactEmail string `orm:"size(256);column(contact_email)"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

func CreateDb() bool {
	BConfig, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		logs.Error("config init error:", err)
		return false
	}
	prefix := BConfig.String("mysql::dbprefix")
	InitdbType, _ := beego.AppConfig.Int("initdb")
	if InitdbType == 1 {
		orm.RegisterModelWithPrefix(prefix,
			new(GiteeUserInfo), new(GiteeTokenInfo),
			new(ResourceInfo), new(ResourceConfigPath),
			new(UserResourceEnv),
		)
		logs.Info("table create success!")
		errosyn := orm.RunSyncdb("default", false, true)
		if errosyn != nil {
			logs.Error(errosyn)
		}
	}
	return true
}
