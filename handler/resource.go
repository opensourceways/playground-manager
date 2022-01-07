package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	ymV2 "gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"playground_backend/common"
	"playground_backend/models"
	"strconv"
	"strings"
	"sync"
	"time"
)

var downLock sync.Mutex
var resPoolLock sync.Mutex

type ReqTmplParase struct {
	Name         string
	Subdomain    string
	NamePassword string
	UserId       string
	ContactEmail string
}

type YamlConfig struct {
	ApiVersion string             `yaml:"apiVersion"`
	Spec       SpecYamlConfig     `yaml:"spec"`
	Kind       string             `yaml:"kind"`
	Metadata   metadataYamlConfig `yaml:"metadata"`
}

type SpecYamlConfig struct {
	InactiveAfterSeconds int64 `yaml:"inactiveAfterSeconds"`
	RecycleAfterSeconds  int64 `yaml:"recycleAfterSeconds"`
}

type metadataYamlConfig struct {
	Name string `yaml:"name"`
}

type ResResourceInfo struct {
	UserId     int64     `json:"userId"`
	CreateTime time.Time `json:"createTime"`
	UserName   string    `json:"authInfo"`
	EndPoint   string    `json:"endPoint"`
	RemainTime int64     `json:"remainSecond"`
	ResName    string    `json:"name"`
	Status     int       `json:"status"`
	UserResId  int64     `json:"userResId"`
}

type ExcelFileInfo struct {
	RemoteFileName string
	ExcelOwner     string
	ExcelRepo      string
	AccessToken    string
	LocalDir       string
}

type ReqResource struct {
	EnvResource  string
	UserId       int64
	ContactEmail string
	ForceDelete  int
	ResourceId   string
}

type ResListStatus struct {
	ServerCreatedFlag  bool
	ServerReadyFlag    bool
	ServerInactiveFlag bool
	ServerRecycledFlag bool
	ServerErroredFlag  bool
	ServerBoundFlag    bool
	ServerCreatedTime  string
	ServerReadyTime    string
	ServerBoundTime    string
	ServerInactiveTime string
	ServerRecycledTime string
	InstanceEndpoint   string
	ErrorInfo          string
}

func DeleteFile(filePath string) {
	fileList := []string{filePath}
	common.DelFile(fileList)
}

type CourseResources struct {
	CourseId     string `yaml:"courseid"`
	ResourceName string `yaml:"resourcename"`
	UserId       string `yaml:"userid"`
	LoginName    string `yaml:"loginname"`
}

func PrintJsonStr(obj *unstructured.Unstructured) {
	logs.Info("-------------------------obj:--------------------------\n", obj)
	// encode back to JSON
	fmt.Println(".........................print unstructured.Unstructured.....................")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	enc.Encode(obj)
}

func PrintJsonList(obj *unstructured.UnstructuredList) {
	logs.Info("--------------------------obj:--------------------------\n", obj)
	// encode back to JSON
	fmt.Println("........................print unstructured.UnstructuredList.....................")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	enc.Encode(obj)
}

func GetResConfig(resourceId string) (resConfig *rest.Config, err error) {
	configPath := beego.AppConfig.String("template::local_dir")
	common.CreateDir(configPath)
	rcp := models.ResourceConfigPath{ResourceId: resourceId}
	rcpErr := models.QueryResourceConfigPath(&rcp, "ResourceId")
	if rcpErr != nil {
		logs.Error("rcpErr: ", rcpErr)
		return resConfig, err
	}
	fileName := common.EncryptMd5(rcp.ResourceContent) + ".json"
	filePath := filepath.Join(configPath, fileName)
	if common.FileExists(filePath) {
		DeleteFile(filePath)
	}
	f, ferr := os.Create(filePath)
	if ferr != nil {
		logs.Error(ferr)
		return resConfig, ferr
	}
	defer DeleteFile(filePath)
	defer f.Close()
	data, baseErr := base64.StdEncoding.DecodeString(rcp.ResourceContent)
	if baseErr == nil {
		strContent := common.DesString(string(data))
		f.Write(strContent)
	} else {
		logs.Error(baseErr)
		return resConfig, baseErr
	}
	resConfig, err = clientcmd.BuildConfigFromFlags("", filePath)
	if err != nil {
		logs.Error("BuildConfigFromFlags, err: ", err)
		return
	}
	return
}

func ResName(tplPath string) string {
	filesuffix := path.Ext(tplPath)
	tplPath = tplPath[0:(len(tplPath) - len(filesuffix))]
	pathSub := strings.ReplaceAll(tplPath, "/", "-")
	return pathSub
}

func InitReqTmplPrarse(rtp *ReqTmplParase, rr ReqResource, cr *CourseResources, itr InitTmplResource) {
	userInfo := models.GiteeUserInfo{UserId: rr.UserId}
	userErr := models.QueryGiteeUserInfo(&userInfo, "UserId")
	if userInfo.UserId == 0 {
		logs.Error("userErr:", userErr)
		return
	}
	cr.LoginName = userInfo.UserLogin
	resourceName := ResName(rr.EnvResource)
	resName := "resources-" + rr.ResourceId + "-" + resourceName + "-" + strconv.FormatInt(rr.UserId, 10)
	resAlias := ""
	if len(itr.Name) > 1 {
		resAlias = itr.Name
	} else {
		resAlias = "res" + common.EncryptMd5(resName)
	}
	cr.UserId = strconv.FormatInt(rr.UserId, 10)
	cr.CourseId = rr.ResourceId
	cr.ResourceName = rr.EnvResource
	rtp.Name = resAlias
	subDomain := ""
	if len(itr.Subdomain) > 1 {
		subDomain = itr.Subdomain
	} else {
		subDomain = resName + rr.EnvResource + common.RandomString(32)
		subDomain = common.EncryptMd5(subDomain)
		if ok := common.IsLetter(rune(subDomain[0])); !ok {
			subDomain = strings.Replace(subDomain, subDomain[:3], "res", 1)
		}
	}
	namePassword := ""
	if len(itr.NamePassword) > 1 {
		namePassword = itr.NamePassword
	} else {
		userName := common.RandomString(32)
		passWord := common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(userName)))
		userName = common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(subDomain)))
		userName = userName[:16]
		namePassword = userName + ":" + passWord
	}
	nameList := strings.Split(namePassword, ":")
	eoi := models.ResourceInfo{ResourceName: resName}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceName")
	if eoi.Id > 0 {
		rtp.Subdomain = subDomain
		rtp.NamePassword = namePassword
		rtp.UserId = userInfo.UserLogin
		eoi.UpdateTime = common.GetCurTime()
		eoi.UserId = rr.UserId
		eoi.Subdomain = subDomain
		eoi.ResourceAlias = resAlias
		eoi.UserName = nameList[0]
		eoi.PassWord = nameList[1]
		models.UpdateResourceInfo(&eoi, "UserId", "UpdateTime", "subDomain", "ResourceAlias", "UserName", "passWord")
	} else {
		logs.Info("queryErr: ", queryErr)
		eoi.ResourceName = resName
		eoi.ResourceAlias = resAlias
		eoi.UserId = rr.UserId
		eoi.CreateTime = common.GetCurTime()
		eoi.CompleteTime = 0
		rtp.Subdomain = subDomain
		eoi.Subdomain = subDomain
		rtp.NamePassword = namePassword
		eoi.UserName = nameList[0]
		eoi.PassWord = nameList[1]
		userId := strconv.FormatInt(rr.UserId, 10) + rr.EnvResource
		rtp.UserId = userInfo.UserLogin
		eoi.ResourId = common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(userId)))
		models.InsertResourceInfo(&eoi)
	}
}

func ParseTmpl(yamlDir string, rr ReqResource, localPath string, itr InitTmplResource, cr *CourseResources) []byte {
	if len(rr.ContactEmail) < 1 {
		rr.ContactEmail = beego.AppConfig.DefaultString("template::contact_email", "contact@openeuler.io")
	}
	rtp := ReqTmplParase{ContactEmail: rr.ContactEmail}
	InitReqTmplPrarse(&rtp, rr, cr, itr)
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
	if len(allFiles) == 0 {
		return []byte{}
	}
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
	content = AddAnnotations(content, cr)
	if common.FileExists(outPutPath) {
		DeleteFile(outPutPath)
	}
	if common.FileExists(tmpLocalPath) {
		DeleteFile(tmpLocalPath)
	}
	UnstructuredYaml(content)
	return content
}

func AddAnnotations(yamlData []byte, cr *CourseResources) []byte {
	yamlValue := make(map[interface{}]interface{})
	met := make(map[interface{}]interface{}, 0)
	decErr := ymV2.Unmarshal(yamlData, &yamlValue)
	if decErr != nil {
		logs.Error("decErr: ", decErr)
		return yamlData
	}
	logs.Info("yamlValue: ", yamlValue)
	if len(yamlValue) > 0 {
		resMap := make(map[interface{}]interface{})
		resMap["userId"] = cr.LoginName
		resMap["resourceName"] = cr.ResourceName
		resMap["courseId"] = cr.CourseId
		metadata, ok := yamlValue["metadata"]
		if ok {
			logs.Info("metadata: ", metadata)
			met = metadata.(map[interface{}]interface{})
		}
		met["annotations"] = resMap
		yamlValue["metadata"] = met
	}
	yamlDt, marErr := ymV2.Marshal(yamlValue)
	if marErr != nil {
		logs.Error("marErr: ", marErr)
		return yamlData
	}
	return yamlDt
}

func DownLoadTemplate(yamlDir, fPath string) (error, string) {
	common.CreateDir(yamlDir)
	fileName := path.Base(fPath)
	preFileName := common.GetRandomString(8)
	downloadUrl := beego.AppConfig.String("template::template_path")
	localPath := filepath.Join(yamlDir, preFileName+"-"+fileName)
	gitUrl := fmt.Sprintf(downloadUrl+"?file=%s", fPath)
	logs.Info("DownLoadTemplate, gitUrl: ", gitUrl)
	resp, err := http.Get(gitUrl)
	if err != nil {
		logs.Error("DownLoadTemplate, error: ", err)
		return err, localPath
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		logs.Error("resp.StatusCode: ", resp.StatusCode, ",resp.Body:", resp.Body)
		return errors.New("Template file download failed"), localPath
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || body == nil {
		logs.Error(err)
		return err, localPath
	}
	if common.FileExists(localPath) {
		DeleteFile(localPath)
	}
	f, ferr := os.Create(localPath)
	if ferr != nil {
		logs.Error(ferr)
		return ferr, localPath
	}
	f.Write(body)
	defer f.Close()
	return nil, localPath
}

func UnstructuredYaml(yamlData []byte) {
	obj := &unstructured.Unstructured{}
	// decode YAML into unstructured.Unstructured
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(yamlData, nil, obj)
	if err != nil {
		logs.Error("dec.Decode, err: ", err)
	}
	// Get the common metadata, and show GVK
	logs.Info(obj.GetName(), gvk.String())
	PrintJsonStr(obj)
}

func GetGVRdyClient(gvk *schema.GroupVersionKind, nameSpace, resourceId string) (dr dynamic.ResourceInterface, err error) {
	config, err := GetResConfig(resourceId)
	if err != nil {
		logs.Error("GetResConfig, err: ", err)
		return
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		logs.Error("NewDiscoveryClientForConfig, err: ", err)
		return
	}
	mapperGVRGVK := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))
	resourceMapper, err := mapperGVRGVK.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		logs.Error("RESTMapping, err: ", err)
		return
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logs.Error("dynamic.NewForConfig, err: ", err)
		return
	}
	if resourceMapper.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = dynamicClient.Resource(resourceMapper.Resource).Namespace(nameSpace)
	} else {
		dr = dynamicClient.Resource(resourceMapper.Resource)
	}
	return
}

func ParsingMap(mapData map[string]interface{}, key string) (map[string]interface{}, bool) {
	if value, ok := mapData[key]; ok {
		data := value.(map[string]interface{})
		return data, true
	}
	return nil, false
}

func ParsingMapStr(mapData map[string]interface{}, key string) (string, bool) {
	if value, ok := mapData[key]; ok {
		data := value.(string)
		return data, true
	}
	return "", false
}

func ParsingMapSlice(mapData map[string]interface{}, key string) ([]interface{}, bool) {
	if value, ok := mapData[key]; ok {
		data := value.([]interface{})
		return data, true
	}
	return nil, false
}

func RecIter(rls *ResListStatus, objGetData *unstructured.Unstructured,
	obj *unstructured.Unstructured, updateFlag bool) {
	metadata, ok := ParsingMap(objGetData.Object, "metadata")
	if !ok {
		logs.Error("metadata does not exist, ", metadata)
		return
	}
	name, ok := ParsingMapStr(metadata, "name")
	if !ok {
		logs.Error("name does not exist, ", name)
		return
	}
	if name != obj.GetName() {
		logs.Error("obj.GetName does not exist, ", obj.GetName())
		return
	}
	if !updateFlag {
		resPoolLock.Lock()
		AddTmplResourceList(*objGetData)
		resPoolLock.Unlock()
	}
	status, ok := ParsingMap(objGetData.Object, "status")
	if !ok {
		logs.Error("status does not exist, ", status)
		return
	}
	conditions, ok := ParsingMapSlice(status, "conditions")
	if !ok {
		logs.Error("conditions does not exist, ", conditions)
		return
	}
	for _, cond := range conditions {
		conds := cond.(map[string]interface{})
		typex, ok := ParsingMapStr(conds, "type")
		if !ok {
			continue
		}
		switch typex {
		case "ServerCreated":
			status, ok := ParsingMapStr(conds, "status")
			if ok && status == "True" {
				rls.ServerCreatedFlag = true
			}
			lastTransitionTime, ok := ParsingMapStr(conds, "lastTransitionTime")
			if ok {
				rls.ServerCreatedTime = lastTransitionTime
			}
			err, ok := ParsingMapStr(conds, "error")
			if ok {
				rls.ErrorInfo = err
			}
		case "ServerReady":
			status, ok := ParsingMapStr(conds, "status")
			if ok && status == "True" {
				rls.ServerReadyFlag = true
			}
			lastTransitionTime, ok := ParsingMapStr(conds, "lastTransitionTime")
			if ok {
				rls.ServerReadyTime = lastTransitionTime
			}
			err, ok := ParsingMapStr(conds, "error")
			if ok {
				rls.ErrorInfo = err
			}
			message, ok := ParsingMap(conds, "message")
			if ok {
				instanceEndpoint, ok := ParsingMapStr(message, "instanceEndpoint")
				if ok {
					rls.InstanceEndpoint = instanceEndpoint
				}
			}
		case "ServerInactive":
			status, ok := ParsingMapStr(conds, "status")
			if ok && status == "True" {
				rls.ServerInactiveFlag = true
			}
			lastTransitionTime, ok := ParsingMapStr(conds, "lastTransitionTime")
			if ok {
				rls.ServerInactiveTime = lastTransitionTime
			}
		case "ServerRecycled":
			status, ok := ParsingMapStr(conds, "status")
			if ok && status == "True" {
				rls.ServerRecycledFlag = true
			}
			lastTransitionTime, ok := ParsingMapStr(conds, "lastTransitionTime")
			if ok {
				rls.ServerRecycledTime = lastTransitionTime
			}
		case "ServerErrored":
			status, ok := ParsingMapStr(conds, "status")
			if ok && status == "True" {
				rls.ServerErroredFlag = true
			}
			message, ok := ParsingMap(conds, "message")
			if ok {
				detail, ok := ParsingMapStr(message, "detail")
				if ok {
					rls.ErrorInfo = detail
				}
			}
		case "ServerBound":
			status, ok := ParsingMapStr(conds, "status")
			if ok && status == "True" {
				rls.ServerBoundFlag = true
			}
			lastTransitionTime, ok := ParsingMapStr(conds, "lastTransitionTime")
			if ok {
				rls.ServerBoundTime = lastTransitionTime
			}
		}
	}
}

func UpdateObjData(cr *CourseResources, objGetData *unstructured.Unstructured, itr InitTmplResource, flag bool) *unstructured.Unstructured {
	metadata, ok := ParsingMap(objGetData.Object, "metadata")
	if !ok {
		logs.Error("metadata does not exist, ", metadata)
		return objGetData
	}
	name, ok := ParsingMapStr(metadata, "name")
	if !ok {
		logs.Error("name does not exist, ", name)
		return objGetData
	}
	if len(cr.ResourceName) > 1 {
		annotations, ok := ParsingMap(metadata, "annotations")
		if !ok {
			annotMap := make(map[string]interface{}, 0)
			annotMap["courseId"] = cr.CourseId
			annotMap["resourceName"] = cr.ResourceName
			annotMap["userId"] = cr.LoginName
			metadata["annotations"] = annotMap
		} else {
			annotations["courseId"] = cr.CourseId
			annotations["resourceName"] = cr.ResourceName
			annotations["userId"] = cr.LoginName
			metadata["annotations"] = annotations
		}
	}
	spec, ok := ParsingMap(objGetData.Object, "spec")
	if !ok {
		logs.Error("spec, does not exist")
		return objGetData
	}
	if len(itr.Subdomain) > 1 {
		spec["subdomain"] = itr.Subdomain
	}
	envs, ok := ParsingMapSlice(spec, "envs")
	if !ok {
		logs.Error("envs, does not exist")
		return objGetData
	}
	tmpEnv := make([]interface{}, 0)
	for _, ev := range envs {
		ev := ev.(map[string]interface{})
		evName, ok := ParsingMapStr(ev, "name")
		if !ok {
			continue
		}
		switch evName {
		case "GOTTY_CREDENTIAL":
			if len(itr.NamePassword) > 1 {
				ev["value"] = itr.NamePassword
			}
		case "COMMUNITY_EMAIL":
			if len(itr.ContactEmail) > 1 {
				ev["value"] = itr.ContactEmail
			}
		case "SHELL_USER":
			if len(cr.LoginName) > 1 {
				ev["value"] = cr.LoginName
			}
		}
		tmpEnv = append(tmpEnv, ev)
	}
	spec["envs"] = tmpEnv
	objGetData.Object["spec"] = spec
	if !flag {
		status, ok := ParsingMap(objGetData.Object, "status")
		if !ok {
			logs.Error("status does not exist, ", status)
			return objGetData
		}
		conditions, ok := ParsingMapSlice(status, "conditions")
		if !ok {
			logs.Error("conditions does not exist, ", conditions)
			return objGetData
		}
		tmpCondition := make([]interface{}, 0)
		for _, cond := range conditions {
			conds := cond.(map[string]interface{})
			typex, ok := ParsingMapStr(conds, "type")
			if !ok {
				continue
			}
			switch typex {
			case "ServerBound":
				status, ok := ParsingMapStr(conds, "status")
				if !ok || status == "False" {
					conds["lastTransitionTime"] = common.GetTZHTime(8)
				}
				conds["lastUpdateTime"] = common.GetTZHTime(8)
				conds["reason"] = fmt.Sprintf("code server has been bound")
				conds["status"] = "True"

			}
			tmpCondition = append(tmpCondition, conds)
			status["conditions"] = tmpCondition
			objGetData.Object["status"] = status
		}
	}
	return objGetData
}

func GetResInfo(objGetData *unstructured.Unstructured, dr dynamic.ResourceInterface,
	config *YamlConfig, obj *unstructured.Unstructured, updateFlag bool) ResListStatus {
	err := error(nil)
	rls := ResListStatus{ServerCreatedFlag: false, ServerReadyFlag: false,
		ServerInactiveFlag: false, ServerRecycledFlag: false, ServerErroredFlag: false}
	objGetData, err = dr.Get(context.TODO(), objGetData.GetName(), metav1.GetOptions{})
	if err != nil {
		logs.Error("objGetData: ", objGetData)
	} else {
		PrintJsonStr(objGetData)
		apiVersion := objGetData.GetAPIVersion()
		if config.ApiVersion == apiVersion {
			RecIter(&rls, objGetData, obj, updateFlag)
		}
	}
	logs.Info("==============Status information of the currently created resource=================\n", rls)
	return rls
}

func RecIterList(listData []unstructured.Unstructured, obj *unstructured.Unstructured,
	dr dynamic.ResourceInterface) {
	for _, items := range listData {
		metadata, ok := ParsingMap(items.Object, "metadata")
		if !ok {
			continue
		}
		name, ok := ParsingMapStr(metadata, "name")
		if !ok {
			continue
		}
		status, ok := ParsingMap(items.Object, "status")
		if !ok {
			continue
		}
		conditions, ok := ParsingMapSlice(status, "conditions")
		if !ok {
			continue
		}
		rls := ResListStatus{ServerCreatedFlag: false, ServerReadyFlag: false,
			ServerInactiveFlag: false, ServerRecycledFlag: false}
		for _, cond := range conditions {
			conds := cond.(map[string]interface{})
			typex, ok := ParsingMapStr(conds, "type")
			if !ok {
				continue
			}
			switch typex {
			case "ServerCreated":
				status, ok := ParsingMapStr(conds, "status")
				if ok && status == "True" {
					rls.ServerCreatedFlag = true
				}
			case "ServerReady":
				status, ok := ParsingMapStr(conds, "status")
				if ok && status == "True" {
					rls.ServerReadyFlag = true
				}
			case "ServerInactive":
				status, ok := ParsingMapStr(conds, "status")
				if ok && status == "True" {
					rls.ServerInactiveFlag = true
				}
			case "ServerRecycled":
				status, ok := ParsingMapStr(conds, "status")
				if ok && status == "True" {
					rls.ServerRecycledFlag = true
				}
			case "ServerBound":
				status, ok := ParsingMapStr(conds, "status")
				if ok && status == "True" {
					rls.ServerBoundFlag = true
				}
			}
		}
		if !rls.ServerReadyFlag && rls.ServerRecycledFlag {
			delErr := dr.Delete(context.TODO(), name, metav1.DeleteOptions{})
			if delErr != nil {
				logs.Error("delete, err: ", delErr)
			} else {
				logs.Info("Data deleted successfully, name: ", name)
			}
		}
		if !rls.ServerBoundFlag && rls.ServerReadyFlag {
			resPoolLock.Lock()
			AddTmplResourceList(items)
			resPoolLock.Unlock()
		}
	}
}

func AddTmplResourceList(items unstructured.Unstructured) {
	metadata, ok := ParsingMap(items.Object, "metadata")
	if !ok {
		logs.Error("metadata, does not exist")
		return
	}
	name, ok := ParsingMapStr(metadata, "name")
	if !ok {
		logs.Error("name, does not exist")
		return
	}
	itr := InitTmplResource{Name: name}
	annotations, ok := ParsingMap(metadata, "annotations")
	if !ok {
		logs.Error("annotations, does not exist")
		return
	}
	courseId, ok := ParsingMapStr(annotations, "courseId")
	if !ok {
		logs.Error("courseId, does not exist")
		return
	}
	spec, ok := ParsingMap(items.Object, "spec")
	if !ok {
		logs.Error("spec, does not exist")
		return
	}
	subdomain, ok := ParsingMapStr(spec, "subdomain")
	if !ok {
		logs.Error("subdomain, does not exist")
		return
	}
	itr.Subdomain = subdomain
	itr.UserId = strconv.Itoa(0)
	envs, ok := ParsingMapSlice(spec, "envs")
	if !ok {
		logs.Error("envs, does not exist")
		return
	}
	for _, ev := range envs {
		ev := ev.(map[string]interface{})
		evName, ok := ParsingMapStr(ev, "name")
		if !ok {
			continue
		}
		switch evName {
		case "GOTTY_CREDENTIAL":
			value, ok := ParsingMapStr(ev, "value")
			if ok && len(value) > 0 {
				itr.NamePassword = value
			}
		case "COMMUNITY_EMAIL":
			value, ok := ParsingMapStr(ev, "value")
			if ok && len(value) > 0 {
				itr.ContactEmail = value
			}
		}
	}
	course, ok := CoursePoolVar.CourseMap[courseId]
	if !ok {
		courseData := make(chan InitTmplResource, 0)
		courseData <- itr
		CoursePoolVar.CourseMap[courseId] = courseData
	} else {
		CoursePoolVar.CourseMap[courseId] <- itr
		logs.Info("len(course)=", len(course))
	}
}

func CreateRes(rri *ResResourceInfo, dr dynamic.ResourceInterface,
	config *YamlConfig, obj *unstructured.Unstructured,
	objCreate *unstructured.Unstructured, cr *CourseResources, itr InitTmplResource) error {
	err := error(nil)
	logs.Info("To start creating a resource, the resource name:", obj.GetName())
	objCreate, err = dr.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		logs.Error("Create err: ", err)
		return err
	}
	//PrintJsonStr(objCreate)
	curCreateTime := ""
	rls := GetResInfo(objCreate, dr, config, obj, true)
	rri.Status = 0
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		if rls.ServerBoundFlag {
			curCreateTime = common.TimeTConverStr(rls.ServerBoundTime)
			rri.Status = 1
		}
		rri.EndPoint = rls.InstanceEndpoint
		if !rls.ServerBoundFlag {
			objCreate = UpdateObjData(cr, objCreate, itr, false)
			_, err = dr.UpdateStatus(context.TODO(), objCreate, metav1.UpdateOptions{})
		}
	}
	if len(rls.ErrorInfo) > 2 {
		logs.Error("err: ", rls.ErrorInfo)
	}
	eoi := models.ResourceInfo{ResourceAlias: config.Metadata.Name}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceAlias")
	if eoi.Id > 0 {
		if len(curCreateTime) > 1 {
			curTime := common.PraseTimeInt(curCreateTime)
			eoi.CreateTime = curCreateTime
			eoi.CompleteTime = curTime + config.Spec.RecycleAfterSeconds
		}
		eoi.KindName = config.Kind
		eoi.RemainTime = config.Spec.RecycleAfterSeconds
		models.UpdateResourceInfo(&eoi, "CreateTime", "KindName", "RemainTime", "CompleteTime")
		ParaseResData(objCreate, rri, eoi)
	} else {
		logs.Error("queryErr: ", queryErr)
		return queryErr
	}
	return nil
}

func UpdateRes(rri *ResResourceInfo, objGetData *unstructured.Unstructured, dr dynamic.ResourceInterface,
	config *YamlConfig, obj *unstructured.Unstructured,
	objCreate *unstructured.Unstructured, cr *CourseResources, itr InitTmplResource) error {
	//PrintJsonStr(objGet)
	err := error(nil)
	curCreateTime := ""
	isDelete := false
	rls := ResListStatus{}
	for {
		objGetData = UpdateObjData(cr, objGetData, itr, false)
		_, err = dr.Update(context.TODO(), objGetData, metav1.UpdateOptions{})
		rls = GetResInfo(objGetData, dr, config, obj, true)
		if rls.ServerRecycledFlag {
			isDelete = true
			break
		}
		if len(rls.ErrorInfo) > 2 {
			logs.Error("rls.ErrorInfo: ", rls.ErrorInfo)
		}
		if (!rls.ServerReadyFlag && !rls.ServerRecycledFlag) || rls.ServerReadyFlag {
			break
		}
	}
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		if rls.ServerBoundFlag {
			curCreateTime = common.TimeTConverStr(rls.ServerBoundTime)
			rri.Status = 1
			rri.EndPoint = rls.InstanceEndpoint
		} else {
			rri.Status = 0
		}
	} else if !rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		rri.Status = 0
	} else {
		isDelete = true
	}
	if (common.PraseTimeInt(common.GetCurTime()) - common.PraseTimeInt(curCreateTime)) > config.Spec.RecycleAfterSeconds {
		isDelete = true
		rri.Status = 0
		logs.Info(common.PraseTimeInt(common.GetCurTime())-common.PraseTimeInt(curCreateTime), config.Spec.RecycleAfterSeconds)
	}
	logs.Info("Start of updating resources, resource name:", obj.GetName())
	if isDelete {
		err = dr.Delete(context.TODO(), objGetData.GetName(), metav1.DeleteOptions{})
		if err != nil {
			logs.Error("delete, err: ", err)
			return err
		}
		objCreate, err = dr.Create(context.TODO(), objGetData, metav1.CreateOptions{})
		if err != nil {
			logs.Error("Create err: ", err)
			return err
		}
		//PrintJsonStr(objCreate)
		rls = GetResInfo(objCreate, dr, config, obj, true)
		rri.Status = 0
		if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
			if rls.ServerBoundFlag {
				curCreateTime = common.TimeTConverStr(rls.ServerBoundTime)
				rri.Status = 1
				rri.EndPoint = rls.InstanceEndpoint
			}
		}
		if len(rls.ErrorInfo) > 2 {
			logs.Error("err: ", rls.ErrorInfo)
		}
	} else {
		rri.Status = 1
	}
	eoi := models.ResourceInfo{ResourceAlias: obj.GetName()}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceAlias")
	if eoi.Id > 0 {
		if len(curCreateTime) > 1 {
			curTime := common.PraseTimeInt(curCreateTime)
			remainTime := eoi.CompleteTime - curTime
			if remainTime < 0 {
				eoi.CreateTime = curCreateTime
				eoi.RemainTime = config.Spec.RecycleAfterSeconds
				eoi.CompleteTime = config.Spec.RecycleAfterSeconds + curTime
				models.UpdateResourceInfo(&eoi, "CreateTime", "RemainTime", "CompleteTime")
			}
		}
		ParaseResData(obj, rri, eoi)
	} else {
		logs.Error("queryErr: ", queryErr)
		return queryErr
	}
	return nil
}

func ForceCreateRes(rri *ResResourceInfo, objUpdate *unstructured.Unstructured, dr dynamic.ResourceInterface,
	config *YamlConfig, obj *unstructured.Unstructured,
	objCreate *unstructured.Unstructured, cr *CourseResources, itr InitTmplResource) error {
	err := error(nil)
	curCreateTime := ""
	logs.Info("Forced deletion of resources begins, resource name:", obj.GetName())
	err = dr.Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		logs.Error("delete, err: ", err)
		return err
	}
	logs.Info("Forcibly delete the resource, create the resource again, the resource name:", obj.GetName())
	objCreate, err = dr.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		logs.Error("Create err: ", err)
		return err
	}
	//PrintJsonStr(objCreate)
	rri.Status = 0
	rls := GetResInfo(objCreate, dr, config, obj, true)
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		if rls.ServerBoundFlag {
			curCreateTime = common.TimeTConverStr(rls.ServerBoundTime)
			rri.Status = 1
			rri.EndPoint = rls.InstanceEndpoint
		}
		if !rls.ServerBoundFlag {
			objCreate = UpdateObjData(cr, objCreate, itr, false)
			objUpdate, err = dr.UpdateStatus(context.TODO(), objCreate, metav1.UpdateOptions{})
		}
	}
	if len(rls.ErrorInfo) > 2 {
		logs.Error("err: ", rls.ErrorInfo)
	}
	eoi := models.ResourceInfo{ResourceAlias: obj.GetName()}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceAlias")
	if eoi.Id > 0 {
		if len(curCreateTime) > 1 {
			curTime := common.PraseTimeInt(curCreateTime)
			eoi.CreateTime = curCreateTime
			eoi.RemainTime = config.Spec.RecycleAfterSeconds
			eoi.CompleteTime = config.Spec.RecycleAfterSeconds + curTime
			models.UpdateResourceInfo(&eoi, "CreateTime", "RemainTime", "CompleteTime")
		}
		ParaseResData(obj, rri, eoi)
	} else {
		logs.Error("queryErr: ", queryErr)
		return queryErr
	}
	return nil
}

func ParaseResData(resData *unstructured.Unstructured, rri *ResResourceInfo, eoi models.ResourceInfo) {
	if len(resData.Object) < 1 {
		logs.Error("resData is nil")
		return
	}
	curTime := common.PraseTimeInt(common.GetCurTime())
	rri.CreateTime = common.LocalTimeToUTC(eoi.CreateTime)
	rri.UserId = eoi.UserId
	remainTime := eoi.CompleteTime - curTime
	rri.UserName = eoi.UserName + ":" + eoi.PassWord
	rri.ResName = eoi.ResourceAlias
	if remainTime < 0 {
		remainTime = 0
		rri.Status = 0
		rri.UserName = ""
		rri.EndPoint = ""
	}
	rri.RemainTime = remainTime
}

func ApplyPoolInstance(yamlData []byte, rri *ResResourceInfo, rr ReqResource, yamlDir, localPath string) error {
	if CoursePoolVar.InitialFlag {
		courseData, ok := CoursePoolVar.CourseMap[rr.ResourceId]
		if ok {
			if len(courseData) > 0 {
				// 1. Allocate unused resources
				itr, ok := <-courseData
				itr.UserId = strconv.FormatInt(rr.UserId, 10)
				downLock.Lock()
				downErr, localPath := DownLoadTemplate(yamlDir, rr.EnvResource)
				downLock.Unlock()
				if downErr != nil {
					logs.Error("File download failed, path: ", rr.EnvResource)
					return downErr
				}
				cr := CourseResources{}
				yamlData = ParseTmpl(yamlDir, rr, localPath, itr, &cr)
				var (
					err       error
					objGet    *unstructured.Unstructured
					objCreate *unstructured.Unstructured
					gvk       *schema.GroupVersionKind
					dr        dynamic.ResourceInterface
				)
				rri.Status = 0
				obj := &unstructured.Unstructured{}
				_, gvk, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(yamlData, nil, obj)
				if err != nil {
					logs.Error("failed to get GVK, err: ", err)
					return err
				}
				dr, err = GetGVRdyClient(gvk, obj.GetNamespace(), rr.ResourceId)
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
				objGet, err = dr.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
				if err != nil {
					logs.Error("resName:", obj.GetName(), "err: ", err)
					err = CreateRes(rri, dr, config, obj, objCreate, &cr, itr)
					if err != nil {
						logs.Error("CreateRes, err: ", err)
						return err
					}
					return nil
				} else {
					if rr.ForceDelete == 2 {
						err = ForceCreateRes(rri, objGet, dr, config, obj, objCreate, &cr, itr)
						if err != nil {
							logs.Error("UpdateRes err: ", err)
							return err
						}
						return nil
					} else {
						err = UpdateRes(rri, objGet, dr, config, obj, objCreate, &cr, itr)
						if err != nil {
							logs.Error("UpdateRes err: ", err)
							return err
						}
						if ok {
							AddResPool(rr.ResourceId, rr.EnvResource)
						}
						return nil
					}
				}
			} else {
				return errors.New("Instance creation failed")
			}
		} else {
			return errors.New("Instance creation failed")
		}
	}
	return nil
}

func CreateInstance(rri *ResResourceInfo, rr ReqResource, yamlDir, localPath string, yamlData []byte, cr *CourseResources) error {
	var (
		err       error
		objGet    *unstructured.Unstructured
		objCreate *unstructured.Unstructured
		gvk       *schema.GroupVersionKind
		dr        dynamic.ResourceInterface
	)
	rri.Status = 0
	itr := InitTmplResource{}
	obj := &unstructured.Unstructured{}
	_, gvk, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(yamlData, nil, obj)
	if err != nil {
		logs.Error("failed to get GVK, err: ", err)
		return err
	}
	dr, err = GetGVRdyClient(gvk, obj.GetNamespace(), rr.ResourceId)
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
	objGet, err = dr.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	if err != nil {
		logs.Notice("Get an instance from the prepared instance, err: ", err)
		err = ApplyPoolInstance(yamlData, rri, rr, yamlDir, localPath)
		if err != nil {
			logs.Error("UpdateInstance, err: ", err)
			err = CreateRes(rri, dr, config, obj, objCreate, cr, itr)
			if err != nil {
				logs.Error("CreateRes, err: ", err)
				return err
			}
		}
	} else {
		if rr.ForceDelete == 2 {
			err = ForceCreateRes(rri, objGet, dr, config, obj, objCreate, cr, itr)
			if err != nil {
				logs.Error("UpdateRes err: ", err)
				return err
			}
		} else {
			err = UpdateRes(rri, objGet, dr, config, obj, objCreate, cr, itr)
			if err != nil {
				logs.Error("UpdateRes err: ", err)
				return err
			}
		}
	}
	return nil
}

// Create resources
func CreateEnvResource(rr ReqResource, rri *ResResourceInfo) {
	yamlDir := beego.AppConfig.DefaultString("template::local_dir", "template")
	downLock.Lock()
	downErr, localPath := DownLoadTemplate(yamlDir, rr.EnvResource)
	downLock.Unlock()
	if downErr != nil {
		logs.Error("File download failed, path: ", rr.EnvResource)
		return
	}
	itr := InitTmplResource{}
	cr := CourseResources{}
	yamlData := ParseTmpl(yamlDir, rr, localPath, itr, &cr)
	createErr := CreateInstance(rri, rr, yamlDir, localPath, yamlData, &cr)
	if createErr != nil {
		logs.Error("createErr: ", createErr)
		return
	}
}

// Poll resource status
func GetCreateRes(yamlData []byte, rri *ResResourceInfo, resourceId string, cr *CourseResources, itr InitTmplResource) error {
	var (
		err error
		gvk       *schema.GroupVersionKind
		dr        dynamic.ResourceInterface
		objUpdate *unstructured.Unstructured
		objGet    *unstructured.Unstructured
	)
	rri.Status = 0
	obj := &unstructured.Unstructured{}
	_, gvk, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(yamlData, nil, obj)
	if err != nil {
		logs.Error("failed to get GVK, err: ", err)
		return err
	}
	dr, err = GetGVRdyClient(gvk, obj.GetNamespace(), resourceId)
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
	curCreateTime := ""
	objGet, err = dr.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	if err != nil {
		logs.Error("err: ", err, ",resourceName: " ,obj.GetName())
		return err
	}
	rls := GetResInfo(objGet, dr, config, obj, true)
	rri.Status = 0
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		if rls.ServerBoundFlag {
			rri.Status = 1
			curCreateTime = common.TimeTConverStr(rls.ServerBoundTime)
			rri.EndPoint = rls.InstanceEndpoint
		}
		if !rls.ServerBoundFlag {
			objGet = UpdateObjData(cr, objGet, itr, false)
			objUpdate, err = dr.Update(context.TODO(), objGet, metav1.UpdateOptions{})
			if err != nil {
				logs.Error("upErr: ", err, objUpdate)
			}
		}
	}
	if len(rls.ErrorInfo) > 2 {
		logs.Error("ErrorInfo: ", rls.ErrorInfo)
	}
	eoi := models.ResourceInfo{ResourceAlias: config.Metadata.Name}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceAlias")
	if eoi.Id > 0 {
		if len(curCreateTime) > 1 {
			curTime := common.PraseTimeInt(curCreateTime)
			eoi.CreateTime = curCreateTime
			eoi.CompleteTime = curTime + config.Spec.RecycleAfterSeconds
		}
		eoi.KindName = config.Kind
		eoi.RemainTime = config.Spec.RecycleAfterSeconds
		models.UpdateResourceInfo(&eoi, "CreateTime", "KindName", "RemainTime", "CompleteTime")
		ParaseResData(obj, rri, eoi)
	} else {
		logs.Error("queryErr: ", queryErr)
		return queryErr
	}
	return nil
}

func GetEnvResource(rr ReqResource, rri *ResResourceInfo) {
	yamlDir := beego.AppConfig.DefaultString("template::local_dir", "template")
	downLock.Lock()
	downErr, localPath := DownLoadTemplate(yamlDir, rr.EnvResource)
	downLock.Unlock()
	if downErr != nil {
		logs.Error("File download failed, path: ", rr.EnvResource)
		return
	}
	itr := InitTmplResource{}
	resourceName := ResName(rr.EnvResource)
	resName := "resources-" + rr.ResourceId + "-" + resourceName + "-" + strconv.FormatInt(rr.UserId, 10)
	ri := models.ResourceInfo{ResourceName: resName}
	queryErr := models.QueryResourceInfo(&ri, "ResourceName")
	if queryErr != nil {
		logs.Error("queryErr: ", queryErr)
		return
	}
	itr.Name = ri.ResourceAlias
	itr.Subdomain = ri.Subdomain
	itr.ContactEmail = rr.ContactEmail
	itr.UserId = strconv.FormatInt(rr.UserId, 10)
	cr := CourseResources{}
	content := ParseTmpl(yamlDir, rr, localPath, itr, &cr)
	GetCreateRes(content, rri, rr.ResourceId, &cr, itr)
}

func CreateUserResourceEnv(rr ReqResource, resourceId string) int64 {
	ure := models.UserResourceEnv{ResourceId: resourceId, UserId: rr.UserId, TemplatePath: rr.EnvResource}
	queryErr := models.QueryUserResourceEnv(&ure, "UserId", "ResourceId", "TemplatePath")
	if queryErr != nil {
		logs.Error("CreateUserResourceEnv, queryErr: ", queryErr)
	}
	if ure.Id > 0 {
		return ure.Id
	} else {
		ure = models.UserResourceEnv{ResourceId: resourceId, UserId: rr.UserId, TemplatePath: rr.EnvResource,
			ContactEmail: rr.ContactEmail, CreateTime: common.GetCurTime()}
		userResId, inertErr := models.InsertUserResourceEnv(&ure)
		if userResId == 0 {
			logs.Error("CreateUserResourceEnv, inertErr: ", inertErr)
		}
		return userResId
	}
}

func QueryUserResourceEnv(ure *models.UserResourceEnv) {
	queryErr := models.QueryUserResourceEnv(ure, "Id")
	if queryErr != nil {
		logs.Error("QueryUserResourceEnv, queryErr: ", queryErr)
	}
}

func SaveResourceTemplate(rr ReqResource) {
	rtr := models.ResourceTempathRel{ResourceId: rr.ResourceId, ResourcePath: rr.EnvResource}
	tempRelErr := models.QueryResourceTempathRel(&rtr, "ResourceId", "ResourcePath")
	if rtr.Id == 0 {
		logs.Info("tempRelErr: ", tempRelErr)
		rtr.ResourceId = rr.ResourceId
		rtr.ResourcePath = rr.EnvResource
		num, inErr := models.InsertResourceTempathRel(&rtr)
		if inErr != nil {
			logs.Error("inErr: ", inErr, ",num: ", num)
		}
	}
}

func ClearInvaildResource() error {
	// 1. Data initialization
	rtr, num, err := models.QueryResourceTempathRelAll()
	if len(rtr) == 0 {
		logs.Info("err: ", err, ",num: ", num)
		return err
	}
	for _, rt := range rtr {
		rr := ReqResource{EnvResource: rt.ResourcePath, UserId: 1000000000,
			ContactEmail: "", ForceDelete: 1, ResourceId: rt.ResourceId}
		yamlDir := beego.AppConfig.DefaultString("template::local_dir", "template")
		downLock.Lock()
		downErr, localPath := DownLoadTemplate(yamlDir, rr.EnvResource)
		downLock.Unlock()
		if downErr != nil {
			logs.Error("File download failed, path: ", rr.EnvResource)
			return downErr
		}
		itr := InitTmplResource{}
		cr := CourseResources{}
		content := ParseTmpl(yamlDir, rr, localPath, itr, &cr)
		// 2. Query unused instances
		var (
			objList *unstructured.UnstructuredList
			gvk     *schema.GroupVersionKind
			dr      dynamic.ResourceInterface
		)
		obj := &unstructured.Unstructured{}
		_, gvk, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(content, nil, obj)
		if err != nil {
			logs.Error("failed to get GVK, err: ", err)
			return err
		}
		dr, err = GetGVRdyClient(gvk, obj.GetNamespace(), rr.ResourceId)
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
		DelInvaildResource(objList, dr, config, obj)
	}
	return nil
}

func DelInvaildResource(objList *unstructured.UnstructuredList, dr dynamic.ResourceInterface,
	config *YamlConfig, obj *unstructured.Unstructured) {
	err := error(nil)
	objList, err = dr.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("objList: ", objList)
	} else {
		PrintJsonList(objList)
		apiVersion := objList.GetAPIVersion()
		if config.ApiVersion == apiVersion {
			if len(objList.Items) > 0 {
				RecIterList(objList.Items, obj, dr)
			}
		}
	}
}
