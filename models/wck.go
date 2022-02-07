package models

import (
	"encoding/json"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
	"github.com/buger/jsonparser"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ua2 = `okhttp/3.12.1;jdmall;android;version/10.1.2;build/89743;screen/1440x3007;os/11;network/wifi;`

type AutoGenerated struct {
	ClientVersion string `json:"clientVersion"`
	Client        string `json:"client"`
	Sv            string `json:"sv"`
	St            string `json:"st"`
	UUID          string `json:"uuid"`
	Sign          string `json:"sign"`
	FunctionID    string `json:"functionId"`
}

var sign = getSign()
var TKey = ""

func getSign() *AutoGenerated {
	data, _ := httplib.Get("https://pan.smxy.xyz/sign").SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36").Bytes()
	t := &AutoGenerated{}
	json.Unmarshal(data, t)
	i := 0
	for {
		time.Sleep(2 * time.Second)
		if t.Sign != "" {
			break
		} else if i == 5 {
			(&JdCookie{}).Push("连续获取Sign错误请联系作者")
			break
		} else {
			i++
			data, _ = httplib.Get("https://hellodns.coding.net/p/sign/d/jsign/git/raw/master/sign").SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36").Bytes()
			json.Unmarshal(data, t)
		}
	}
	if t != nil {
		t.FunctionID = "genToken"
	}
	return t
}

func getKey(WSCK string) (string, error) {

	ptKey, _ := getOKKey(WSCK)

	var count = 0
	for {
		count++
		if strings.Contains(ptKey, "app_open") {
			return ptKey, nil
		} else {
			time.Sleep(time.Second * 20)
			TKey = ""
			ptKey, _ = getOKKey(WSCK)
		}
		if count == 4 {
			return "转换失败", nil
		}
	}
}

func getOKKey(WSCK string) (string, error) {
	if TKey == "" {
		v := url.Values{}
		//s := getSign()
		s := sign
		logs.Info(s.Sign)
		logs.Info("获取sign成功")
		v.Add("functionId", s.FunctionID)
		v.Add("clientVersion", s.ClientVersion)
		v.Add("client", s.Client)
		v.Add("uuid", s.UUID)
		v.Add("st", s.St)
		v.Add("sign", s.Sign)
		v.Add("sv", s.Sv)
		random := browser.Random()
		req := httplib.Post(`https://api.m.jd.com/client.action?` + v.Encode())
		req.Header("cookie", WSCK)
		req.Header("User-Agent", random)
		req.Header("content-type", `application/x-www-form-urlencoded; charset=UTF-8`)
		req.Header("charset", `UTF-8`)
		req.Header("accept-encoding", `br,gzip,deflate`)
		//req.Body(`body=%7B%22to%22%3A%22https%253a%252f%252fplogin.m.jd.com%252fjd-mlogin%252fstatic%252fhtml%252fappjmp_blank.html%22%7D&`)
		req.Body(`body=%7B%22action%22%3A%22to%22%2C%22to%22%3A%22https%253A%252F%252Fplogin.m.jd.com%252Fcgi-bin%252Fm%252Fthirdapp_auth_page%253Ftoken%253DAAEAIEijIw6wxF2s3bNKF0bmGsI8xfw6hkQT6Ui2QVP7z1Xg%2526client_type%253Dandroid%2526appid%253D879%2526appup_type%253D1%22%7D&`)
		data, err := req.Bytes()
		if err != nil {
			return "", err
		}
		tokenKey, _ := jsonparser.GetString(data, "tokenKey")
		if tokenKey == "" {
			logs.Info("token为空")
			sign = getSign()
		}
		logs.Info(tokenKey)
		ptKey, _ := appjmp(tokenKey)
		return ptKey, nil
	} else {
		ptKey, _ := appjmp(TKey)
		return ptKey, nil
	}

}

func appjmp(tokenKey string) (string, error) {

	v := url.Values{}
	v.Add("tokenKey", tokenKey)
	v.Add("to", `https://plogin.m.jd.com/jd-mlogin/static/html/appjmp_blank.html`)
	v.Add("client_type", "android")
	v.Add("appid", "879")
	v.Add("appup_type", "1")
	req := httplib.Get(`https://un.m.jd.com/cgi-bin/app/appjmp?` + v.Encode())
	random := browser.Random()
	req.Header("User-Agent", random)

	req.Header("accept", `accept:text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`)
	req.Header("x-requested-with", "com.jingdong.app.mall")
	req.SetCheckRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})
	rsp, err := req.Response()
	if err != nil {
		return "", err
	}
	cookies := strings.Join(rsp.Header.Values("Set-Cookie"), " ")
	//ptKey := FetchJdCookieValue("pt_key", cookies)
	logs.Info(cookies)
	return cookies, nil
}

//func getOKKey1(WSCK string) (string, error) {
//	s := getToken()
//	random := browser.Random()
//	req := httplib.Post(`https://api.m.jd.com/client.action?` + s + "&functionId=genToken")
//	req.Header("cookie", WSCK)
//	req.Header("User-Agent", random)
//	req.Header("content-type", `application/x-www-form-urlencoded; charset=UTF-8`)
//	req.Header("charset", `UTF-8`)
//	req.Header("accept-encoding", `br,gzip,deflate`)
//	req.Body(`body=%7B%22to%22%3A%22https%253a%252f%252fplogin.m.jd.com%252fjd-mlogin%252fstatic%252fhtml%252fappjmp_blank.html%22%7D&`)
//	data, err := req.Bytes()
//	if err != nil {
//		return "", err
//	}
//	logs.Info(string(data))
//	tokenKey, _ := jsonparser.GetString(data, "tokenKey")
//	ptKey, _ := appjmp(tokenKey)
//	return ptKey, nil
//}

//func getToken() string {
//	data, _ := httplib.Get("https://api.jds.codes/gentoken").SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36").Bytes()
//	getString, _ := jsonparser.GetString(data, "data", "sign")
//	return getString
//}
