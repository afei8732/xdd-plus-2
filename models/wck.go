package models

import (
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
	"github.com/buger/jsonparser"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AutoGenerated struct {
	ClientVersion string `json:"clientVersion"`
	Client        string `json:"client"`
	Sv            string `json:"sv"`
	St            string `json:"st"`
	UUID          string `json:"uuid"`
	Sign          string `json:"sign"`
	FunctionID    string `json:"functionId"`
}

func getKey(WSCK string) (string, error) {

	//ptKey, _ := GetWsKey(WSCK)
	ptKey, _ := getTokenKey(WSCK)

	var count = 0
	for {
		count++
		if strings.Contains(ptKey, "app_open") || strings.Contains(ptKey, "fake") {
			return ptKey, nil
		} else {
			time.Sleep(time.Duration(rand.Int63n(10)) * time.Second)
			//ptKey, _ = GetWsKey(WSCK)
			ptKey, _ = getTokenKey(WSCK)
		}
		if count == 20 {
			return ptKey, nil
		}
	}
}

/*
个人接口  限流
*/
//
//var sign = getSign()
//
//func getSign() *AutoGenerated {
//	data, _ := httplib.Get("https://hellodns.coding.net/p/sign/d/jsign/git/raw/master/sign").SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36").Bytes()
//	t := &AutoGenerated{}
//	json.Unmarshal(data, t)
//	i := 0
//	for {
//		time.Sleep(2 * time.Second)
//		if t.Sign != "" {
//			break
//		} else if i == 5 {
//			(&JdCookie{}).Push("连续获取Sign错误请联系作者")
//			break
//		} else {
//			i++
//			data, _ = httplib.Get("https://hellodns.coding.net/p/sign/d/jsign/git/raw/master/sign").SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36").Bytes()
//			json.Unmarshal(data, t)
//		}
//	}
//	if t != nil {
//		t.FunctionID = "genToken"
//	}
//	return t
//}
//
//func getOKKey(WSCK string) (string, error) {
//	v := url.Values{}
//	s := sign
//	v.Add("functionId", s.FunctionID)
//	v.Add("clientVersion", s.ClientVersion)
//	v.Add("client", s.Client)
//	v.Add("uuid", s.UUID)
//	v.Add("st", s.St)
//	v.Add("sign", s.Sign)
//	v.Add("sv", s.Sv)
//	random := browser.Random()
//	req := httplib.Post(`https://api.m.jd.com/client.action?` + v.Encode())
//	req.Header("cookie", WSCK)
//	req.Header("User-Agent", random)
//	req.Header("content-type", `application/x-www-form-urlencoded; charset=UTF-8`)
//	req.Header("charset", `UTF-8`)
//	req.Header("accept-encoding", `br,gzip,deflate`)
//	//req.Body(`body=%7B%22to%22%3A%22https%253a%252f%252fplogin.m.jd.com%252fjd-mlogin%252fstatic%252fhtml%252fappjmp_blank.html%22%7D&`)
//	req.Body(`body=%7B%22action%22%3A%22to%22%2C%22to%22%3A%22https%253A%252F%252Fplogin.m.jd.com%252Fcgi-bin%252Fm%252Fthirdapp_auth_page%253Ftoken%253DAAEAIEijIw6wxF2s3bNKF0bmGsI8xfw6hkQT6Ui2QVP7z1Xg%2526client_type%253Dandroid%2526appid%253D879%2526appup_type%253D1%22%7D&`)
//	data, err := req.Bytes()
//	if err != nil {
//		return "", err
//	}
//	tokenKey, _ := jsonparser.GetString(data, "tokenKey")
//	logs.Info(tokenKey)
//	ptKey, _ := appjmp(tokenKey)
//	return ptKey, nil
//}
//
//func appjmp(tokenKey string) (string, error) {
//	v := url.Values{}
//	v.Add("tokenKey", tokenKey)
//	v.Add("to", ``)
//	v.Add("client_type", "android")
//	v.Add("appid", "879")
//	v.Add("appup_type", "1")
//	req := httplib.Get(`https://un.m.jd.com/cgi-bin/app/appjmp?` + v.Encode())
//	random := browser.Random()
//	req.Header("User-Agent", random)
//	req.Header("accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3`)
//	req.SetCheckRedirect(func(req *http.Request, via []*http.Request) error {
//		return http.ErrUseLastResponse
//	})
//	rsp, err := req.Response()
//	if err != nil {
//		return "", err
//	}
//	cookies := strings.Join(rsp.Header.Values("Set-Cookie"), " ")
//	//ptKey := FetchJdCookieValue("pt_key", cookies)
//	return cookies, nil
//}

/*

下面是动物园接口
*/

func getToken() string {
	post := httplib.Post("https://api.jds.codes/genToken")
	post.Header("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	bytes, _ := post.Bytes()
	logs.Info(string(bytes))
	getString, _ := jsonparser.GetString(bytes, "data", "sign")
	return getString
}

func getTokenKey(WSCK string) (string, error) {
	s := getToken()
	logs.Info(s)
	random := browser.Random()
	req := httplib.Post(`https://api.m.jd.com/client.action?` + s + "&functionId=genToken")
	req.Header("cookie", WSCK)
	req.Header("User-Agent", random)
	req.Header("content-type", `application/x-www-form-urlencoded; charset=UTF-8`)
	req.Header("charset", `UTF-8`)
	req.Header("accept-encoding", `br,gzip,deflate`)
	req.Body(`body=%7B%22to%22%3A%22https%253a%252f%252fplogin.m.jd.com%252fjd-mlogin%252fstatic%252fhtml%252fappjmp_blank.html%22%7D&`)
	data, err := req.Bytes()
	if err != nil {
		return "", err
	}
	logs.Info(string(data))
	tokenKey, _ := jsonparser.GetString(data, "tokenKey")
	ptKey, _ := appjmp(tokenKey)
	return ptKey, nil
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
	//logs.Info(cookies)
	return cookies, nil
}

/*
Zy接口注释 一下全部走ZY接口
*/

//func GetWsKey(wsKey string) (string, error) {
//	url := checkCloud()
//	return getToken(url, wsKey)
//}
//
//func checkCloud() string {
//	urlList := []string{"aHR0cDovLzQzLjEzNS45MC4yMy8=", "aHR0cHM6Ly9zaGl6dWt1Lm1sLw==", "aHR0cHM6Ly9jZi5zaGl6dWt1Lm1sLw=="}
//	for i := range urlList {
//		decodeString, _ := base64.StdEncoding.DecodeString(urlList[i])
//		_, err := httplib.Get(string(decodeString)).String()
//		if err == nil {
//			return string(decodeString)
//		}
//	}
//	return ""
//}
//
//type T struct {
//	Code      int    `json:"code"`
//	Update    int    `json:"update"`
//	Jdurl     string `json:"jdurl"`
//	UserAgent string `json:"User-Agent"`
//}
//type T2 struct {
//	FunctionId    string `json:"functionId"`
//	ClientVersion string `json:"clientVersion"`
//	Build         string `json:"build"`
//	Client        string `json:"client"`
//	Partner       string `json:"partner"`
//	Oaid          string `json:"oaid"`
//	SdkVersion    string `json:"sdkVersion"`
//	Lang          string `json:"lang"`
//	HarmonyOs     string `json:"harmonyOs"`
//	NetworkType   string `json:"networkType"`
//	Uemps         string `json:"uemps"`
//	Ext           string `json:"ext"`
//	Ef            string `json:"ef"`
//	Ep            string `json:"ep"`
//	St            int64  `json:"st"`
//	Sign          string `json:"sign"`
//	Sv            string `json:"sv"`
//}
//
//func cloudInfo(url string) string {
//	req := httplib.Get(url + "check_api")
//	req.Header("authorization", "Bearer Shizuku")
//	s, _ := req.Bytes()
//	t := T{}
//	json.Unmarshal(s, &t)
//	return t.UserAgent
//}
//
//func getToken(urls string, wskey string) (string, error) {
//	ua := cloudInfo(urls)
//	req1 := httplib.Get(urls + "genToken")
//	req1.Header("User-Agent", ua)
//	s, _ := req1.Bytes()
//	t := T2{}
//	json.Unmarshal(s, &t)
//	v := url.Values{}
//	v.Add("functionId", t.FunctionId)
//	v.Add("clientVersion", t.ClientVersion)
//	v.Add("build", t.Build)
//	v.Add("client", t.Client)
//	v.Add("partner", t.Partner)
//	v.Add("oaid", t.Oaid)
//	v.Add("sdkVersion", t.SdkVersion)
//	v.Add("lang", t.Lang)
//	v.Add("harmonyOs", t.HarmonyOs)
//	v.Add("networkType", t.NetworkType)
//	v.Add("uemps", t.Uemps)
//	v.Add("ext", t.Ext)
//	v.Add("ef", t.Ef)
//	v.Add("ep", t.Ep)
//	v.Add("st", strconv.FormatInt(t.St, 10))
//	v.Add("sign", t.Sign)
//	v.Add("sv", t.Sv)
//	req := httplib.Post(`https://api.m.jd.com/client.action?` + v.Encode())
//	req.Header("cookie", wskey)
//	req.Header("User-Agent", ua)
//	req.Header("content-type", `application/x-www-form-urlencoded; charset=UTF-8`)
//	req.Header("charset", `UTF-8`)
//	req.Header("accept-encoding", `br,gzip,deflate`)
//	req.Body(`body=%7B%22to%22%3A%22https%253a%252f%252fplogin.m.jd.com%252fjd-mlogin%252fstatic%252fhtml%252fappjmp_blank.html%22%7D&`)
//	data, err := req.Bytes()
//	if err != nil {
//		return "", err
//	}
//	tokenKey, _ := jsonparser.GetString(data, "tokenKey")
//	cookie, err := appjmp(tokenKey)
//	logs.Info(cookie)
//	if err != nil {
//		return "", err
//	}
//	return cookie, nil
//}
