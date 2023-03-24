package wx

import (
	"catering/common/xhttp"
	"context"
	"errors"
	"fmt"
	"log"
)

type ChatMsg struct {
	TplId    string
	Page     string
	ToUser   string
	Data     interface{}
	AppState AppState
}

const (
	// WechatMsgTplAddApi 添加模板
	WechatMsgTplAddApi = "https://api.weixin.qq.com/wxaapi/newtmpl/addtemplate?access_token=%s"
	// WechatMsgTplDelApi 删除模板
	WechatMsgTplDelApi = "https://api.weixin.qq.com/wxaapi/newtmpl/deltemplate?access_token=%s"
	// WechatMsgSendApi 发送消息
	WechatMsgSendApi = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s"
)

type AppState string

const (
	AppStateDev    AppState = "developer" // 开发版
	AppStateTrial  AppState = "trial"     // 体验版
	AppStateFormal AppState = "formal"    // 正式版本
)

var lang = "zh_CN"

// SendMsg 给客户发送信息
// https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-message-management/subscribe-message/sendMessage.html
func (c *Wechat) SendMsg(ctx context.Context, msg ChatMsg) error {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}
	mp := make(map[string]interface{})
	mp["touser"] = msg.ToUser
	mp["template_id"] = msg.TplId
	mp["page"] = msg.Page
	mp["data"] = msg.Data
	mp["miniprogram_state"] = msg.AppState
	mp["lang"] = lang

	resp, bytes, err := xhttp.NewClient().Type(xhttp.TypeJSON).
		Post(fmt.Sprintf(WechatMsgSendApi, accessToken.AccessToken.AccessToken)).
		SendBodyMap(mp).
		EndBytes(ctx)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("网络错误：%d", resp.StatusCode))
	}
	log.Println(resp)
	log.Println(string(bytes))
	return nil
}
