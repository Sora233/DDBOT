package aliyun

import (
	"encoding/json"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
)

const (
	SceneNormal = "normal"
	SceneSexy   = "sexy"
	ScenePorn   = "porn"
)

var client *sdk.Client

var qpsLimit = make(chan interface{}, 2)

func InitAliyun() {
	keyId := config.GlobalConfig.GetString("aliyun.accessKeyID")
	keySecret := config.GlobalConfig.GetString("aliyun.accessKeySecret")
	c, err := sdk.NewClientWithAccessKey("cn-shanghai", keyId, keySecret)
	if err != nil {
		panic(err)
	}
	client = c
}

func Audit(url string) (*AuditResponse, error) {
	qpsLimit <- struct{}{}
	defer func() { <-qpsLimit }()
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https"
	request.Domain = "imageaudit.cn-shanghai.aliyuncs.com"
	request.Version = "2019-12-30"
	request.ApiName = "ScanImage"
	request.QueryParams["Scene.1"] = "porn"
	request.QueryParams["Task.1.ImageURL"] = url

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		return nil, err
	}
	auditResponse := new(AuditResponse)
	if err = json.Unmarshal(response.GetHttpContentBytes(), auditResponse); err != nil {
		return nil, err
	}
	return auditResponse, nil

}
