package handler

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	ymV2 "gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"os"
	"path"
	"path/filepath"
	"playground_backend/common"
	"playground_backend/models"
	"strconv"
	"strings"
	"time"
)

type ResourceData struct {
	EnvResource  string
	ResourceId   string
	ResourceName string
}

type InitTmplResource struct {
	Name         string
	Subdomain    string
	NamePassword string
	UserId       string
	ContactEmail string
}

var CoursePoolVar = CoursePool{}

type CoursePool struct {
	InitialFlag bool
	CourseMap   map[string]chan InitTmplResource
}

func InitialMemoryRes() {
	CoursePoolVar.InitialFlag = false
	CoursePoolVar.CourseMap = make(map[string]chan InitTmplResource, 0)
}

func InitPoolTmplPrarse(rtp *InitTmplResource, rd *ResourceData, cr *CourseResources) {
	resourceName := ResName(rd.EnvResource)
	resName := "res" + rd.ResourceId + "-" + resourceName + "-" +
		strconv.FormatInt(time.Now().Unix(), 10) + common.RandomString(32)
	rtp.UserId = "default"
	resName = "res" + common.EncryptMd5(resName)
	cr.UserId = rtp.UserId
	cr.CourseId = rd.ResourceId
	cr.ResourceName = ResName(rd.EnvResource)
	rtp.Name = resName
	rd.ResourceName = resName
	subDomain := resName + rd.EnvResource + common.RandomString(32)
	subDomain = common.EncryptMd5(subDomain)
	if ok := common.IsLetter(rune(subDomain[0])); !ok {
		subDomain = strings.Replace(subDomain, subDomain[:3], "res", 1)
	}
	rtp.Subdomain = subDomain
	userName := resName + common.RandomString(32)
	passWord := common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(userName)))
	userName = common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(subDomain)))
	userName = userName[:16]
	rtp.NamePassword = userName + ":" + passWord
}

func PoolParseTmpl(yamlDir string, rd *ResourceData, localPath string) []byte {
	contactEmail := beego.AppConfig.DefaultString("template::contact_email", "contact@openeuler.io")
	rtp := InitTmplResource{ContactEmail: contactEmail}
	cr := CourseResources{}
	InitPoolTmplPrarse(&rtp, rd, &cr)
	var templates *template.Template
	var allFiles []string
	files, dirErr := ioutil.ReadDir(yamlDir)
	if dirErr != nil {
		logs.Error("dirErr: ", dirErr)
		return []byte{}
	}
	tmpLocalPath := localPath
	localPath = strings.ReplaceAll(localPath, "\\", "/")
	fileName := path.Base(localPath)
	for _, file := range files {
		pFileName := file.Name()
		if fileName == pFileName {
			fullPath := filepath.Join(yamlDir, pFileName)
			allFiles = append(allFiles, fullPath)
		}
	}
	logs.Info("allFiles: ", allFiles)
	tempErr := error(nil)
	templates, tempErr = template.ParseFiles(allFiles...)
	if tempErr != nil {
		logs.Error("tempErr: ", tempErr)
		return []byte{}
	}
	defer common.DelFile(allFiles)
	s1 := templates.Lookup(fileName)
	if s1 == nil {
		logs.Error("fileName is nil")
		return []byte{}
	}
	preFileName := common.GetRandomString(8)
	outFileName := preFileName + "-" + rtp.Name + ".yaml"
	outPutPath := filepath.Join(yamlDir, outFileName)
	f, ferr := os.Create(outPutPath)
	if ferr != nil {
		logs.Error("ferr: ", ferr)
		return []byte{}
	}
	exErr := s1.Execute(f, rtp)
	if exErr != nil {
		f.Close()
		logs.Error("exErr: ", exErr)
		return []byte{}
	}
	f.Close()
	content, fErr := common.ReadAll(outPutPath)
	if fErr != nil {
		logs.Error("common.ReadAll, fErr: ", fErr)
		return []byte{}
	}
	content = AddAnnotations(content, &cr)
	if common.FileExists(outPutPath) {
		DeleteFile(outPutPath)
	}
	if common.FileExists(tmpLocalPath) {
		DeleteFile(tmpLocalPath)
	}
	//UnstructuredYaml(content)
	return content
}

func CreateSingleRes(yamlData []byte, rd *ResourceData) error {
	var (
		err       error
		objCreate *unstructured.Unstructured
		gvk       *schema.GroupVersionKind
		dr        dynamic.ResourceInterface
	)
	obj := &unstructured.Unstructured{}
	_, gvk, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(yamlData, nil, obj)
	if err != nil {
		logs.Error("failed to get GVK, err: ", err)
		return err
	}
	dr, err = GetGVRdyClient(gvk, obj.GetNamespace(), rd.ResourceId)
	if err != nil {
		logs.Error("failed to get dr: ", err)
		return err
	}
	// store db
	config := new(YamlConfig)
	err = ymV2.Unmarshal(yamlData, config)
	if err != nil {
		logs.Error("yaml1.Unmarshal, err: ", err)
		return err
	}
	logs.Info("To start creating a resource, the resource name:", obj.GetName())
	objCreate, err = dr.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		logs.Error("Create err: ", err)
		return err
	}
	//PrintJsonStr(objCreate)
	rls := GetResInfo(objCreate, dr, config, obj, false)
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		logs.Info("Resource created successfully, resourceName: ", obj.GetName(), ", InstanceEndpoint: ", rls.InstanceEndpoint)
	} else {
		logs.Info("Resource is being created, resourceName: ", obj.GetName(), ", InstanceEndpoint: ", rls.InstanceEndpoint)
	}
	return nil
}

func QueryResourceList(rt models.ResourceTempathRel) error {
	yamlDir := beego.AppConfig.DefaultString("template::local_dir", "template")
	downLock.Lock()
	downErr, localPath := DownLoadTemplate(yamlDir, rt.ResourcePath)
	downLock.Unlock()
	if downErr != nil {
		logs.Error("File download failed, path: ", rt.ResourcePath)
		return downErr
	}
	rd := ResourceData{EnvResource: rt.ResourcePath, ResourceId: rt.ResourceId}
	content := PoolParseTmpl(yamlDir, &rd, localPath)
	// 2. Query unused instances
	var (
		objList *unstructured.UnstructuredList
		gvk     *schema.GroupVersionKind
		dr      dynamic.ResourceInterface
	)
	initRes, ok := CoursePoolVar.CourseMap[rt.ResourceId]
	if !ok {
		resCh := make(chan InitTmplResource, rt.ResPoolSize)
		CoursePoolVar.CourseMap[rt.ResourceId] = resCh
	} else {
		logs.Info("initRes: ", initRes)
	}
	err := errors.New("")
	obj := &unstructured.Unstructured{}
	_, gvk, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(content, nil, obj)
	if err != nil {
		logs.Error("failed to get GVK, err: ", err)
		return err
	}
	dr, err = GetGVRdyClient(gvk, obj.GetNamespace(), rt.ResourceId)
	if err != nil {
		logs.Error("failed to get dr: ", err)
		return err
	}
	// store db
	config := new(YamlConfig)
	err = ymV2.Unmarshal(content, config)
	if err != nil {
		logs.Error("yaml1.Unmarshal, err: ", err)
		return err
	}
	objList, err = dr.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("objList: ", objList)
	} else {
		PrintJsonList(objList)
		apiVersion := objList.GetAPIVersion()
		if config.ApiVersion == apiVersion {
			if len(objList.Items) > 0 {
				if len(objList.Items) > rt.ResPoolSize {
					RecIterList(objList.Items[:rt.ResPoolSize], obj, dr)
				} else {
					RecIterList(objList.Items, obj, dr)
				}
			}
		}
	}
	return nil
}

func CreatePoolResource(rd *ResourceData) {
	yamlDir := beego.AppConfig.DefaultString("template::local_dir", "template")
	downLock.Lock()
	downErr, localPath := DownLoadTemplate(yamlDir, rd.EnvResource)
	downLock.Unlock()
	if downErr != nil {
		logs.Error("File download failed, path: ", rd.EnvResource)
		return
	}
	content := PoolParseTmpl(yamlDir, rd, localPath)
	createErr := CreateSingleRes(content, rd)
	if createErr != nil {
		logs.Error("createErr: ", createErr)
		return
	}
}

func AddResPool(resourceId, envResource string) {
	rtr := models.ResourceTempathRel{ResourceId: resourceId, ResourcePath: envResource}
	queryErr := models.QueryResourceTempathRel(&rtr, "ResourceId", "ResourcePath")
	if queryErr != nil {
		logs.Error("queryErr: ", queryErr)
		return
	}
	rd := ResourceData{ResourceId: resourceId, EnvResource: envResource}
	courseData, ok := CoursePoolVar.CourseMap[resourceId]
	if ok {
		if len(courseData) == 0 {
			CoursePoolVar.CourseMap[resourceId] = make(chan InitTmplResource, rtr.ResPoolSize)
			for i := 0; i < rtr.ResPoolSize; i++ {
				CreatePoolResource(&rd)
			}
		} else {
			CreatePoolResource(&rd)
		}
	} else {
		CoursePoolVar.CourseMap[resourceId] = make(chan InitTmplResource, rtr.ResPoolSize)
		for i := 0; i < rtr.ResPoolSize; i++ {
			CreatePoolResource(&rd)
		}
	}
}

func InitalResPool(rtr []models.ResourceTempathRel) {
	if CoursePoolVar.InitialFlag == true {
		logs.Info("Course resource initialization completed, data: ", CoursePoolVar.InitialFlag)
		return
	}
	for _, rt := range rtr {
		queryErr := QueryResourceList(rt)
		if queryErr != nil {
			logs.Error("QueryResourceList, queryErr: ", queryErr)
		}
		rd := ResourceData{ResourceId: rt.ResourceId, EnvResource: rt.ResourcePath}
		courseId, ok := CoursePoolVar.CourseMap[rt.ResourceId]
		if !ok {
			// 1. Resource does not exist, create resource
			resCh := make(chan InitTmplResource, rt.ResPoolSize)
			CoursePoolVar.CourseMap[rt.ResourceId] = resCh
			for i := 0; i < rt.ResPoolSize; i++ {
				CreatePoolResource(&rd)
			}
		} else {
			// 1. Get existing resources
			resCount := len(courseId)
			if cap(courseId) == 0 {
				resCh := make(chan InitTmplResource, rt.ResPoolSize)
				CoursePoolVar.CourseMap[rt.ResourceId] = resCh
			}
			createResCount := rt.ResPoolSize - resCount
			if createResCount > 0 {
				for i := 0; i < createResCount; i++ {
					CreatePoolResource(&rd)
				}
			}
		}
	}
	CoursePoolVar.InitialFlag = true
}

func PrintResPool() {
	logs.Info("================Start printing resource pool data========================")
	logs.Info("Initial test completion mark: ", CoursePoolVar.InitialFlag)
	if CoursePoolVar.InitialFlag {
		for key, val := range CoursePoolVar.CourseMap {
			logs.Info("Course id:", key, ",len(val): ", len(val))
			logs.Info("val:", val)
		}
	}
	logs.Info("================End of printing resource pool data========================")
}

func InitialResourcePool() {
	// 1. Query the resource data that needs to be initialized
	rtr, num, queryErr := models.QueryResourceTempathRelAll()
	if len(rtr) == 0 {
		logs.Error("No curriculum resources need to "+
			"build initial resources, num: ", num, ",queryErr: ", queryErr)
		return
	}
	// 3. Query for available resources
	InitalResPool(rtr)
	// 4. Print resource pool data
	PrintResPool()
}

func AddInstanceToPool() error {
	InitialResourcePool()
	return nil
}
