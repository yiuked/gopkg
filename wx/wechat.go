// Package wx 登录逻辑
// 1.点击小程序，通过小程序授权获得openid,查看该openid是否已注册，已注册则生成token
// 2.如果未注册，获取微信手机号，并用微信手机号完成注册
// 3.如果未获取到手机号，返回一个状态码，客户端根据该状态码走自定义手机号注册流程
package wx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg/xhttp"
	"log"
	"sync"
	"time"
)

// WechatName 频道名称
const WechatName = "wechat"

const (
	WechatScopeUserInfo = "snsapi_userinfo"
)

const (
	// WechatLiteAppLoginApi GET 小程序 https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/login/auth.code2Session.html
	WechatLiteAppLoginApi = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
	// WechatAccessTokenApi GET 获取accessToken https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/access-token/auth.getAccessToken.html
	WechatAccessTokenApi = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	// WechatGetUserPhoneApi POST 获取用户手机号 https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/phonenumber/phonenumber.getPhoneNumber.html#method-http
	WechatGetUserPhoneApi = "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s"
	// WechatStableAccessTokenApi GET 获取accessToken https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getStableAccessToken.html
	WechatStableAccessTokenApi = "https://api.weixin.qq.com/cgi-bin/stable_token"
)

type LiteAppOption struct {
	Env                string `mapstructure:"env" json:"env" yaml:"env"`
	AppID              string `mapstructure:"app-id" json:"app-id" yaml:"app-id"`
	AppSecret          string `mapstructure:"app-secret" json:"app-secret" yaml:"app-secret"`
	MsgPreorderSettle  string `mapstructure:"msg-preorder-settle" json:"msg-preorder-settle" yaml:"msg-preorder-settle"`
	MsgPreorderCancel  string `mapstructure:"msg-preorder-cancel" json:"msg-preorder-cancel" yaml:"msg-preorder-cancel"`
	MsgPreorderConfirm string `mapstructure:"msg-preorder-confirm" json:"msg-preorder-confirm" yaml:"msg-preorder-confirm"`
}

// Wechat 微信结构体
type Wechat struct {
	Options LiteAppOption
	token   *AccessToken
	mu      sync.RWMutex
}

// WechatUser 微信获取到的用户信息
type WechatUser struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"` // -1 系统繁忙 0 成功 40029 code无效 45011 频率限制，每个用户每分钟100次 40226 高风险等级用户，小程序登录拦截
	ErrMsg     string `json:"errmsg"`
}

// WechatAccessToken  微信token信息
type WechatAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode"` // -1 系统繁忙 0 成功 40001 AppSecret 错误 40002 请确保 grant_type 字段值为 client_credential 40013 不合法的 AppID
	ErrMsg      string `json:"errmsg"`
}

type AccessToken struct {
	ExpireAt    time.Time
	AccessToken *WechatAccessToken
}

type GetUserPhoneNumber struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     int    `json:"countryCode"`
		Watermark       struct {
			Timestamp int    `json:"timestamp"`
			Appid     string `json:"appid"`
		} `json:"watermark"`
	} `json:"phone_info"`
}

func NewWechat(opts LiteAppOption) *Wechat {
	return &Wechat{
		Options: opts,
	}
}

// Login 登录获取 openid
// https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/login/auth.code2Session.html
func (c *Wechat) Login(ctx context.Context, code string) (*WechatUser, error) {
	resp, bytes, err := xhttp.NewClient().
		Get(fmt.Sprintf(WechatLiteAppLoginApi, c.Options.AppID, c.Options.AppSecret, code)).
		EndBytes(ctx)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("网络错误：%d", resp.StatusCode))
	}
	user := &WechatUser{}
	json.Unmarshal(bytes, user)
	if user.ErrCode != 0 {
		return nil, errors.New(fmt.Sprintf("登录失败：%d", user.ErrCode))
	}
	return user, nil
}

// GetAccessToken 获取Token,拿到token后可以用来获取手机号，注意，使用方式1时，如果有多端要用token,不适用！！！
// 方式1 https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/access-token/auth.getAccessToken.html
// 方式2 https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getStableAccessToken.html
func (c *Wechat) GetAccessToken(ctx context.Context) (*AccessToken, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.token != nil {
		log.Println("now:", time.Now().String())
		log.Println("exp:", c.token.ExpireAt.String())

		if c.token.ExpireAt.Before(time.Now()) {
			c.token = nil
		} else {
			return c.token, nil
		}
	}

	mp := make(map[string]interface{})
	mp["grant_type"] = "client_credential"
	mp["appid"] = c.Options.AppID
	mp["secret"] = c.Options.AppSecret
	resp, bytes, err := xhttp.NewClient().
		Post(WechatStableAccessTokenApi).
		SendBodyMap(mp).
		EndBytes(ctx)

	//resp, bytes, err := xhttp.NewClient().
	//	Get(fmt.Sprintf(WechatAccessTokenApi, c.Options.AppID, c.Options.AppSecret)).
	//	EndBytes(ctx)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("网络错误：%d", resp.StatusCode))
	}
	token := &WechatAccessToken{}
	json.Unmarshal(bytes, token)
	if token.ErrCode != 0 {
		return nil, errors.New(fmt.Sprintf("获取token失败：%d", token.ErrCode))
	}
	c.token = &AccessToken{
		AccessToken: token,
		ExpireAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}
	return c.token, nil
}

// GetUserPhoneNumber 获取用户手机号
// https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/phonenumber/phonenumber.getPhoneNumber.html
func (c *Wechat) GetUserPhoneNumber(ctx context.Context, code string) (*GetUserPhoneNumber, error) {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	mp := make(map[string]interface{})
	mp["code"] = code

	resp, bytes, err := xhttp.NewClient().Type(xhttp.TypeJSON).
		Post(fmt.Sprintf(WechatGetUserPhoneApi, accessToken.AccessToken.AccessToken)).
		SendBodyMap(mp).
		EndBytes(ctx)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("网络错误：%d", resp.StatusCode))
	}
	phone := &GetUserPhoneNumber{}
	json.Unmarshal(bytes, phone)
	if phone.Errcode != 0 {
		return nil, errors.New(fmt.Sprintf("获取手机号失败：%d,%s", phone.Errcode, phone.Errmsg))
	}
	return phone, nil
}
