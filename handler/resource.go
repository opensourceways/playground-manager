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
//var resPoolLock sync.Mutex

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
	CourseId   string    `json:"courseId"`
	ChapterId  string    `json:"chapterId"`
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
	CourseId     string
	ChapterId    string
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

type CourseRes struct {
	CourseId    string
	ResPoolSize int
}

func DeleteFile(filePath string) {
	fileList := []string{filePath}
	common.DelFile(fileList)
}

type CourseResources struct {
	CourseId     string `yaml:"courseid"`
	ChapterId    string `yaml:"chapterid"`
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

func RetUserName(userInfo models.AuthUserInfo) (userName string) {
	if len(userInfo.Name) > 0 {
		userName = userInfo.Name
	} else if len(userInfo.NickName) > 0 {
		userName = userInfo.NickName
	} else if len(userInfo.PhoneNumber) > 0 {
		userName = userInfo.PhoneNumber
	} else {
		userName = userInfo.Email
	}
	return
}

func QueryTmpData(rtp *ReqTmplParase, rr ReqResource, cr *CourseResources, itr *InitTmplResource) {
	userInfo := models.AuthUserInfo{UserId: rr.UserId}
	userErr := models.QueryAuthUserInfo(&userInfo, "UserId")
	if userInfo.UserId == 0 {
		logs.Error("userErr:", userErr)
		return
	}
	cr.ChapterId = rr.CourseId
	cr.CourseId = rr.CourseId
	cr.LoginName = RetUserName(userInfo)
	resourceName := ResName(rr.EnvResource)
	resName := "resources-" + rr.CourseId + "-" + rr.ResourceId + "-" +
		resourceName + "-" + strconv.FormatInt(rr.UserId, 10)
	resAlias := ""
	eoi := models.ResourceInfo{ResourceName: resName}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceName")
	if eoi.Id > 0 {
		rtp.Subdomain = eoi.Subdomain
		rtp.NamePassword = fmt.Sprintf("%s:%s", eoi.UserName, eoi.PassWord)
		rtp.UserId = RetUserName(userInfo)
		rtp.Name = eoi.ResourceAlias
		rtp.ContactEmail = rr.ContactEmail
		itr.Subdomain = eoi.Subdomain
		itr.NamePassword = fmt.Sprintf("%s:%s", eoi.UserName, eoi.PassWord)
		itr.UserId = RetUserName(userInfo)
		itr.Name = eoi.ResourceAlias
		itr.ContactEmail = rr.ContactEmail
	} else {
		logs.Error("QueryTmpData, queryErr: ", queryErr)
		resAlias = "res" + rr.CourseId + "-" + rr.ResourceId + "-" + resourceName + "-" +
			strconv.FormatInt(time.Now().Unix(), 10) + common.RandomString(32)
		resAlias = "res" + common.EncryptMd5(resAlias)
		subDomain := resName + rr.EnvResource + common.RandomString(32)
		subDomain = common.EncryptMd5(subDomain)
		if ok := common.IsLetter(rune(subDomain[0])); !ok {
			subDomain = strings.Replace(subDomain, subDomain[:3], "res", 1)
		}
		userName := common.RandomString(32)
		passWord := common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(userName)))
		userName = common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(subDomain)))
		userName = userName[:16]
		namePassword := userName + ":" + passWord
		rtp.Subdomain = subDomain
		rtp.NamePassword = namePassword
		rtp.UserId = RetUserName(userInfo)
		rtp.Name = resAlias
		rtp.ContactEmail = rr.ContactEmail
		itr.Subdomain = subDomain
		itr.NamePassword = namePassword
		itr.UserId = RetUserName(userInfo)
		itr.Name = resAlias
		itr.ContactEmail = rr.ContactEmail
	}
	cr.UserId = strconv.FormatInt(rr.UserId, 10)
	cr.CourseId = rr.ResourceId
	cr.ResourceName = rr.EnvResource
}

func InitReqTmplPrarse(rtp *ReqTmplParase, rr ReqResource, cr *CourseResources, itr *InitTmplResource) {
	userInfo := models.AuthUserInfo{UserId: rr.UserId}
	userErr := models.QueryAuthUserInfo(&userInfo, "UserId")
	if userInfo.UserId == 0 {
		logs.Error("userErr:", userErr)
		return
	}
	cr.LoginName = RetUserName(userInfo)
	resourceName := ResName(rr.EnvResource)
	resName := "resources-" + rr.CourseId + "-" + rr.ResourceId + "-" +
		resourceName + "-" + strconv.FormatInt(rr.UserId, 10)
	resAlias := itr.Name
	cr.UserId = strconv.FormatInt(rr.UserId, 10)
	cr.CourseId = rr.CourseId
	cr.ChapterId = rr.ChapterId
	cr.ResourceName = rr.EnvResource
	rtp.Name = resAlias
	subDomain := itr.Subdomain
	namePassword := itr.NamePassword
	nameList := strings.Split(namePassword, ":")
	eoi := models.ResourceInfo{ResourceName: resName}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceName")
	if eoi.Id > 0 {
		rtp.Subdomain = subDomain
		rtp.NamePassword = namePassword
		rtp.UserId = RetUserName(userInfo)
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
		rtp.UserId = RetUserName(userInfo)
		eoi.ResourId = common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(userId)))
		models.InsertResourceInfo(&eoi)
	}
}

func ParseTmpl(yamlDir string, rr ReqResource, localPath string, itr *InitTmplResource, cr *CourseResources, queryFlag bool) []byte {
	if len(rr.ContactEmail) < 1 {
		rr.ContactEmail = beego.AppConfig.DefaultString("template::contact_email", "contact@openeuler.io")
	}
	rtp := ReqTmplParase{ContactEmail: rr.ContactEmail}
	if queryFlag {
		QueryTmpData(&rtp, rr, cr, itr)
	} else {
		InitReqTmplPrarse(&rtp, rr, cr, itr)
	}
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
	// Check the course duration
	courseDur := int64(0)
	if len(cr.CourseId) > 0 {
		cs := models.Courses{CourseId: cr.CourseId}
		csErr := models.QueryCourse(&cs, "CourseId")
		if csErr == nil {
			estInt, err := strconv.ParseInt(cs.Estimated, 10, 64)
			if err == nil {
				courseDur = estInt * 60
			}
		}
	}
	yamlValue := make(map[interface{}]interface{})
	met := make(map[interface{}]interface{}, 0)
	spec := make(map[interface{}]interface{}, 0)
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
		specdata, ok := yamlValue["spec"]
		if ok {
			logs.Info("specdata: ", specdata)
			spec = specdata.(map[interface{}]interface{})
		}
		if courseDur > 0 {
			spec["recycleAfterSeconds"] = courseDur
			yamlValue["spec"] = spec
		}
		yamlDt, metErr := ymV2.Marshal(yamlValue)
		if metErr != nil {
			logs.Error("metErr: ", metErr)
			return yamlData
		}
		return yamlDt
	}
	return yamlData
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
		rls.ServerErroredFlag = true
		return
	}
	name, ok := ParsingMapStr(metadata, "name")
	if !ok {
		logs.Error("name does not exist, ", name)
		rls.ServerErroredFlag = true
		return
	}
	if name != obj.GetName() {
		logs.Error("obj.GetName does not exist, ", obj.GetName())
		rls.ServerErroredFlag = true
		return
	}
	if !updateFlag {
		crs := CourseRes{}
		isSet := AddTmplResourceList(*objGetData, crs)
		if !isSet {
			rls.ServerErroredFlag = true
		}
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

func UpdateObjData(dr dynamic.ResourceInterface, cr *CourseResources, objGetData *unstructured.Unstructured,
	itr InitTmplResource, flag bool) *unstructured.Unstructured {
	err := error(nil)
	objGetData, err = dr.Get(context.TODO(), objGetData.GetName(), metav1.GetOptions{})
	if err != nil {
		logs.Error("objGetData: ", objGetData)
		return objGetData
	}
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
		rls.ServerErroredFlag = true
	} else {
		apiVersion := objGetData.GetAPIVersion()
		if config.ApiVersion == apiVersion {
			RecIter(&rls, objGetData, obj, updateFlag)
		}
	}
	logs.Info("==============Status information of the currently created resource=================\n", rls)
	return rls
}

func RecIterList(listData []unstructured.Unstructured, obj *unstructured.Unstructured,
	dr dynamic.ResourceInterface, addFlag bool, crs CourseRes) {
	containerTimeout, ok := beego.AppConfig.Int64("image::container_timeout")
	if ok != nil {
		containerTimeout = 60
	}
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
				lastTransitionTime, ok := ParsingMapStr(conds, "lastTransitionTime")
				if ok {
					rls.ServerReadyTime = lastTransitionTime
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
		deleteFlag := false
		if !rls.ServerReadyFlag {
			if len(rls.ServerReadyTime) > 1 {
				if (common.PraseTimeInt(common.GetCurTime()) -
					common.PraseTimeInt(common.TimeTConverStr(rls.ServerReadyTime))) > containerTimeout {
					logs.Error("Create image timeout is removed, resName: ", name)
					deleteFlag = true
				}
			}
		}
		if rls.ServerRecycledFlag {
			logs.Error("Images are recycled after use, resName: ", name)
			deleteFlag = true
		}
		if !deleteFlag && addFlag && !rls.ServerBoundFlag {
			ok := AddTmplResourceList(items, crs)
			if !ok {
				deleteFlag = true
			}
		}
		if deleteFlag {
			delErr := dr.Delete(context.TODO(), name, metav1.DeleteOptions{})
			if delErr != nil {
				logs.Error("delete, err: ", delErr)
			} else {
				logs.Info("Data deleted successfully, resName: ", name)
			}
		}
	}
}

func AddTmplResourceList(items unstructured.Unstructured, crs CourseRes) bool {
	metadata, ok := ParsingMap(items.Object, "metadata")
	if !ok {
		logs.Error("metadata, does not exist")
		return false
	}
	name, ok := ParsingMapStr(metadata, "name")
	if !ok || len(name) < 1 {
		logs.Error("name, does not exist")
		return false
	}
	itr := InitTmplResource{Name: name}
	annotations, ok := ParsingMap(metadata, "annotations")
	if !ok {
		logs.Error("annotations, does not exist")
		return false
	}
	courseId, ok := ParsingMapStr(annotations, "courseId")
	if !ok || len(courseId) < 1 {
		logs.Error("courseId, does not exist")
		return false
	}
	if len(crs.CourseId) < 1 {
		crs.CourseId = courseId
	}
	if crs.ResPoolSize < 1 {
		rtr := models.ResourceTempathRel{CourseId: courseId}
		quryErr := models.QueryResourceTempathRel(&rtr, "CourseId")
		if quryErr == nil {
			crs.ResPoolSize = rtr.ResPoolSize
		}
	}
	spec, ok := ParsingMap(items.Object, "spec")
	if !ok {
		logs.Error("spec, does not exist")
		return false
	}
	subdomain, ok := ParsingMapStr(spec, "subdomain")
	if !ok {
		logs.Error("subdomain, does not exist")
		return false
	}
	itr.Subdomain = subdomain
	itr.UserId = strconv.Itoa(0)
	envs, ok := ParsingMapSlice(spec, "envs")
	if !ok {
		logs.Error("envs, does not exist")
		return false
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
	if courseId == crs.CourseId {
		courseChan, ok := CoursePoolVar.Get(courseId)
		if !ok {
			courseData := make(chan InitTmplResource, crs.ResPoolSize)
			courseData <- itr
			CoursePoolVar.Set(courseId, courseData)
			logs.Info("courseId: ", courseId, "len(courseData)=", len(courseData))
		} else {
			if len(courseChan) >= crs.ResPoolSize {
				logs.Error("delete data, itr:", itr)
				return false
			}
			courseChan <- itr
			CoursePoolVar.Set(courseId, courseChan)
			logs.Info("courseId: ", courseId, "len(courseChan)=", len(courseChan))
		}
	}
	return true
}

func UpdateRes(rri *ResResourceInfo, objGetData *unstructured.Unstructured, dr dynamic.ResourceInterface,
	config *YamlConfig, obj *unstructured.Unstructured,
	objCreate *unstructured.Unstructured, cr *CourseResources, itr InitTmplResource) error {
	//PrintJsonStr(objGet)
	err := error(nil)
	curCreateTime := ""
	isDelete := false
	rls := ResListStatus{}
	containerTimeout, ok := beego.AppConfig.Int64("image::container_timeout")
	if ok != nil {
		containerTimeout = 20
	}
	for {
		rls = GetResInfo(objGetData, dr, config, obj, true)
		if rls.ServerRecycledFlag || rls.ServerErroredFlag {
			isDelete = true
			break
		}
		if len(rls.ErrorInfo) > 2 {
			logs.Error("rls.ErrorInfo: ", rls.ErrorInfo)
		}
		if !rls.ServerReadyFlag && !rls.ServerRecycledFlag {
			if len(rls.ServerReadyTime) > 1 {
				if (common.PraseTimeInt(common.GetCurTime()) -
					common.PraseTimeInt(common.TimeTConverStr(rls.ServerReadyTime))) <= containerTimeout {
					logs.Info("1.Environment is preparing...resName: ", objGetData.GetName())
					time.Sleep(time.Second)
				} else {
					isDelete = true
					break
				}
			} else {
				logs.Info("2.Environment is preparing...resName: ", objGetData.GetName())
				time.Sleep(time.Second)
			}
		}
		if rls.ServerReadyFlag {
			logs.Info("Mirror environment is ready...resName: ", objGetData.GetName())
			objGetData = UpdateObjData(dr, cr, objGetData, itr, false)
			_, err = dr.Update(context.TODO(), objGetData, metav1.UpdateOptions{})
			break
		}
	}
	rls = GetResInfo(objGetData, dr, config, obj, true)
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
		logs.Info("Created image has timed out",
			common.PraseTimeInt(common.GetCurTime())-common.PraseTimeInt(curCreateTime), config.Spec.RecycleAfterSeconds)
	}
	logs.Info("Start of updating resources, resource name:", obj.GetName())
	if isDelete {
		err = dr.Delete(context.TODO(), objGetData.GetName(), metav1.DeleteOptions{})
		if err != nil {
			logs.Error("delete, err: ", err)
		}
		return errors.New("deleted")
	}
	eoi := models.ResourceInfo{ResourceAlias: objGetData.GetName()}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceAlias")
	if eoi.Id > 0 {
		if len(curCreateTime) > 1 {
			curTime := common.PraseTimeInt(curCreateTime)
			eoi.CreateTime = curCreateTime
			eoi.RemainTime = config.Spec.RecycleAfterSeconds
			eoi.CompleteTime = config.Spec.RecycleAfterSeconds + curTime
			eoi.KindName = config.Kind
			eoi.RemainTime = config.Spec.RecycleAfterSeconds
			models.UpdateResourceInfo(&eoi, "CreateTime", "KindName", "RemainTime", "CompleteTime")
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
		courseData, ok := CoursePoolVar.Get(rr.CourseId)
		if ok {
			for {
				downLock.Lock()
				downErr, localPath := DownLoadTemplate(yamlDir, rr.EnvResource)
				downLock.Unlock()
				if downErr != nil {
					logs.Error("File download failed, path: ", rr.EnvResource)
					break
				}
				itr := <-courseData
				logs.Info("Information obtained by the resource pool: ", itr)
				itr.UserId = strconv.FormatInt(rr.UserId, 10)
				cr := CourseResources{}
				yamlData = ParseTmpl(yamlDir, rr, localPath, &itr, &cr, false)
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
					AddResPool(rr.CourseId, rr.ResourceId, rr.EnvResource)
					break
				}
				dr, err = GetGVRdyClient(gvk, obj.GetNamespace(), rr.ResourceId)
				if err != nil {
					logs.Error("failed to get dr: ", err)
					AddResPool(rr.CourseId, rr.ResourceId, rr.EnvResource)
					break
				}
				// store db
				config := new(YamlConfig)
				err = ymV2.Unmarshal(yamlData, config)
				if err != nil {
					logs.Error("yaml1.Unmarshal, err: ", err)
					AddResPool(rr.CourseId, rr.ResourceId, rr.EnvResource)
					break
				}
				objGet, err = dr.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
				if err != nil {
					logs.Error("ApplyPoolInstance, dr.Get, err: ", err)
					AddResPool(rr.CourseId, rr.ResourceId, rr.EnvResource)
					continue
				} else {
					err = UpdateRes(rri, objGet, dr, config, obj, objCreate, &cr, itr)
					if err != nil {
						logs.Error("UpdateRes err: ", err)
						AddResPool(rr.CourseId, rr.ResourceId, rr.EnvResource)
						continue
					}
					AddResPool(rr.CourseId, rr.ResourceId, rr.EnvResource)
					break
				}
			}
		} else {
			return errors.New("Instance creation failed")
		}
	} else {
		return errors.New("Instance creation failed")
	}
	return nil
}

func CreateInstance(rri *ResResourceInfo, rr ReqResource, yamlDir, localPath string,
	yamlData []byte, cr *CourseResources, itr *InitTmplResource) error {
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
		logs.Notice("Get an instance from the prepared instance, err: ", err)
		err = ApplyPoolInstance(yamlData, rri, rr, yamlDir, localPath)
		if err != nil {
			logs.Error("ApplyPoolInstance, err: ", err)
			return err
		}
	} else {
		if rr.ForceDelete == 2 {
			resName := objGet.GetName()
			err = dr.Delete(context.TODO(), resName, metav1.DeleteOptions{})
			if err != nil {
				logs.Error("delete, err: ", err)
			} else {
				logs.Info("resName: ", resName, ", Forced to delete. rr.ForceDelete: ", rr.ForceDelete)
			}
			rr.ForceDelete = 1
			err = ApplyPoolInstance(yamlData, rri, rr, yamlDir, localPath)
			if err != nil {
				logs.Error("ApplyPoolInstance, err: ", err)
				return err
			}
		} else {
			err = UpdateRes(rri, objGet, dr, config, obj, objCreate, cr, *itr)
			if err != nil {
				err = ApplyPoolInstance(yamlData, rri, rr, yamlDir, localPath)
				if err != nil {
					logs.Error("ApplyPoolInstance, err: ", err)
					return err
				}
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
	cr := CourseResources{CourseId: rr.CourseId, ChapterId: rr.ChapterId}
	yamlData := ParseTmpl(yamlDir, rr, localPath, &itr, &cr, true)
	createErr := CreateInstance(rri, rr, yamlDir, localPath, yamlData, &cr, &itr)
	if createErr != nil {
		logs.Error("createErr: ", createErr)
		return
	}
}

// Poll resource status
func GetCreateRes(yamlData []byte, rri *ResResourceInfo, resourceId string,
	cr *CourseResources, itr InitTmplResource) error {
	var (
		err       error
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
		logs.Error("err: ", err, ",resourceName: ", obj.GetName())
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
			objGet = UpdateObjData(dr, cr, objGet, itr, false)
			objUpdate, err = dr.Update(context.TODO(), objGet, metav1.UpdateOptions{})
			if err != nil {
				logs.Error("upErr: ", err, objUpdate)
			}
			rls = GetResInfo(objGet, dr, config, obj, true)
			if rls.ServerReadyFlag && rls.ServerBoundFlag {
				rri.Status = 1
				curCreateTime = common.TimeTConverStr(rls.ServerBoundTime)
				rri.EndPoint = rls.InstanceEndpoint
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
	resName := "resources-" + rr.CourseId + "-" + rr.ResourceId + "-" +
		resourceName + "-" + strconv.FormatInt(rr.UserId, 10)
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
	itr.NamePassword = fmt.Sprintf("%s:%s", ri.UserName, ri.PassWord)
	cr := CourseResources{CourseId: rr.CourseId}
	content := ParseTmpl(yamlDir, rr, localPath, &itr, &cr, true)
	GetCreateRes(content, rri, rr.ResourceId, &cr, itr)
}

func CreateUserResourceEnv(rr ReqResource) int64 {
	ure := models.UserResourceEnv{CourseId: rr.CourseId, ResourceId: rr.ResourceId,
		UserId: rr.UserId}
	queryErr := models.QueryUserResourceEnv(&ure,
		"UserId", "CourseId", "ResourceId")
	if queryErr != nil {
		logs.Error("CreateUserResourceEnv, queryErr: ", queryErr)
	}
	if ure.Id > 0 {
		ure.TemplatePath = rr.EnvResource
		ure.ChapterId = rr.ChapterId
		ure.UpdateTime = common.GetCurTime()
		ure.CourseId = rr.CourseId
		ure.UserId = rr.UserId
		ure.ResourceId = rr.ResourceId
		upErr := models.UpdateUserResourceEnv(&ure, "TemplatePath",
			"ChapterId", "UpdateTime", "CourseId", "UserId", "ResourceId")
		if upErr != nil {
			logs.Error("UpdateUserResourceEnv, upErr: ", upErr)
		}
		return ure.Id
	} else {
		ure = models.UserResourceEnv{CourseId: rr.CourseId, ChapterId: rr.ChapterId,
			ResourceId: rr.ResourceId, UserId: rr.UserId, TemplatePath: rr.EnvResource,
			ContactEmail: rr.ContactEmail, CreateTime: common.GetCurTime()}
		userResId, inertErr := models.InsertUserResourceEnv(&ure)
		if userResId == 0 {
			logs.Error("InsertUserResourceEnv, inertErr: ", inertErr)
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

func SaveResourceTemplate(rr *ReqResource) error {
	resPoolSize, ok := beego.AppConfig.Int("courses::course_pool")
	if ok != nil {
		resPoolSize = 5
	}
	rtr := models.ResourceTempathRel{ResourceId: rr.ResourceId,
		ResourcePath: rr.EnvResource, CourseId: rr.CourseId}
	tempRelErr := models.QueryResourceTempathRel(&rtr,
		"CourseId", "ResourcePath", "ResourceId")
	if rtr.Id == 0 {
		logs.Info("tempRelErr: ", tempRelErr)
		rtr.ResourceId = rr.ResourceId
		rtr.ResourcePath = rr.EnvResource
		rtr.CourseId = rr.CourseId
		rtr.ResPoolSize = resPoolSize
		rtr.CreateTime = common.GetCurTime()
		num, inErr := models.InsertResourceTempathRel(&rtr)
		if inErr != nil {
			logs.Error("inErr: ", inErr, ",num: ", num)
			return inErr
		}
	} else {
		if len(rtr.CreateTime) < 1 {
			rtr.CreateTime = common.GetCurTime()
		}
		rtr.ResourceId = rr.ResourceId
		rtr.ResourcePath = rr.EnvResource
		rtr.CourseId = rr.CourseId
		rtr.UpdateTime = common.GetCurTime()
		if rtr.ResPoolSize < 1 {
			rtr.ResPoolSize = resPoolSize
		}
		upErr := models.UpdateResourceTempathRel(&rtr,
			"ResourceId", "ResourcePath",
			"ResPoolSize", "CreateTime", "UpdateTime")
		if upErr != nil {
			logs.Error("upErr: ", upErr)
			return upErr
		}
	}
	return nil
}

func ClearInvaildResource() error {
	// 1. Data initialization
	rtr, num, err := models.QueryResourceTempathRelAll()
	if len(rtr) == 0 {
		logs.Info("err: ", err, ",num: ", num)
		return err
	}
	for _, rt := range rtr {
		yamlDir := beego.AppConfig.DefaultString("template::local_dir", "template")
		downLock.Lock()
		downErr, localPath := DownLoadTemplate(yamlDir, rt.ResourcePath)
		downLock.Unlock()
		if downErr != nil {
			logs.Error("File download failed, path: ", rt.ResourcePath)
			return downErr
		}
		rd := ResourceData{EnvResource: rt.ResourcePath,
			ResourceId: rt.ResourceId, CourseId: rt.CourseId, ResPoolSize: rt.ResPoolSize}
		content := PoolParseTmpl(yamlDir, &rd, localPath)
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
		return
	}
	crs := CourseRes{}
	apiVersion := objList.GetAPIVersion()
	if config.ApiVersion == apiVersion {
		if len(objList.Items) > 0 {
			RecIterList(objList.Items, obj, dr, false, crs)
		}
	}
}
