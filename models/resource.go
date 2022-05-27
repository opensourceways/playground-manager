package models

import (
	"encoding/base64"
	"fmt"
	"playground_backend/common"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

func QueryResourceInfo(eoi *ResourceInfo, field ...string) error {
	o := orm.NewOrm()

	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertResourceInfo(eoi *ResourceInfo) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateResourceInfo(eoi *ResourceInfo, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryResourceConfigPath(eoi *ResourceConfigPath, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

func QueryUserResourceEnv(eoi *UserResourceEnv, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertUserResourceEnv(eoi *UserResourceEnv) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateUserResourceEnv(eoi *UserResourceEnv, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryResourceTempathRel(eoi *ResourceTempathRel, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertResourceTempathRel(eoi *ResourceTempathRel) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateResourceTempathRel(eoi *ResourceTempathRel, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func DeleteResourceTempathRel(eoi *ResourceTempathRel, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Delete(eoi, fields...)
	return err
}

func QueryResourceTempathRelAll() (ite []ResourceTempathRel, num int64, err error) {
	o := orm.NewOrm()
	num, err = o.Raw("select *" +
		" from pg_resource_tempath_rel").QueryRows(&ite)
	if err == nil && num > 0 {
		logs.Info("QueryResourceTempathRelAll, num: ", num)
	} else {
		logs.Error("QueryResourceTempathRelAll, err: ", err)
	}
	return
}
func MakeResourceContent() {
	resourceData := `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1EUXlOakE0TWpJMU1Gb1hEVE15TURReU16QTRNakkxTUZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTjZaCnkxclZpVm9UY21MTEZHb3hXcWxYRmdYdm5jRWtwUFdGK1lCM0N4endBM2dXaVh5UXRFVU5KKzhDSklrU0RPV3UKY2N4d2NJWmIvZG04QU1GOGZNVWJSQUZzbVhSNWl5RG5oOFpaM1FDT0drN3g1Qm9IRnowWXlNRTg3bWR5SStXVwp4RGRjekZZYmh2dVdhOWxIaWhnWjNjbmdIMjJPbk9YRGRBd0ZiRWE3VzJRQnhEVm05bkpBMzYrSWhiZ1V4NEZYCm1BRVpsaERNRERvaVFpajF4dHUwaEhHc3RvTnEvYXlmbWsrZ3dNYkVLbWIvMmNockRKSHRmL2tUL1FjYkE0R3UKOHpLK2REeERscFBIT3FLOWptZkE4eHYrMmVEUmVCU1FRS2EvdGpDaDVHR0NqdjNiYjFpSzk1eGZNaTJhczYwMgp3ekwwTDdKVXNEYUZqa2tycnprQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0tVTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFBZ3N5Nnd4OEFyT1U2OUJKMVU5bVRkaGlUWGYKSmxsZm9kWGpDS1VNQUpUU3NXUGE0ZXpObS9lTlQ4b0J4TjFHTXQrZklzQWwxR1doR0lFV2k4dERncEQzOWM3TQpUWjA5eStITm0zTE50MDdKa3JIM0NneTZheUl0ckVnRzZUNy9ZOEQ0d093eHNZbHNFKzZpbFVxTDY2M0lITzhICmFUcFFyT1ZVL0ZzRlFlVDUwWHRwdllGcTZGNWFjSkZ0cFZTVkNPN3B1eWZWcXhWUFZFbC8vU1FFdFF5SlVuczQKNW9WK2JDUU5HMUhGUG15bjdCNGNTODNQa2tYdHYvVHllVnB0WEZaUGR5UDR6ZHhxS3hQaG1Vc0NlaTJGVlllUgptcUpud1NFWVBReHJRaWpLck91SjFwVnJaRUloZW9ZN2VORlhEVE05bkxqWnZoQ1lGNmZaUjNoNWt5az0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://cls-jfymu8tt.ccs.tencent-cloud.com
  name: cls-jfymu8tt
contexts:
- context:
    cluster: cls-jfymu8tt
    user: "100011348129"
  name: cls-jfymu8tt-100011348129-context-default
current-context: cls-jfymu8tt-100011348129-context-default
kind: Config
preferences: {}
users:
- name: "100011348129"
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM5RENDQWR5Z0F3SUJBZ0lJZXNPUWJrUXhhZU13RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TWpBME1qWXdPREl5TlRCYUZ3MDBNakEwTWpFd09ETXdNVEZhTURZeApFakFRQmdOVkJBb1RDWFJyWlRwMWMyVnljekVnTUI0R0ExVUVBeE1YTVRBd01ERXhNelE0TVRJNUxURTJOVEE1Ck5qRTRNVEV3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRRG55Q2pqLzhkeXhMVGEKY3ZWN3VQOU9jVk1WZDZHaXhOdW1HeTN6WjVYVDRnOUF5T3Z4c05Ba2NrcVFkY000R0Y1aXRyVTlTTzQzVnNZaQpZeTZya0dnbmJmTnoyVGg2d2J5NS9LNzVMNnNHaDdQbHk5ZHd4OGZTT2htNUN5dFlYaUhUbzVhSldUblgvWWJKCk54VU85eHdVYnhPZHcvdWxsenliYU1ad3kxbUdLUTJDR2RxZVhMclhFYXkvMXJvZmpGaVhSN1NDK2FyZHh1ZVQKRldmR3hJQVJhVHNaY2lPZFVydDZrVE0zL2JHZENmaUFmSHJlZGFjNFpQemdnTG5kYTMwZy9nSUxJNHk5V1hXUApsSTZhSm5QdFBOSVpndjhyb0pCZlhKSm1yK0g2b2FTVkkvb0g5NXBzZ2w2R1JjQXNVTnlIdEx1ZzNQTytmRDFzCkVUZkE4eU1aQWdNQkFBR2pKekFsTUE0R0ExVWREd0VCL3dRRUF3SUZvREFUQmdOVkhTVUVEREFLQmdnckJnRUYKQlFjREFqQU5CZ2txaGtpRzl3MEJBUXNGQUFPQ0FRRUFpVWEzZzFJVUVQZXZ3L1JYejZzNm1ESXFocVUvbEUzcgpmOVFoNFZwNEMvdkx5L1c0VldMYnREcVhyU2dMb2c1aHpTUXcxYlcvc2JuNzNaNFNqWmkxOGV3dU1jTGc0Ylc4CmV4S3dPTDhlRzBDMTU0eUhjc2RIRm9ILyt1MTZpSm9CSDdEQkdGd0NjdmFMVldBQXVSNkY4RVlTWVRrWW1mcG8KckNSL3NUYXlVaWJhN003c2JKT1VSZjRLR3lWY2FlMUZkTXFuVFoxRmVJOG9ibzdwWW4xWk1vUlVBMDVUSjdUSwpHdzZFQW44bXFra29VUk9SM2x1QjlIYjNmbllGU1RjNXlLSDFESG9ha0M5S0xmQlhjcHJ3OHNVK0dMMU9wNVhmCjRSTWk0UTI1cVBiQVB4Wk9jQVdYcEJHc0dKNDgwcFhDaE5JQkJXSmF4WnErRlpQTEt1d1p0QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBNThnbzQvL0hjc1MwMm5MMWU3ai9UbkZURlhlaG9zVGJwaHN0ODJlVjArSVBRTWpyCjhiRFFKSEpLa0hYRE9CaGVZcmExUFVqdU4xYkdJbU11cTVCb0oyM3pjOWs0ZXNHOHVmeXUrUytyQm9lejVjdlgKY01mSDBqb1p1UXNyV0Y0aDA2T1dpVms1MS8yR3lUY1ZEdmNjRkc4VG5jUDdwWmM4bTJqR2NNdFpoaWtOZ2huYQpubHk2MXhHc3Y5YTZINHhZbDBlMGd2bXEzY2Jua3hWbnhzU0FFV2s3R1hJam5WSzdlcEV6Ti8yeG5RbjRnSHg2CjNuV25PR1Q4NElDNTNXdDlJUDRDQ3lPTXZWbDFqNVNPbWlaejdUelNHWUwvSzZDUVgxeVNacS9oK3FHa2xTUDYKQi9lYWJJSmVoa1hBTEZEY2g3UzdvTnp6dm53OWJCRTN3UE1qR1FJREFRQUJBb0lCQVFDbTBoQmtNVmNheW1zawpndk1lVGtNcThUT01VdU02SkplMmtQOWNuZVJmY09mcmgvRVJybUhRcVpYekxWOEpnN2lETWQ5Mk1ZU0t2ZUN6CkpWR1UxOFd2QTFyaXVhZ0Y2bVRzTStxM25OQnFUY2QwRUdlS1c2LzlKaVlYWWV1Ym1YeWRON09FUFh6OWNSTE8KeGQ5Wk93K3h3VlNNQzErV2lpaHI3bGhOSEg3ZHFGV2RORDBhcFBsVzhKS0ozUUdxMHY5V3RlZ3lFVGJFdTNMRQoyWXlYUlF4TzNKT1RaSXhBRHZYcGZxR0J4K056NVlPVC83UEYvL094VWNERDJVakhkZTZzdjhSQnRsNGJROVNTCk0zQ1F3TzZQZDd4dStySmVJWjhjcU01K0U4OUtMMWFyUmx0RHR4N2x6QWJkUXJwK0wwNUlJVUFTcWg4YWNzNE8KK2pVWkgwY2xBb0dCQVB4UGFoTDFkNEZ1K2VRdnA3SEUwSUE3ZDNTR1VzeW9hR21ER25KNG1rV1p1SWdQckY5TwpjUzZPU1NBcGVGblZ5MDE4SHdHRzhMbGNHVHZaZ2d2c3hKam9yRHNKaWQyMDI1RU9uYWNyRGhqOHVJcHQ1aFlFCmJJdW9uQ1ZYYlhnalorN2lhdVVQWmxybmJvUW5FM3k1SEdwRUNGaU5iS2JDZkN1RjFmdXlhSVdiQW9HQkFPc3IKNUhqS2ZBaFQvRG42L1duWEk5ZVJYbUZpK2MycE0rRkp1dnY4d0lpOWpuQ2VmdmcvM05MT3lXUDdHWkVCUGJSdgpvTzVadXNxUXZCWnh1T2gvR29rWGtEYUI1SW0yTCt6TDU0dkxCbjRIdlE0dm5zckowajB0aTNHbTE1TSszZlpQCjZkMXNGelRvOEs0NWhiUTZ5Z3dTVEdTMFg0YTJIbkw3NCtpMnA3OWJBb0dBS3FCQUMzUHMvTEVEQnNvR1NzSTEKZDNTVWVkczNvZHZSeUFHZU5qaXAxNWhnMUp2UlEwaTlWbUF6ZW51SEdhWkU4cEpGcXJ4aGJ1OWdVL3dyUEZpRAozbEZ3eDRpVkFoL0wrSFcvckw1WlkxOU96aFJEQ3ZVMFlXUGEvWFFIeW9Rd3l1cjFwRDAxemFYTHhnZlVBdjVECkRyRHZ1QVlzbFAvR2VwUGgwdVFSUklFQ2dZQmpQWHFFbnE0SXRhaFNyMkFSTWdDbUQycE1ubi9jRWZNYXR3cDUKSEFnRHJEcFh2QXJJcCtwLzYxT0JKWTE4YTVHbWV4VG1nR2NhNUVqN0Q3S0FLbU1BUnpsTVJ6UXlDUGZnYll1ZwpxbVJxK3NrRkc0TmZBQndBUlIvN0xmVDY1aVMwdExSMEJCRW0rc1hXUDkvMFZucTg3VnZmZzE1c2NwNFcxOFV0Cmh5YnkwUUtCZ0JFeE55N2krNnZaTU45REdzakYySTBaNkcvb2E0YzJ3Q1BLQXRNak5jYWhlUFNIQm1LdFdvN0QKbjNQcHduNlAyTXVMNU1KcmhzYWp6NlNoR0lMNHg3c2hidzNKZFB2eWsxbmQwUnMzL3NQYUhlU3Q3TGZ6aFVIdgp4Ym12TnpzSHB5V1IxbG5FUGxxeThBUGlIcFd0aXVWTDNzWXQrdHF4UG5zTHdhTFUvRS9PCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
`

	temp := common.AesString([]byte(resourceData))
	temp = base64.StdEncoding.EncodeToString([]byte(temp))
	data, baseErr := base64.StdEncoding.DecodeString(string(temp))
	if baseErr == nil {
		strContent := common.DesString(string(data))
		if resourceData == string(strContent) {
			fmt.Println("===result============:", string(temp))
		} else {
			fmt.Println("===failed============")
		}
	}

}
