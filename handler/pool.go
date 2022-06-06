package handler

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"playground_backend/common"
	"playground_backend/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	ymV2 "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
)

type ResourceData struct {
	EnvResource  string
	ResourceId   string
	ResourceName string
	CourseId     string
	ResPoolSize  int
}

type InitTmplResource struct {
	Name         string
	Subdomain    string
	NamePassword string
	UserId       string
	ContactEmail string
}

var CoursePoolVar = CoursePool{}
var PoolSync sync.RWMutex

type CoursePool struct {
	InitialFlag bool
	CourseMap   map[string]chan InitTmplResource
}

func NewCoursePool(n int) {
	CoursePoolVar = CoursePool{
		CourseMap:   make(map[string]chan InitTmplResource, n),
		InitialFlag: false,
	}
}
func (c *CoursePool) Get(key string) (chan InitTmplResource, bool) {
	PoolSync.RLock()
	defer PoolSync.RUnlock()
	v, existed := c.CourseMap[key]
	return v, existed
}

func (c *CoursePool) Set(key string, v chan InitTmplResource) {
	PoolSync.Lock()
	defer PoolSync.Unlock()
	c.CourseMap[key] = v
}

func (c *CoursePool) Delete(key string) {
	PoolSync.Lock()
	defer PoolSync.Unlock()
	delete(c.CourseMap, key)
}

func (c *CoursePool) Len() int {
	PoolSync.RLock()
	defer PoolSync.RUnlock()
	return len(c.CourseMap)
}

func (c *CoursePool) Each() {
	PoolSync.RLock()
	defer PoolSync.RUnlock()
	for key, val := range c.CourseMap {
		logs.Info("Course id:", key, ",len(val): ", len(val))
		logs.Info("val:", val)
	}
}

func InitPoolTmplPrarse(rtp *InitTmplResource, rd *ResourceData, cr *CourseResources) {
	resourceName := ResName(rd.EnvResource)
	resName := "res" + rd.CourseId + "-" + rd.ResourceId + "-" + resourceName + "-" +
		strconv.FormatInt(time.Now().Unix(), 10) + common.RandomString(32)
	rtp.UserId = "default"
	resName = "res" + common.EncryptMd5(resName)
	cr.UserId = rtp.UserId
	cr.CourseId = rd.CourseId
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
	contactEmail := beego.AppConfig.DefaultString("template::contact_email", "contact@openeuler.sh")
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
		logs.Error("---------yamlData:----", string(yamlData))
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
	coursePool, _ := CoursePoolVar.Get(rd.CourseId)
	if len(coursePool) >= rd.ResPoolSize {
		logs.Info("The current resources are sufficient and there is "+
			"no need to create new resources, len(coursePool): ", len(coursePool), ",CourseId: ", rd.CourseId)
		return errors.New("too many resources")
	}
	logs.Info("To start creating a resource, the resource name:", obj.GetName(), ",len(coursePool) = ", len(coursePool))
	objCreate, err = dr.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		logs.Error("Create err: ", err)
		return err
	}
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
	crs := CourseRes{CourseId: rt.CourseId, ResPoolSize: rt.ResPoolSize}
	rd := ResourceData{EnvResource: rt.ResourcePath, ResourceId: rt.ResourceId,
		CourseId: rt.CourseId, ResPoolSize: rt.ResPoolSize}
	content := PoolParseTmpl(yamlDir, &rd, localPath)
	// 2. Query unused instances
	var (
		objList *unstructured.UnstructuredList
		gvk     *schema.GroupVersionKind
		dr      dynamic.ResourceInterface
	)
	//initRes, ok := CoursePoolVar.CourseMap[rt.CourseId]
	initRes, ok := CoursePoolVar.Get(rt.CourseId)
	if !ok {
		resCh := make(chan InitTmplResource, rt.ResPoolSize)
		//CoursePoolVar.CourseMap[rt.CourseId] = resCh
		CoursePoolVar.Set(rt.CourseId, resCh)
	} else {
		if len(initRes) >= rt.ResPoolSize {
			logs.Info("CourseId: ", rt.CourseId, ", loading finished: ", initRes)
			return nil
		}
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
		apiVersion := objList.GetAPIVersion()
		if config.ApiVersion == apiVersion {
			if len(objList.Items) > 0 {
				RecIterList(objList.Items, obj, dr, true, crs)
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
		time.Sleep(time.Second * 1)
		return
	}
	content := PoolParseTmpl(yamlDir, rd, localPath)
	fmt.Println("--------------------------------CreatePoolResource:content:", string(content))
	createErr := CreateSingleRes(content, rd)
	if createErr != nil {
		logs.Error("createErr: -----------------", createErr)
		time.Sleep(time.Minute * 2)
		return
	}
}

func AddResPool(courseId, resourceId, envResource string) {
	rtr := models.ResourceTempathRel{CourseId: courseId, ResourceId: resourceId, ResourcePath: envResource}
	queryErr := models.QueryResourceTempathRel(&rtr, "CourseId", "ResourceId", "ResourcePath")
	if queryErr != nil {
		logs.Error("queryErr: ", queryErr)
		return
	}
	rd := ResourceData{ResourceId: resourceId, EnvResource: envResource,
		CourseId: courseId, ResPoolSize: rtr.ResPoolSize}
	CreatePoolResource(&rd)
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
		rd := ResourceData{ResourceId: rt.ResourceId,
			EnvResource: rt.ResourcePath, CourseId: rt.CourseId, ResPoolSize: rt.ResPoolSize}
		//courseId, ok := CoursePoolVar.CourseMap[rt.CourseId]
		coursePool, ok := CoursePoolVar.Get(rt.CourseId)
		fmt.Println(cap(coursePool), ok, "===========1===== CoursePoolVar.Get=====================")
		if !ok || cap(coursePool) == 0 {
			// 1. Resource does not exist, create resource
			resCh := make(chan InitTmplResource, rt.ResPoolSize)
			CoursePoolVar.Set(rt.CourseId, resCh)
			for i := 0; i < rt.ResPoolSize; i++ {
				CreatePoolResource(&rd)
			}
		} else {
			for {
				coursePool, _ = CoursePoolVar.Get(rt.CourseId)
				fmt.Println(len(coursePool), "=========2======= CoursePoolVar.Get=====================", rt.ResPoolSize)
				if len(coursePool) < rt.ResPoolSize {
					CreatePoolResource(&rd)
				} else {
					fmt.Println(len(coursePool), "=========3======= break=====================", rt.ResPoolSize)

					break
				}
			}
		}
	}
	CoursePoolVar.InitialFlag = true
}

func PrintResPool() {
	logs.Info("================Start printing resource pool data========================")
	logs.Info("Initial completion mark: ", CoursePoolVar.InitialFlag)
	CoursePoolVar.Each()
	logs.Info("================End of printing resource pool data========================")
}

func InitialResourcePool() {
	// 1. sync course
	SyncCourseData()
	// 2. Query the resource data that needs to be initialized
	rtr, num, queryErr := models.QueryResourceTempathRelAll()
	if len(rtr) == 0 {
		logs.Error("No curriculum resources need to "+
			"build initial resources, num: ", num, ",queryErr: ", queryErr)
		return
	}
	// 3. Query for available resources
	fmt.Println("----------------InitalResPool------")
	InitalResPool(rtr)
	// 4. Print resource pool data
	fmt.Println("---1-------------PrintResPool------")
	PrintResPool()
}

func ApplyCoursePool(rtr []models.ResourceTempathRel) error {
	for _, rt := range rtr {
		rd := ResourceData{ResourceId: rt.ResourceId,
			EnvResource: rt.ResourcePath, CourseId: rt.CourseId, ResPoolSize: rt.ResPoolSize}
		coursePool, ok := CoursePoolVar.Get(rt.CourseId)
		if !ok || cap(coursePool) == 0 {
			// 1. Resource does not exist, create resource
			resCh := make(chan InitTmplResource, rt.ResPoolSize)
			CoursePoolVar.Set(rt.CourseId, resCh)
			for i := 0; i < rt.ResPoolSize; i++ {
				CreatePoolResource(&rd)
			}
		} else {
			for {
				coursePool, ok = CoursePoolVar.Get(rt.CourseId)
				if ok {
					if len(coursePool) < rt.ResPoolSize {
						CreatePoolResource(&rd)
					} else {
						break
					}
				} else {
					resCh := make(chan InitTmplResource, rt.ResPoolSize)
					CoursePoolVar.Set(rt.CourseId, resCh)
					for i := 0; i < rt.ResPoolSize; i++ {
						CreatePoolResource(&rd)
					}
					break
				}
			}
		}
	}
	CoursePoolVar.InitialFlag = true
	return nil
}

func ApplyCoursePoolTask() error {
	// 1. Query the resource data that needs to be initialized
	rtr, num, queryErr := models.QueryResourceTempathRelAll()
	if len(rtr) == 0 {
		logs.Error("No curriculum resources need to "+
			"build initial resources, num: ", num, ",queryErr: ", queryErr)
		return queryErr
	}
	// 3. Query for available resources
	appErr := ApplyCoursePool(rtr)
	if appErr != nil {
		logs.Error("appErr: ", appErr)
		return appErr
	}
	// 4. Print resource pool data
	PrintResPool()
	return nil
}
