package handler

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"os"
	"path/filepath"
	"playground_backend/common"
	"playground_backend/models"
	"sort"
	"strconv"
	"strings"
)

type StatisticsData struct {
	UserId        int64
	UserName      string
	UserEmail     string
	OperationTime string
	EventType     string
	Course        CourseData
	State         string
	StateMessage  string
}

type CourseData struct {
	CourseId      string
	CourseName    string
	chapterId     string
	ChapterName   string
	CourseDur     string
	ChapterDur    string
	CourseStatus  int
	ChapterStatus int
	ResName       string
}

func CreateStatistLog(logFile string) (string, error) {
	configPath := beego.AppConfig.String("statistics::local_dir")
	common.CreateDir(configPath)
	if len(logFile) == 0 {
		logFile = beego.AppConfig.String("statistics::log_file")
	}
	filePath := filepath.Join(configPath, logFile)
	if !common.FileExists(filePath) {
		f, err := os.Create(filePath)
		if err != nil {
			logs.Error("Failed to create file, err: ", err, ",filePath: ", filePath)
			return "", err
		}
		defer f.Close()
	}
	return filePath, nil
}

func ConvertStrToInt(num string) int64 {
	intNum, _ := strconv.ParseInt(num, 10, 64)
	return intNum
}

func RenameStatistLog(filePath string) error {
	dir := beego.AppConfig.String("statistics::local_dir")
	fileSuffix := beego.AppConfig.String("statistics::log_file_suffix")
	files, _ := ioutil.ReadDir(dir)
	if len(files) > 0 {
		fileName := ""
		nameList := make([]string, 0)
		for _, f := range files {
			nameList = append(nameList, f.Name())
		}
		sort.Strings(nameList)
		lastFile := nameList[len(nameList)-1]
		splitFile := strings.Split(lastFile, ".log")
		if len(splitFile) < 2 {
			fileName = lastFile + fileSuffix
		} else {
			intNum := ConvertStrToInt(splitFile[1]) + 1
			format := "%0" + strconv.Itoa(len(fileSuffix)) + "d"
			fileName = lastFile + fmt.Sprintf(format, intNum)
		}
		err := os.Rename(filePath, fileName)
		if err != nil {
			logs.Error("file renaming failed, ", filePath, "====>", fileName)
			return err
		}
		CreateStatistLog(filePath)
	}
	return nil
}

func SplitStatistLog(filePath string) error {
	f, err := os.Stat(filePath)
	if err != nil {
		logs.Error("Failed to get file attributes, err: ", err, ",filePath: ", filePath)
		return err
	}
	logFileSize, err := beego.AppConfig.Int64("statistics::log_file_size")
	if err != nil {
		logs.Error("Failed to get split size of file, err: ", err, ",filePath: ", filePath)
		return err
	}
	if f.Size() > logFileSize {
		err = RenameStatistLog(filePath)
		if err != nil {
			logs.Error("RenameStatistLog, Failed to split file, err:", err)
			return err
		}
	}
	return nil
}

func WriteStatistLog(filePath string, byteData []byte) error {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0600)
	defer f.Close()
	if err != nil {
		logs.Error("fail to open the file, err: ", err, ",filePath: ", filePath)
		return err
	}
	_, err = f.Write(byteData)
	return nil
}

func DataFormatConver(sd StatisticsData) []byte {
	data := fmt.Sprintf("[%v/%v]operationTime:%v,"+
		"userId:%v,userName:%v,userEmail:%v,eventType:%v,"+
		"courseId:%v,courseName:%v,chapterId:%v,chapterName:%v,"+
		"courseDur:%v,chapterDur:%v,courseStatus:%v,chapterStatus:%v," +
		"resName:%v,state:%v,stateMessage:%v",
		common.GetCurTime(), sd.UserName, sd.OperationTime, sd.UserId, sd.UserName,
		sd.UserEmail, sd.EventType, sd.Course.CourseId, sd.Course.CourseName, sd.Course.chapterId,
		sd.Course.ChapterName, sd.Course.CourseDur, sd.Course.ChapterDur,
		sd.Course.CourseStatus, sd.Course.ChapterStatus,
		sd.Course.ResName, sd.State, sd.StateMessage)
	return []byte(data)
}

func StatisticsLog(sd StatisticsData) error {
	// 0. Query login information
	if sd.UserId > 0 && (len(sd.UserName) < 1 || len(sd.UserEmail) < 1) {
		gui := models.GiteeUserInfo{UserId: sd.UserId}
		queryErr := models.QueryGiteeUserInfo(&gui, "UserId")
		if queryErr == nil {
			sd.UserName = gui.UserLogin
			sd.UserEmail = gui.UserEmail
		}
	}
	// 1. Create a log file
	filePath, fErr := CreateStatistLog("")
	if fErr != nil {
		logs.Error("StatisticsLog, Failed to create log file, fErr: ", fErr)
		return fErr
	}
	// 2. Determine the file size and split large files
	splErr := SplitStatistLog(filePath)
	if splErr != nil {
		logs.Error("StatisticsLog, File segmentation failed, splErr: ", splErr)
		return splErr
	}
	// 3. Convert the data format to a writable file format
	byteData := DataFormatConver(sd)
	// 4. Write the data to a file in a fixed format
	writeErr := WriteStatistLog(filePath, byteData)
	if writeErr != nil {
		logs.Error("StatisticsLog, Failed to write data, writeErr: ", writeErr)
		return writeErr
	}
	return nil
}

func WriteCourseData(userId int64, courseId, ChapterId, eventType, resName,
	states, stateMes string, courseStatus, status int,
	crd *models.Courses, ccp *models.CoursesChapter) {
	sd := StatisticsData{UserId: userId, UserName: "",
		OperationTime: common.GetCurTime(), EventType: eventType,
		State: states, StateMessage: stateMes}
	cd := CourseData{}
	cd.CourseId = courseId
	models.QueryCourse(crd, "CourseId")
	if crd.Id > 0 {
		cd.CourseName = crd.Title
		cd.CourseDur = crd.Estimated
	}
	cd.CourseStatus = courseStatus
	cd.chapterId = ChapterId
	cd.ChapterStatus = status
	cd.ResName = resName
	models.QueryCourseChapter(ccp, "CourseId", "ChapterId")
	if ccp.Id > 0 {
		cd.ChapterName = ccp.Title
		cd.ChapterDur = ccp.Estimated
	}
	sd.Course = cd
	sdErr := StatisticsLog(sd)
	if sdErr != nil {
		logs.Error("CourseChapterControllers, post, sdErr: ", sdErr)
	}
}
