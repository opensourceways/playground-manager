package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type AuthUserInfo struct {
	UserId              int64  `orm:"pk;auto;column(id)"`
	SubUid              string `orm:"size(256);column(uid);unique" description:"用户的uuid"`
	Name                string `orm:"size(512);column(name)"description:"姓名"`
	UserName            string `orm:"size(512);column(user_name)"description:"用户名"`
	PhoneNumber         string `orm:"size(32);column(phone_number)"description:"手机号"`
	PhoneNumberVerified int8   `orm:"column(phone_number_verified)" description:"手机号是否认证, 0: 未认证;1:已认证"`
	NickName            string `orm:"size(512);column(nick_name)"description:"昵称"`
	Picture             string `orm:"type(text);column(picture)"description:"头像"`
	Email               string `orm:"size(512);column(email)"description:"电子邮箱"`
	EmailVerified       int8   `orm:"column(email_verified)" description:"邮箱是否被认证, 0: 未认证;1:已认证"`
	Status              int8   `orm:"column(status)" description:"1:正常;2:已删除"`
	AccessToken         string `orm:"type(text);column(access_token)"`
	ExpirationTime      string `orm:"size(32);column(expiration_time)" description:"token的过期时间"`
	CreateTime          string `orm:"size(32);column(create_time);"description:"创建时间"`
	UpdateTime          string `orm:"size(32);column(update_time);null"description:"更新时间"`
	DeleteTime          string `orm:"size(32);column(delete_time);null"description:"删除时间"`
}

type AuthUserDetail struct {
	UserDetailId      int64  `orm:"pk;auto;column(id)"`
	UserId            int64  `orm:"column(user_id);index" description:"用户id"`
	IdentityId        string `orm:"size(256);column(identity_id);unique" description:"外部身份识别id"`
	GivenName         string `orm:"size(512);column(given_name)"description:"名字"`
	FamilyName        string `orm:"size(512);column(family_name)"description:"姓氏"`
	MiddleName        string `orm:"size(512);column(middle_name)"description:"中间名"`
	PreferredNsername string `orm:"size(512);column(preferred_username)"description:"希望被称呼的名字"`
	Profile           string `orm:"size(32);column(profile)" description:"基础资料"`
	Website           string `orm:"type(text);colnum(website)"description:"网站链接"`
	Gender            string `orm:"size(512);column(gender)"description:"性别"`
	Birthdate         string `orm:"size(512);column(birthdate)"description:"生日"`
	Zoneinfo          string `orm:"size(512);colnum(zoneinfo)"description:"时区"`
	Locale            string `orm:"size(512);column(locale)"description:"区域"`
	Formatted         string `orm:"size(512);column(formatted)"description:"详细地址"`
	StreetAddress     string `orm:"size(512);column(street_address)"description:"街道地址"`
	Locality          string `orm:"size(512);column(locality)"description:"城市"`
	Region            string `orm:"size(512);column(region)"description:"省"`
	PostalCode        string `orm:"size(512);column(postal_code)"description:"邮编"`
	Country           string `orm:"size(512);column(country)"description:"国家"`
	Unionid           string `orm:"size(512);column(union_id)"description:""`
	Openid            string `orm:"size(512);column(openid)"description:""`
	CurStatus         string `orm:"size(128);column(cur_status);"description:"是否激活"`
	UpdatedAt         string `orm:"size(64);column(updated_at)"description:"信息更新时间"`
	CreatedAt         string `orm:"size(64);column(created_at)"description:""`
	Provider          string `orm:"size(512);column(provider)"description:""`
	ExtIdpId          string `orm:"size(512);column(ex_id)"description:""`
	UserIdInIdp       string `orm:"size(512);column(user_in_id)"description:""`
	UserName          string `orm:"size(512);column(user_name)"description:""`
	NickName          string `orm:"size(512);column(nick_name)"description:""`
	Photo             string `orm:"size(512);column(photo)"description:""`
	Company           string `orm:"size(512);column(company)"description:""`
	City              string `orm:"size(512);column(city)"description:""`
	Email             string `orm:"size(512);column(email)"description:""`
	CreateTime        string `orm:"size(32);column(create_time);"description:"创建时间"`
	UpdateTime        string `orm:"size(32);column(update_time);null"description:"更新时间"`
	DeleteTime        string `orm:"size(32);column(delete_time);null"description:"删除时间"`
}

type AuthTokenInfo struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	UserId       int64  `orm:"column(user_id);unique" description:"用户id"`
	AccessToken  string `orm:"type(text);column(access_token)"`
	TokenType    string `orm:"size(128);column(token_type)"`
	ExpiresIn    int64  `orm:"colnum(expires_in)"`
	RefreshToken string `orm:"type(text);column(refresh_token)"`
	IdToken      string `orm:"type(text);column(id_token)"`
	AuthCode     string `orm:"size(512);column(auth_code)"`
	Scope        string `orm:"type(text);column(scope)"`
	CreatedAt    int64  `orm:"colnum(created_at)"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

type ResourceInfo struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"column(user_id);index" description:"用户id"`
	ResourceName  string `orm:"size(256);column(res_name);unique"`
	ResourceAlias string `orm:"size(256);column(res_alias);unique"`
	Subdomain     string `orm:"size(256);column(sub_domain)"`
	UserName      string `orm:"size(256);column(user_name)"`
	PassWord      string `orm:"size(256);column(pass_word)"`
	ResourId      string `orm:"size(256);column(res_id)"`
	KindName      string `orm:"size(256);column(kind_name)"`
	RemainTime    int64  `orm:"colnum(remain_time)"`
	CompleteTime  int64  `orm:"colnum(complete_time)"`
	CreateTime    string `orm:"size(32);column(create_time);"`
	UpdateTime    string `orm:"size(32);column(update_time);null"`
	DeleteTime    string `orm:"size(32);column(delete_time);null"`
}

type ResourceConfigPath struct {
	Id              int64  `orm:"pk;auto;column(id)"`
	ResourceId      string `orm:"size(32);column(resource_id);unique"`
	EulerBranch     string `orm:"size(512);column(euler_branch)"`
	ResourcePath    string `orm:"size(512);column(resource_path)"`
	ResourceContent string `orm:"type(text);column(resource_content)"`
	EncryptionType  string `orm:"size(32);column(encrypt_type)"`
}

type UserResourceEnv struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	UserId       int64  `orm:"column(user_id);index" description:"用户id"`
	ResourceId   string `orm:"size(32);column(resource_id)"`
	CourseId     string `orm:"size(128);column(course_id);index" description:"课程id"`
	ChapterId    string `orm:"size(256);column(chapter_id);index" description:"课程章节id"`
	TemplatePath string `orm:"size(512);column(template_path)"`
	ContactEmail string `orm:"size(256);column(contact_email)"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

type ResourceTempathRel struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	ResourceId   string `orm:"size(32);column(resource_id)"`
	CourseId     string `orm:"size(128);column(course_id);index" description:"课程id"`
	ResourcePath string `orm:"size(512);column(resource_path)"`
	ResPoolSize  int    `orm:"colnum(pool_size);default(10)" description:"每个课程当前已申请的资源空闲数量，默认：5"`
	ResAlarmSize int    `orm:"colnum(alarm_size);default(1)" description:"每个课程当前已空闲的数量低于当前值，就开始告警，默认：1"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
}

type Courses struct {
	Id          int64  `orm:"pk;auto;column(id)"`
	CourseId    string `orm:"size(128);column(course_id);unique" description:"课程id"`
	Name        string `orm:"size(256);column(course_name)"`
	Title       string `orm:"size(256);column(course_title)"`
	Description string `orm:"type(text);column(course_desc)"`
	Icon        string `orm:"size(256);column(course_icon)"`
	Poster      string `orm:"size(256);column(course_poster)"`
	Banner      string `orm:"size(256);column(course_banner)"`
	EulerBranch string `orm:"size(512);column(euler_branch)"`
	Estimated   string `orm:"size(32);column(estimated_time)" description:"课程容器可用时间，单位：min"`
	Status      int8   `orm:"default(1);column(status)" description:"1: 正常；2:下线/删除"`
	Flag        int8   `orm:"default(1);column(flag)" description:"1: 正常；2:正在处理中"`
	CreateTime  string `orm:"size(32);column(create_time);"`
	UpdateTime  string `orm:"size(32);column(update_time);null"`
	DeleteTime  string `orm:"size(32);column(delete_time);null"`
}

type CoursesChapter struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	CId          int64  `orm:"column(c_id)" description:"课程内部id"`
	CourseId     string `orm:"size(128);column(course_id);index" description:"课程id"`
	ChapterId    string `orm:"size(256);column(chapter_id);index" description:"课程章节id"`
	Title        string `orm:"size(256);column(chapter_name)"`
	Description  string `orm:"type(text);column(chapter_desc)"`
	ResourcePath string `orm:"size(512);column(resource_path)"`
	EulerBranch  string `orm:"size(512);column(euler_branch)"`
	Estimated    string `orm:"size(32);column(estimated_time)" description:"章节学习预计完成时间，单位：min"`
	Status       int8   `orm:"default(1);column(status)" description:"1: 正常；2:下线/删除"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

type UserCourse struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"column(user_id);index" description:"用户id"`
	CId           int64  `orm:"column(c_id)" description:"内部课程id"`
	CourseId      string `orm:"size(128);column(course_id);index" description:"外部课程id"`
	CourseName    string `orm:"size(256);column(course_name)" description:"外部课程名称"`
	CompletedFlag int    `orm:"colnum(completed_flag);default(1)" description:"1: 课程学习中; 2: 课程完成学习"`
	StudyTime     int64  `orm:"column(study_time)" description:"课程学习时长, 单位：秒"`
	Status        int8   `orm:"default(1);column(status)" description:"1: 正常；2:下线/删除"`
	CreateTime    string `orm:"size(32);column(create_time);"`
	UpdateTime    string `orm:"size(32);column(update_time);null"`
	DeleteTime    string `orm:"size(32);column(delete_time);null"`
}

type UserCourseChapter struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"column(user_id);index" description:"用户id"`
	CId           int64  `orm:"column(c_id)" description:"内部课程id"`
	TId           int64  `orm:"column(t_id)" description:"内部章节id"`
	UcId          int64  `orm:"column(uc_id);index" description:"用户课程内部id"`
	CourseId      string `orm:"size(128);column(course_id);index" description:"外部课程id"`
	CourseName    string `orm:"size(256);column(course_name)" description:"外部课程名称"`
	ChapterId     string `orm:"size(256);column(chapter_id);index" description:"外部课程的章节id"`
	ChapterName   string `orm:"size(256);column(chapter_name)" description:"外部课程的章节名称"`
	CompletedFlag int    `orm:"colnum(completed_flag);default(1)" description:"1: 课程章节学习中; 2: 课程章节完成学习"`
	ResourcePath  string `orm:"size(512);column(resource_path)"`
	StudyTime     int64  `orm:"column(study_time)" description:"章节学习时长, 单位：秒"`
	Status        int8   `orm:"default(1);column(status)" description:"1: 正常；2:下线/删除"`
	CreateTime    string `orm:"size(32);column(create_time);"`
	UpdateTime    string `orm:"size(32);column(update_time);null"`
	DeleteTime    string `orm:"size(32);column(delete_time);null"`
}

func CreateDb() bool {
	BConfig, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		logs.Error("config init error:", err.Error())
		return false
	}
	prefix := BConfig.String("mysql::dbprefix")
	InitdbType, _ := beego.AppConfig.Int("initdb")
	if InitdbType == 1 {
		orm.RegisterModelWithPrefix(prefix,
			new(AuthUserDetail),
			new(AuthUserInfo), new(AuthTokenInfo),
			new(ResourceInfo), new(ResourceConfigPath),
			new(UserResourceEnv), new(ResourceTempathRel),
			new(Courses), new(CoursesChapter),
			new(UserCourse), new(UserCourseChapter),
		)
		logs.Info("table create success!")
		errosyn := orm.RunSyncdb("default", false, true)
		if errosyn != nil {
			logs.Error(errosyn)
		}
	}
	return true
}
