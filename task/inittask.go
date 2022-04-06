package task

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"playground_backend/handler"
)

// start task
func StartTask() {
	toolbox.StartTask()
}

func StopTask() {
	toolbox.StopTask()
}

// Clear used resource image instance resources
func ClearInstanceTask(clInvalidInstance string) {
	invalidTask := toolbox.NewTask("ClearInvaildResource",
		clInvalidInstance, handler.ClearInvaildResource)
	toolbox.AddTask("ClearInvaildResource", invalidTask)
}

// Synchronized course list and chapter information
func SyncCourseTask(syncCourse string) {
	syncCourseTask := toolbox.NewTask("SyncCourse", syncCourse, handler.SyncCourse)
	toolbox.AddTask("SyncCourse", syncCourseTask)
}

// Ensure that new courses can generate resource pools
func ApplyCoursePoolTask(applyCoursePool string) {
	applyCoursePoolTask := toolbox.NewTask("ApplyCoursePoolTask",
		applyCoursePool, handler.ApplyCoursePoolTask)
	toolbox.AddTask("ApplyCoursePoolTask", applyCoursePoolTask)
}

//InitTask Timing task initialization
func InitTask() bool {
	// Clear used resource image instance resources
	clInvalidInstanFlag, err := beego.AppConfig.Int("crontab::cl_invalid_instances_flag")
	if clInvalidInstanFlag == 1 && err == nil {
		clInvalidInstance := beego.AppConfig.String("crontab::cl_invalid_instances")
		ClearInstanceTask(clInvalidInstance)
	}

	// Synchronized course list and chapter information
	syncCourseFlag, err := beego.AppConfig.Int("crontab::sync_course_flag")
	if syncCourseFlag == 1 && err == nil {
		syncCourse := beego.AppConfig.String("crontab::sync_course")
		SyncCourseTask(syncCourse)
	}
	// Ensure that new courses can generate resource pools
	applyCourseFlag, err := beego.AppConfig.Int("crontab::apply_course_pool_flag")
	if applyCourseFlag == 1 && err == nil {
		applyCoursePool := beego.AppConfig.String("crontab::apply_course_pool")
		ApplyCoursePoolTask(applyCoursePool)
	}
	return true
}
