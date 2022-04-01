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
	UserEmail      string `orm:"size(512);column(user_email)"`
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
	ResourceId      string `orm:"size(256);column(resource_id);unique"`
	EulerBranch     string `orm:"size(512);column(euler_branch)"`
	ResourcePath    string `orm:"size(512);column(resource_path)"`
	ResourceContent string `orm:"type(text);column(resource_content)"`
	EncryptionType  string `orm:"size(32);column(encrypt_type)"`
}

type UserResourceEnv struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	UserId       int64  `orm:"column(user_id);index" description:"用户id"`
	ResourceId   string `orm:"size(256);column(resource_id)"`
	CourseId     string `orm:"size(256);column(course_id);index" description:"课程id"`
	ChapterId    string `orm:"size(256);column(chapter_id);index" description:"课程章节id"`
	TemplatePath string `orm:"size(512);column(template_path)"`
	ContactEmail string `orm:"size(256);column(contact_email)"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

type ResourceTempathRel struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	ResourceId   string `orm:"size(256);column(resource_id)"`
	CourseId     string `orm:"size(256);column(course_id);index" description:"课程id"`
	ResourcePath string `orm:"size(512);column(resource_path)"`
	ResPoolSize  int    `orm:"colnum(pool_size);default(5)" description:"每个课程当前已申请的资源空闲数量，默认：5"`
	ResAlarmSize int    `orm:"colnum(alarm_size);default(1)" description:"每个课程当前已空闲的数量低于当前值，就开始告警，默认：1"`
}

type Courses struct {
	Id          int64  `orm:"pk;auto;column(id)"`
	CourseId    string `orm:"size(256);column(course_id);unique" description:"课程id"`
	Name        string `orm:"size(256);column(course_name)"`
	Title       string `orm:"size(256);column(course_title)"`
	Description string `orm:"type(text);column(course_desc)"`
	Icon        string `orm:"size(256);column(course_icon)"`
	Poster      string `orm:"size(256);column(course_poster)"`
	Banner      string `orm:"size(256);column(course_banner)"`
	EulerBranch string `orm:"size(512);column(euler_branch)"`
	Estimated   string `orm:"size(32);column(estimated_time)" description:"课程容器可用时间，单位：min"`
	Status      int8   `orm:"default(1);column(status)" description:"1: 正常；2:下线/删除"`
	CreateTime  string `orm:"size(32);column(create_time);"`
	UpdateTime  string `orm:"size(32);column(update_time);null"`
	DeleteTime  string `orm:"size(32);column(delete_time);null"`
}

type CoursesChapter struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	CId          int64  `orm:"column(c_id)" description:"课程内部id"`
	CourseId     string `orm:"size(256);column(course_id);index" description:"课程id"`
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
	CourseId      string `orm:"size(256);column(course_id);index" description:"外部课程id"`
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
	CourseId      string `orm:"size(256);column(course_id);index" description:"外部课程id"`
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
		logs.Error("config init error:", err)
		return false
	}
	prefix := BConfig.String("mysql::dbprefix")
	InitdbType, _ := beego.AppConfig.Int("initdb")
	if InitdbType == 1 {
		orm.RegisterModelWithPrefix(prefix,
			new(GiteeUserInfo), new(GiteeTokenInfo),
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
