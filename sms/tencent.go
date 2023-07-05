package sms

import (
	"context"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"log"
	"math/rand"
	"time"
)

var ErrSmsSend = fmt.Errorf("send sms fail")
var endpoint = "sms.tencentcloudapi.com"
var region = "ap-guangzhou"

type TencentSmsOption struct {
	Enable           bool   `mapstructure:"enable" json:"enable" yaml:"enable"`
	SecretId         string `mapstructure:"secret-id" json:"secret-id" yaml:"secret-id"`
	SecretKey        string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key"`
	Sign             string `mapstructure:"sign" json:"sign" yaml:"sign"`
	AppID            string `mapstructure:"app-id" json:"app-id" yaml:"app-id"`
	TemplateVeryCode string `mapstructure:"template-very-code" json:"template-very-code" yaml:"template-very-code"`
}

type TencentSmsClient struct {
	Config TencentSmsOption
}

func NewTencentSmsClient(option TencentSmsOption) *TencentSmsClient {
	return &TencentSmsClient{
		Config: option,
	}
}

func (c *TencentSmsClient) CreateCode() string {
	if !c.Config.Enable {
		return "0000"
	}
	return fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000)) //这里面前面的04v是和后面的1000相对应的
}

// Send 发送短信
func (c *TencentSmsClient) Send(ctx context.Context, phone, templateId string, args ...string) (*tencentsms.SendStatus, error) {
	if !c.Config.Enable {
		return &tencentsms.SendStatus{Code: core.String("0"), Message: core.String("服务未启用")}, nil
	}
	credential := common.NewCredential(c.Config.SecretId, c.Config.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = endpoint
	smsClient, _ := tencentsms.NewClient(credential, region, cpf)

	request := tencentsms.NewSendSmsRequest()
	request.PhoneNumberSet = common.StringPtrs([]string{"+86" + phone})
	request.SignName = common.StringPtr(c.Config.Sign)
	request.TemplateId = common.StringPtr(templateId)
	request.TemplateParamSet = common.StringPtrs(args)
	request.SmsSdkAppId = common.StringPtr(c.Config.AppID)

	response, err := smsClient.SendSms(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Println("短信[腾讯云]调用接口错误", err.Error())
		return nil, ErrSmsSend
	}
	if err != nil {
		log.Println("短信[腾讯云]调用接口错误", err.Error())
		return nil, ErrSmsSend
	}
	// 当前只有一条
	statusSet := response.Response.SendStatusSet
	code := *statusSet[0].Code
	if code == "Ok" {
		log.Println("短信[腾讯云]发信成功", response.ToJsonString())
		return statusSet[0], nil
	} else {
		log.Println("短信[腾讯云]发信失败", response.ToJsonString())
		return statusSet[0], ErrSmsSend
	}
}
