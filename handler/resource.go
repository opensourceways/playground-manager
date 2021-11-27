package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	yaml1 "gopkg.in/yaml.v2"
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
}

type ResListStatus struct {
	ServerCreatedFlag  bool
	ServerReadyFlag    bool
	ServerInactiveFlag bool
	ServerRecycledFlag bool
	ServerErroredFlag  bool
	ServerCreatedTime  string
	ServerReadyTime    string
	ServerInactiveTime string
	ServerRecycledTime string
	InstanceEndpoint   string
	ErrorInfo          string
}

func DeleteFile(filePath string) {
	fileList := []string{filePath}
	common.DelFile(fileList)
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

func InitReqTmplPrarse(rtp *ReqTmplParase, rr ReqResource, resourceId string) {
	userInfo := models.GiteeUserInfo{UserId: rr.UserId}
	userErr := models.QueryGiteeUserInfo(&userInfo, "UserId")
	if userInfo.UserId == 0 {
		logs.Error("userErr:", userErr)
		return
	}
	filenameall := path.Base(rr.EnvResource)
	filesuffix := path.Ext(rr.EnvResource)
	filePathList := strings.Split(rr.EnvResource, "/")
	pathSub := ""
	if len(filePathList) > 1 {
		pathSub = filePathList[len(filePathList)-2]
	} else {
		pathSub = common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(rr.EnvResource)))
	}
	fileprefix := filenameall[0 : len(filenameall)-len(filesuffix)]
	resName := "resources-" + resourceId + "-" + pathSub + "-" + fileprefix + "-" + strconv.FormatInt(rr.UserId, 10)
	rtp.Name = resName
	subDomain := resName + rr.EnvResource + common.RandomString(32)
	subDomain = common.EncryptMd5(subDomain)
	if ok := common.IsLetter(rune(subDomain[0])); !ok {
		subDomain = strings.Replace(subDomain, subDomain[:3], "res", 1)
	}
	eoi := models.ResourceInfo{ResourceName: resName}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceName")
	if eoi.Id > 0 {
		rtp.Subdomain = subDomain
		rtp.NamePassword = eoi.UserName + ":" + eoi.PassWord
		rtp.UserId = userInfo.UserLogin
		eoi.UpdateTime = common.GetCurTime()
		eoi.UserId = rr.UserId
		eoi.Subdomain = subDomain
		models.UpdateResourceInfo(&eoi, "UserId", "UpdateTime", "subDomain")
	} else {
		logs.Info("queryErr: ", queryErr)
		eoi.ResourceName = resName
		eoi.UserId = rr.UserId
		eoi.CreateTime = common.GetCurTime()
		eoi.CompleteTime = 0
		rtp.Subdomain = subDomain
		eoi.Subdomain = subDomain
		userName := common.RandomString(32)
		passWord := common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(userName)))
		userName = common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(subDomain)))
		userName = userName[:16]
		rtp.NamePassword = userName + ":" + passWord
		eoi.UserName = userName
		eoi.PassWord = passWord
		userId := strconv.FormatInt(rr.UserId, 10) + rr.EnvResource
		rtp.UserId = userInfo.UserLogin
		eoi.ResourId = common.EncryptMd5(base64.StdEncoding.EncodeToString([]byte(userId)))
		models.InsertResourceInfo(&eoi)
	}
}

func ParseTmpl(yamlDir string, rr ReqResource, resourceId, localPath string) []byte {
	if len(rr.ContactEmail) < 1 {
		rr.ContactEmail = beego.AppConfig.DefaultString("template::contact_email", "contact@openeuler.io")
	}
	rtp := ReqTmplParase{ContactEmail: rr.ContactEmail}
	InitReqTmplPrarse(&rtp, rr, resourceId)
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
	if common.FileExists(outPutPath) {
		DeleteFile(outPutPath)
	}
	if common.FileExists(tmpLocalPath) {
		DeleteFile(tmpLocalPath)
	}
	UnstructuredYaml(content)
	return content
}

func DownLoadTemplate(yamlDir, fPath string) (error, string) {
	common.CreateDir(yamlDir)
	fileName := path.Base(fPath)
	preFileName := common.GetRandomString(8)
	downloadUrl := beego.AppConfig.String("template::template_path")
	localPath := filepath.Join(yamlDir, preFileName+"-"+fileName)
	gitUrl := fmt.Sprintf(downloadUrl, fPath)
	logs.Info("DownLoadTemplate, gitUrl: ", gitUrl)
	resp, err := http.Get(gitUrl)
	if err != nil {
		logs.Error("DownLoadTemplate, error: ", err)
		return err, localPath
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || body == nil {
		logs.Error(err)
		return err, localPath
	}
	var contents map[string]interface{}
	err = json.Unmarshal(body, &contents)
	if err != nil {
		logs.Error(err)
		return err, localPath
	}
	if contents == nil || contents["type"] == nil {
		logs.Error("contents is nil or contents[type] is nil ", contents["type"])
		return errors.New("contents is nil"), localPath
	}
	if common.FileExists(localPath) {
		DeleteFile(localPath)
	}
	f, ferr := os.Create(localPath)
	if ferr != nil {
		logs.Error(ferr)
		return ferr, localPath
	}
	defer f.Close()
	fileType := contents["type"].(string)
	encoding := contents["encoding"].(string)
	content := contents["content"].(string)
	if fileType == "file" && encoding == "base64" {
		data, baseErr := base64.StdEncoding.DecodeString(content)
		if baseErr == nil {
			f.Write(data)
		}
	} else {
		f.WriteString(content)
	}
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

func RecIter(rls *ResListStatus, objGetData *unstructured.Unstructured, obj *unstructured.Unstructured) {
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
		}
	}
}

func GetResInfo(objGetData *unstructured.Unstructured, dr dynamic.ResourceInterface,
	config *YamlConfig, obj *unstructured.Unstructured) ResListStatus {
	err := error(nil)
	rls := ResListStatus{ServerCreatedFlag: false, ServerReadyFlag: false,
		ServerInactiveFlag: false, ServerRecycledFlag: false, ServerErroredFlag: false}
	objGetData, err = dr.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	if err != nil {
		logs.Error("objGetData: ", objGetData)
	} else {
		PrintJsonStr(objGetData)
		apiVersion := objGetData.GetAPIVersion()
		if config.ApiVersion == apiVersion {
			RecIter(&rls, objGetData, obj)
		}
	}
	logs.Info("==============Status information of the currently created resource=================\n", rls)
	return rls
}

func CreateRes(rri *ResResourceInfo, objGetData *unstructured.Unstructured, dr dynamic.ResourceInterface,
	config *YamlConfig, obj *unstructured.Unstructured, objCreate *unstructured.Unstructured) error {
	err := error(nil)
	objCreate, err = dr.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		logs.Error("Create err: ", err)
		return err
	}
	PrintJsonStr(objCreate)
	curCreateTime := ""
	rls := GetResInfo(objGetData, dr, config, obj)
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		curCreateTime = common.TimeTConverStr(rls.ServerReadyTime)
		rri.Status = 1
		rri.EndPoint = rls.InstanceEndpoint
	} else {
		rri.Status = 0
	}
	if len(rls.ErrorInfo) > 2 {
		logs.Error("err: ", rls.ErrorInfo)
	}
	eoi := models.ResourceInfo{ResourceName: config.Metadata.Name}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceName")
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
	objCreate *unstructured.Unstructured, objUpdate *unstructured.Unstructured, objGet *unstructured.Unstructured) error {
	PrintJsonStr(objGet)
	err := error(nil)
	curCreateTime := ""
	isDelete := false
	rls := GetResInfo(objGetData, dr, config, obj)
	if rls.ServerRecycledFlag {
		isDelete = true
	}
	if len(rls.ErrorInfo) > 2 {
		logs.Error("err: ", rls.ErrorInfo)
	}
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		curCreateTime = common.TimeTConverStr(rls.ServerReadyTime)
		rri.Status = 1
		rri.EndPoint = rls.InstanceEndpoint
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
	if isDelete {
		err = dr.Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{})
		if err != nil {
			logs.Error("delete, err: ", err)
			return err
		}
		objCreate, err = dr.Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil {
			logs.Error("Create err: ", err)
			return err
		}
		PrintJsonStr(objCreate)
		rri.Status = 0
	} else {
		rri.Status = 1
	}
	rls = GetResInfo(objGetData, dr, config, obj)
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		curCreateTime = common.TimeTConverStr(rls.ServerReadyTime)
		rri.Status = 1
		rri.EndPoint = rls.InstanceEndpoint
	} else {
		rri.Status = 0
	}
	if len(rls.ErrorInfo) > 2 {
		logs.Error("err: ", rls.ErrorInfo)
	}
	eoi := models.ResourceInfo{ResourceName: obj.GetName()}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceName")
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
	rri.ResName = eoi.ResourceName
	if remainTime < 0 {
		remainTime = 0
		rri.Status = 0
		rri.UserName = ""
		rri.EndPoint = ""
	}
	rri.RemainTime = remainTime
}

func StartCreateRes(yamlData []byte, rri *ResResourceInfo, resourceId string) error {
	var (
		err        error
		objGet     *unstructured.Unstructured
		objCreate  *unstructured.Unstructured
		objUpdate  *unstructured.Unstructured
		objGetData *unstructured.Unstructured
		gvk        *schema.GroupVersionKind
		dr         dynamic.ResourceInterface
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
	err = yaml1.Unmarshal(yamlData, config)
	if err != nil {
		logs.Error("yaml1.Unmarshal, err: ", err)
		return err
	}
	objGet, err = dr.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	if err != nil {
		err = CreateRes(rri, objGetData, dr, config, obj, objCreate)
		if err != nil {
			logs.Error("CreateRes err: ", err)
			return err
		}
	} else {
		err = UpdateRes(rri, objGetData, dr, config, obj, objCreate, objUpdate, objGet)
		if err != nil {
			logs.Error("UpdateRes err: ", err)
			return err
		}
	}
	return nil
}

// Create resources
func CreateEnvResource(rr ReqResource, rri *ResResourceInfo, resourceId string) {
	yamlDir := beego.AppConfig.DefaultString("template::local_dir", "template")
	downLock.Lock()
	downErr, localPath := DownLoadTemplate(yamlDir, rr.EnvResource)
	downLock.Unlock()
	if downErr != nil {
		logs.Error("File download failed, path: ", rr.EnvResource)
		return
	}
	content := ParseTmpl(yamlDir, rr, resourceId, localPath)
	StartCreateRes(content, rri, resourceId)
}

// Poll resource status
func GetCreateRes(yamlData []byte, rri *ResResourceInfo, resourceId string) error {
	var (
		err        error
		objGetData *unstructured.Unstructured
		gvk        *schema.GroupVersionKind
		dr         dynamic.ResourceInterface
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
	err = yaml1.Unmarshal(yamlData, config)
	if err != nil {
		logs.Error("yaml1.Unmarshal, err: ", err)
		return err
	}
	curCreateTime := ""
	rls := GetResInfo(objGetData, dr, config, obj)
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag {
		rri.Status = 1
		curCreateTime = common.TimeTConverStr(rls.ServerReadyTime)
		rri.EndPoint = rls.InstanceEndpoint
	} else {
		rri.Status = 0
	}
	if len(rls.ErrorInfo) > 2 {
		logs.Error("err: ", rls.ErrorInfo)
	}
	eoi := models.ResourceInfo{ResourceName: config.Metadata.Name}
	queryErr := models.QueryResourceInfo(&eoi, "ResourceName")
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

func GetEnvResource(rr ReqResource, rri *ResResourceInfo, resourceId string) {
	yamlDir := beego.AppConfig.DefaultString("template::local_dir", "template")
	downLock.Lock()
	downErr, localPath := DownLoadTemplate(yamlDir, rr.EnvResource)
	downLock.Unlock()
	if downErr != nil {
		logs.Error("File download failed, path: ", rr.EnvResource)
		return
	}
	content := ParseTmpl(yamlDir, rr, resourceId, localPath)
	GetCreateRes(content, rri, resourceId)
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
