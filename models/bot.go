package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/buger/jsonparser"
	"github.com/skip2/go-qrcode"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
)

var SendQQ = func(a int64, b interface{}) {

}
var SendQQGroup = func(a int64, b int64, c interface{}) {

}

type ArkResData struct {
	Status uint `json:"status"`
}

type ArkRes struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    ArkResData `json:"data"`
}

type ViVoData struct {
	Autologin  int    `json:"autologin"`
	Gsalt      string `json:"gsalt"`
	GUID       string `json:"guid"`
	Lsid       string `json:"lsid"`
	NeedAuth   int    `json:"need_auth"`
	ReturnPage string `json:"return_page"`
	RsaModulus string `json:"rsa_modulus"`
}

type ViVoRes struct {
	Data    ViVoData `json:"data"`
	ErrCode int      `json:"err_code"`
	ErrMsg  string   `json:"err_msg"`
}

var ListenQQPrivateMessage = func(uid int64, msg string) {
	SendQQ(uid, handleMessage(msg, "qq", int(uid)))
}

var ListenQQTempPrivateMessage = func(uid int64, msg string) {
	SendQQ(uid, handleMessage(msg, "qq", int(uid)))
}

var ListenQQGroupMessage = func(gid int64, uid int64, msg string) {
	if gid == Config.QQGroupID {
		if Config.QbotPublicMode {
			SendQQGroup(gid, uid, handleMessage(msg, "qqg", int(uid), int(gid)))
		} else {
			SendQQ(uid, handleMessage(msg, "qq", int(uid)))
		}
	}
}

var pcodes = make(map[int]string)
var replies = map[string]string{}
var riskcodes = make(map[int]string)
var riskcodes1 = make(map[string]ViVoData)
var tytlist = make(map[string]int)
var tytno = 0
var tytnum = 0

func InitReplies() {
	f, err := os.Open(ExecPath + "/conf/reply.php")
	if err == nil {
		defer f.Close()
		data, _ := ioutil.ReadAll(f)
		ss := regexp.MustCompile("`([^`]+)`\\s*=>\\s*`([^`]+)`").FindAllStringSubmatch(string(data), -1)
		for _, s := range ss {
			replies[s[1]] = s[2]
		}
	}
	if _, ok := replies["壁纸"]; !ok {
		replies["壁纸"] = "https://acg.toubiec.cn/random.php"
	}
}

var handleMessage = func(msgs ...interface{}) interface{} {
	time.Sleep(time.Second * time.Duration(rand.Intn(5)))
	msg := msgs[0].(string)
	args := strings.Split(msg, " ")
	head := args[0]
	contents := args[1:]
	sender := &Sender{
		UserID:   msgs[2].(int),
		Type:     msgs[1].(string),
		Contents: contents,
	}
	if len(msgs) >= 4 {
		sender.ChatID = msgs[3].(int)
	}
	if sender.Type == "tgg" {
		sender.MessageID = msgs[4].(int)
		sender.Username = msgs[5].(string)
		sender.ReplySenderUserID = msgs[6].(int)
	}
	if sender.UserID == Config.TelegramUserID || sender.UserID == int(Config.QQID) {
		sender.IsAdmin = true
	}
	if sender.IsAdmin == false {
		if IsUserAdmin(strconv.Itoa(sender.UserID)) {
			sender.IsAdmin = true
		}
	}
	for i := range codeSignals {
		for j := range codeSignals[i].Command {
			if codeSignals[i].Command[j] == head {
				return func() interface{} {
					if codeSignals[i].Admin && !sender.IsAdmin {
						return "你没有权限操作"
					}
					return codeSignals[i].Handle(sender)
				}()
			}
		}
	}
	if Config.VIP {
		switch msg {
		default:

			//校验卡密
			{
				if strings.HasPrefix(msg, "XDD") {
					return useKey(msg, sender.UserID)
				}
			}

			//{
			//	if sender.IsAdmin {
			//		if strings.Contains(msg, "膨胀") {
			//			rsp := httplib.Post("http://jd.zack.xin/api/jd/ulink.php")
			//			rsp.Param("url", msg)
			//			rsp.Param("type", "hy")
			//			data, err := rsp.Response()
			//
			//			if err != nil {
			//				return "口令转换失败"
			//			}
			//			body, _ := ioutil.ReadAll(data.Body)
			//			if strings.Contains(string(body), "口令转换失败") {
			//				return "口令转换失败"
			//			} else {
			//				if strings.Contains(string(body), "shareType=expandHelp") {
			//					inviterCode := regexp.MustCompile(`inviteId=(\S+)(&|&amp;)mpin`).FindStringSubmatch(string(body))
			//					k, flag := startpz(inviterCode[1])
			//					if flag {
			//						return fmt.Sprintf("助力完成，一共助力%d账号", k)
			//					} else {
			//						return fmt.Sprintf("助力失败，一共助力%d账号", k)
			//					}
			//				}
			//			}
			//		}
			//	}
			//}

			//挖宝统计
			//{
			//	if strings.Contains(msg, "https://bnzf.jd.com/") {
			//		f, err := os.OpenFile(ExecPath+"/wblj.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
			//		if err != nil {
			//			logs.Warn("wb.txt失败，", err)
			//		}
			//		if GetCoin(sender.UserID) > 25 {
			//			f.WriteString(msg + "\n")
			//			RemCoin(sender.UserID, 25)
			//			sender.Reply(fmt.Sprintf("已提交转订单，扣除积分25，剩余积分：%d", GetCoin(sender.UserID)))
			//		} else {
			//			sender.Reply("积分不足")
			//		}
			//		f.Close()
			//	}
			//}
			//转码
			{
				if strings.Contains(msg, "https://kpl.m.jd.com/product") {
					ss := regexp.MustCompile(`wareId=(\S+)(&|&amp;)utm_source`).FindStringSubmatch(msg)
					url := fmt.Sprintf("https://wqdeal.jd.com/deal/confirmorder/main?commlist=%s,,1,%s,1,0,0", ss[1], ss[1])
					data, _ := qrcode.Encode(url, qrcode.Medium, 256)
					err2 := ioutil.WriteFile("./output.jpg", data, 0666)
					if err2 != nil {
						logs.Error(err2)
					}
					return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
				} else if strings.Contains(msg, "item.m.jd.com/product") {
					var s = msg[strings.Index(msg, "product/")+8 : strings.Index(msg, ".html?")]
					logs.Info(s)
					url := fmt.Sprintf("https://wqdeal.jd.com/deal/confirmorder/main?commlist=%s,,1,%s,1,0,0", s, s)
					data, _ := qrcode.Encode(url, qrcode.Medium, 256)
					err2 := ioutil.WriteFile("./output.jpg", data, 0666)
					if err2 != nil {
						logs.Error(err2)
					}
					return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
				}
			}

			{
				if strings.Contains(msg, "口令") {
					rsp := httplib.Post("http://jd.zack.xin/api/jd/ulink.php")
					rsp.Param("url", msg)
					rsp.Param("type", "hy")
					//rsp.Body(fmt.Sprintf(`url=%s&type=hy`, msg))
					data, err := rsp.Response()

					if err != nil {
						return "口令转换失败"
					}
					body, _ := ioutil.ReadAll(data.Body)
					if strings.Contains(string(body), "口令转换失败") {
						return "口令转换失败"
					} else {
						return string(body)
					}
				}
			}

			//验证码
			{
				regex := "^\\d{5}(\\d|X|x)$"
				reg := regexp.MustCompile(regex)
				if reg.MatchString(msg) {
					logs.Info("进入验证码阶段")
					var addr string
					if len(Config.Jdcurl) > 0 {
						addr = Config.Jdcurl
					} else if len(Config.Madurl) > 0 {
						addr = Config.Madurl
					}
					phone := pcodes[sender.UserID]
					if len(addr) > 0 {
						//若兰登录

						risk := riskcodes[sender.UserID]
						logs.Info(sender.UserID)
						if strings.EqualFold(risk, "true") {
							logs.Info("进入风险验证阶段")
							if phone != "" {
								req := httplib.Post(addr + "/api/VerifyCardCode")
								req.Header("content-type", "application/json")
								data, _ := req.Body(`{"Phone":"` + phone + `","QQ":"` + strconv.Itoa(sender.UserID) + `","qlkey":0,"Code":"` + msg + `"}`).Bytes()
								var arkRes ArkRes
								json.Unmarshal(data, &arkRes)
								if arkRes.Success || strings.Contains(arkRes.Message, "添加xdd成功") {
									sender.Reply("登录成功。可以继续登录下一个账号")
									go func() {
										Save <- &JdCookie{}
									}()
								} else if !arkRes.Success {
									sender.Reply("验证失败,可能填写错误")
								}
							}
							riskcodes[sender.UserID] = "false"
						} else {
							logs.Info("进入验证码阶段")
							if phone != "" {
								req := httplib.Post(addr + "/api/VerifyCode")
								req.Header("content-type", "application/json")
								data, _ := req.Body(`{"Phone":"` + phone + `","QQ":"` + strconv.Itoa(sender.UserID) + `","qlkey":0,"Code":"` + msg + `"}`).Bytes()
								var arkRes ArkRes
								json.Unmarshal(data, &arkRes)
								if arkRes.Data.Status == 555 {
									//验证
									sender.Reply("你的账号需要验证才能登陆，请输入你的京东账号绑定的身份证前两位和后四位，最后一位如果是X，请输入大写X\n例如：31122X")
									//做个标记
									riskcodes[sender.UserID] = "true"
									if arkRes.Message != "" {
										sender.Reply(arkRes.Message)
									}
								} else if strings.Contains(arkRes.Message, "添加xdd成功") {
									sender.Reply("登录成功。可以继续登录下一个账号")
									go func() {
										Save <- &JdCookie{}
									}()
								} else {
									if arkRes.Message != "" {
										sender.Reply(arkRes.Message)
									} else {
										sender.Reply("登陆失败，请重新登录，多次尝试失败请联系管理员")
									}
								}
							}
						}
					}
				}
			}

			//手机号
			{
				ist := pcodes[(sender.UserID)]
				if strings.EqualFold(ist, "true") {
					regular := `^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\d{8}$`
					reg := regexp.MustCompile(regular)
					if reg.MatchString(msg) {
						//诺兰登录
						if len(Config.Jdcurl) > 0 {
							sender.Reply("请耐心等待...")
							addr := Config.Jdcurl
							req := httplib.Post(addr + "/api/SendSMS")
							req.Header("content-type", "application/json")
							data, _ := req.Body(`{"Phone":"` + msg + `","qlkey":0}`).Bytes()
							message, _ := jsonparser.GetString(data, "message")
							success, _ := jsonparser.GetBoolean(data, "success")
							status, _ := jsonparser.GetInt(data, "data", "status")
							captcha, _ := jsonparser.GetInt(data, "data", "captcha")
							if captcha == 0 {
								captcha = 1
							}
							if message != "" && status != 666 {
								sender.Reply(message)
							}
							i := 1

							if success {
								pcodes[sender.UserID] = msg
								logs.Info(string(sender.UserID))
								sender.Reply("请输入6位验证码：")
								break
							}
							//{"success":true,"message":"","data":{"ckcount":0,"tabcount":3}}
							if !success && status == 666 && i < 5 && captcha == 2 {

								sender.Reply("正在进行验证...")
								for {
									req = httplib.Post(addr + "/api/AutoCaptcha")
									req.Header("content-type", "application/json")
									data, _ := req.Body(`{"Phone":"` + msg + `"}`).Bytes()
									message, _ := jsonparser.GetString(data, "message")
									success, _ := jsonparser.GetBoolean(data, "success")
									status, _ := jsonparser.GetInt(data, "data", "status")
									if !success {
										//s.Reply("滑块验证失败：" + string(data))
									}
									if success {
										pcodes[sender.UserID] = msg
										sender.Reply("请输入6位验证码：")
										break
									}
									if i > 5 {
										//pcodes[sender.UserID] = msg
										//s := Config.Jdcurl + "/Captcha/" + msg
										//sender.Reply(fmt.Sprintf("请访问网址进行手动验证%s", s))
										sender.Reply("滑块验证失败,请尝试重新登录")
										break
									}
									if status == 666 {
										i++
										sender.Reply(fmt.Sprintf("正在进行第%d次滑块验证...", i))
										continue
									}
									if strings.Contains(message, "上限") {
										i = 6
										sender.Reply(message)
									}
									//sender.Reply(message)
								}
								//} else if !success && captcha == 2 {
								//	pcodes[string(sender.UserID)] = msg
								//	s := Config.Jdcurl + "/Captcha/" + msg
								//	sender.Reply(fmt.Sprintf("请访问网址进行手动验证%s", s))

							} else {

								sender.Reply("滑块失败，请网页登录")
							}
							//{"success":true,"message":"","data":{"ckcount":0,"tabcount":3}}
						} else if len(Config.Madurl) > 0 {
							sender.Reply("请耐心等待...")
							addr := Config.Madurl
							req := httplib.Post(addr + "/api/SendSMS")
							req.Header("content-type", "application/json")
							data, _ := req.Body(`{"Phone":"` + msg + `","qlkey":0}`).Bytes()
							message, _ := jsonparser.GetString(data, "message")
							success, _ := jsonparser.GetBoolean(data, "success")
							status, _ := jsonparser.GetInt(data, "data", "status")
							if message != "" && status != 666 {
								sender.Reply(message)
							}
							i := 1

							if success {
								pcodes[sender.UserID] = msg
								logs.Info(string(sender.UserID))
								sender.Reply("请输入6位验证码：")
								break
							}
							//{"success":true,"message":"","data":{"ckcount":0,"tabcount":3}}
							if !success && status == 666 && i < 5 {

								sender.Reply("正在进行验证...")
								for {
									req = httplib.Post(addr + "/api/AutoCaptcha")
									req.Header("content-type", "application/json")
									data, _ := req.Body(`{"Phone":"` + msg + `"}`).Bytes()
									message, _ := jsonparser.GetString(data, "message")
									success, _ := jsonparser.GetBoolean(data, "success")
									status, _ := jsonparser.GetInt(data, "data", "status")
									if !success {
										//s.Reply("滑块验证失败：" + string(data))
									}
									if success {
										pcodes[sender.UserID] = msg
										sender.Reply("请输入6位验证码：")
										break
									}
									if i > 5 {
										//pcodes[sender.UserID] = msg
										//s := Config.Jdcurl + "/Captcha/" + msg
										//sender.Reply(fmt.Sprintf("请访问网址进行手动验证%s", s))
										sender.Reply("滑块验证失败,请尝试重新登录")
										break
									}
									if status == 666 {
										i++
										sender.Reply(fmt.Sprintf("正在进行第%d次滑块验证...", i))
										continue
									}
									if strings.Contains(message, "上限") {
										i = 6
										sender.Reply(message)
									}
								}
							} else {
								sender.Reply("滑块失败，请网页登录")
							}
						}
					}
				}
			}

			//识别登录
			{
				if strings.Contains(msg, "登录") || strings.Contains(msg, "登陆") {
					var tabcount string
					var addr string
					if len(Config.Jdcurl) > 0 {
						addr = Config.Jdcurl
					} else if len(Config.Madurl) > 0 {
						addr = Config.Madurl
					}
					if addr == "" {
						return "暂未对接机器人登录"
					}
					logs.Info(addr + "/api/Config")
					if addr != "" {
						data, _ := httplib.Get(addr + "/api/Config").Bytes()
						logs.Info(string(data) + "返回数据")
						tabcount, _ = jsonparser.GetString(data, "data", "autocount")
						if tabcount != "0" {
							pcodes[sender.UserID] = "true"
							riskcodes[sender.UserID] = "false"
							if len(Config.Jdcurl) > 0 {
								sender.Reply("若兰为您服务，请输入11位手机号：")
							} else if len(Config.Madurl) > 0 {
								sender.Reply("疯兔为您服务，请输入11位手机号：")
							}
						} else {
							sender.Reply("服务忙，请稍后再试。")
						}
					}

					//sender.Reply("服务升级中，目前登录请私聊群主谢谢")
				}
			}

		}
	}

	switch msg {
	default:

		{ //沃邮箱
			ss := regexp.MustCompile(`https://nyan.mail.*3D`).FindStringSubmatch(msg)
			if len(ss) > 0 {
				var u User
				if db.Where("number = ?", sender.UserID).First(&u).Error != nil {
					return 0
				}
				db.Model(u).Updates(map[string]interface{}{
					"womail": ss[0],
				})
				sender.Reply(fmt.Sprintf("沃邮箱提交成功!"))
				return nil
			}
		}

		{
			if strings.Contains(msg, "wskey=") {
				logs.Info(msg + "开始WSKEY登录")
				wsKey := FetchJdCookieValue("wskey", msg)
				ptPin := FetchJdCookieValue("pin", msg)
				if len(ptPin) == 0 {
					ptPin = FetchJdCookieValue("pt_pin", msg)
				}
				if len(wsKey) > 0 && len(ptPin) > 0 {
					wkey := "pin=" + ptPin + ";wskey=" + wsKey + ";"
					rsp, err := getKey(wkey)
					if err != nil {
						logs.Error(err)
					}
					if strings.Contains(rsp, "fake_") {
						logs.Error("wskey错误")
						sender.Reply(fmt.Sprintf("wskey错误 除京东APP皆不可用"))
					} else {
						ptKey := FetchJdCookieValue("pt_key", rsp)
						ptPin := FetchJdCookieValue("pt_pin", rsp)
						ck := JdCookie{
							PtPin: ptPin,
							PtKey: ptKey,
							WsKey: wsKey,
						}
						if CookieOK(&ck) {

							if sender.IsQQ() {
								ck.QQ = sender.UserID
							} else if sender.IsTG() {
								ck.Telegram = sender.UserID
							}
							if nck, err := GetJdCookie(ck.PtPin); err == nil {
								nck.InPool(ck.PtKey)
								if nck.WsKey == "" || len(nck.WsKey) == 0 {
									if sender.IsQQ() {
										ck.Update(QQ, ck.QQ)
									}
									nck.Update(WsKey, ck.WsKey)
									msg := fmt.Sprintf("写入WsKey，并更新账号%s", ck.PtPin)
									sender.Reply(fmt.Sprintf(msg))
									(&JdCookie{}).Push(msg)
									logs.Info(msg)
								} else {
									if nck.WsKey == ck.WsKey {
										msg := fmt.Sprintf("重复写入")
										sender.Reply(fmt.Sprintf(msg))
										(&JdCookie{}).Push(msg)
										logs.Info(msg)
									} else {
										nck.Updates(JdCookie{
											WsKey: ck.WsKey,
										})
										msg := fmt.Sprintf("更新WsKey，并更新账号%s", ck.PtPin)
										sender.Reply(fmt.Sprintf(msg))
										(&JdCookie{}).Push(msg)
										logs.Info(msg)
									}
								}

							} else {
								NewJdCookie(&ck)

								msg := fmt.Sprintf("添加账号，账号名:%s", ck.PtPin)

								if sender.IsQQ() {
									ck.Update(QQ, ck.QQ)
								}

								sender.Reply(fmt.Sprintf(msg))
								sender.Reply(ck.Query())
								(&JdCookie{}).Push(msg)
							}
						}
						go func() {
							Save <- &JdCookie{}
						}()
						return nil
					}
				}
			}
			//ss := regexp.MustCompile(`pin=([^;=\s]+);wskey=([^;=\s]+)`).FindAllStringSubmatch(msg, -1)
			//if len(ss) > 0 {
			//	for _, s := range ss {
			//		wkey := "pin=" + s[1] + ";wskey=" + s[2] + ";"
			//		//rsp := cmd(fmt.Sprintf(`python3 test.py "%s"`, wkey), &Sender{})
			//		rsp, err := getKey(wkey)
			//		if err != nil {
			//			logs.Error(err)
			//		}
			//		if strings.Contains(rsp, "fake_") {
			//			logs.Error("wskey错误")
			//			sender.Reply(fmt.Sprintf("wskey错误 除京东APP皆不可用"))
			//		} else {
			//			ptKey := FetchJdCookieValue("pt_key", rsp)
			//			ptPin := FetchJdCookieValue("pt_pin", rsp)
			//			ck := JdCookie{
			//				PtPin: ptPin,
			//				PtKey: ptKey,
			//				WsKey: s[2],
			//			}
			//			if CookieOK(&ck) {
			//
			//				if sender.IsQQ() {
			//					ck.QQ = sender.UserID
			//				} else if sender.IsTG() {
			//					ck.Telegram = sender.UserID
			//				}
			//				if nck, err := GetJdCookie(ck.PtPin); err == nil {
			//					nck.InPool(ck.PtKey)
			//					if nck.WsKey == "" || len(nck.WsKey) == 0 {
			//						if sender.IsQQ() {
			//							ck.Update(QQ, ck.QQ)
			//						}
			//						nck.Update(WsKey, ck.WsKey)
			//						msg := fmt.Sprintf("写入WsKey，并更新账号%s", ck.PtPin)
			//						sender.Reply(fmt.Sprintf(msg))
			//						(&JdCookie{}).Push(msg)
			//						logs.Info(msg)
			//					} else {
			//						if nck.WsKey == ck.WsKey {
			//							msg := fmt.Sprintf("重复写入")
			//							sender.Reply(fmt.Sprintf(msg))
			//							(&JdCookie{}).Push(msg)
			//							logs.Info(msg)
			//						} else {
			//							nck.Updates(JdCookie{
			//								WsKey: ck.WsKey,
			//							})
			//							msg := fmt.Sprintf("更新WsKey，并更新账号%s", ck.PtPin)
			//							sender.Reply(fmt.Sprintf(msg))
			//							(&JdCookie{}).Push(msg)
			//							logs.Info(msg)
			//						}
			//					}
			//
			//				} else {
			//					NewJdCookie(&ck)
			//
			//					msg := fmt.Sprintf("添加账号，账号名:%s", ck.PtPin)
			//
			//					if sender.IsQQ() {
			//						ck.Update(QQ, ck.QQ)
			//					}
			//
			//					sender.Reply(fmt.Sprintf(msg))
			//					sender.Reply(ck.Query())
			//					(&JdCookie{}).Push(msg)
			//				}
			//			}
			//			go func() {
			//				Save <- &JdCookie{}
			//			}()
			//			return nil
			//		}
			//	}
			//}
		}

		//{ //
		//	ss := regexp.MustCompile(`pt_key=([^;=\s]+);pt_pin=([^;=\s]+)`).FindAllStringSubmatch(msg, -1)
		//
		//	if len(ss) > 0 {
		//
		//		xyb := 0
		//		for _, s := range ss {
		//			ck := JdCookie{
		//				PtKey: s[1],
		//				PtPin: s[2],
		//			}
		//			xyb++
		//			if sender.IsQQ() {
		//				ck.QQ = sender.UserID
		//			} else if sender.IsTG() {
		//				ck.Telegram = sender.UserID
		//			}
		//			if HasKey(ck.PtKey) {
		//				sender.Reply(fmt.Sprintf("重复提交"))
		//			} else {
		//				if nck, err := GetJdCookie(ck.PtPin); err == nil {
		//					nck.InPool(ck.PtKey)
		//					msg := fmt.Sprintf("更新账号，%s", ck.PtPin)
		//					(&JdCookie{}).Push(msg)
		//					logs.Info(msg)
		//				} else {
		//					if Cdle {
		//						ck.Hack = True
		//					}
		//					NewJdCookie(&ck)
		//					msg := fmt.Sprintf("添加账号，%s", ck.PtPin)
		//					sender.Reply(fmt.Sprintf("很棒，许愿币+1，余额%d", AddCoin(sender.UserID)))
		//					logs.Info(msg)
		//				}
		//			}
		//
		//		}
		//		go func() {
		//			Save <- &JdCookie{}
		//		}()
		//		return nil
		//	}
		//}
		//{
		//	//k1k
		//	ss := regexp.MustCompile(`launchid=(\S+)(&|&amp;)ptag`).FindStringSubmatch(msg)
		//	if len(ss) > 0 {
		//		if !sender.IsAdmin {
		//			sender.Reply("仅管理员可用")
		//		} else {
		//			sender.Reply(fmt.Sprintf("砍价开始，管理员通道"))
		//			runTask(&Task{Path: "jd_kanjia.js", Envs: []Env{
		//				{Name: "launchid", Value: ss[1]},
		//			}}, sender)
		//		}
		//		return nil
		//	}
		//}

		{ //tyt
			if strings.Contains(msg, "49f40d2f40b3470e8d6c39aa4866c7ff") {
				no := tytno
				tytno += 1
				split := strings.Split(msg, "&amp;")
				for i := range split {
					if strings.Contains(split[i], "packetId=") {
						//f, err := os.OpenFile(ExecPath+"/tytlj.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
						//if err != nil {
						//	logs.Warn("tytlj.txt失败，", err)
						//}
						//logs.Info(split[i])
						env := strings.Split(split[i], "=")
						if strings.Contains(env[1], "微信") {
							sender.Reply("微信渠道暂时无法识别")
						}
						//f.WriteString(env[1] + "\n")
						//f.Close()
						if !sender.IsAdmin {
							coin := GetCoin(sender.UserID)
							if coin < Config.Tyt {
								return fmt.Sprintf("推一推需要%d个积分", Config.Tyt)
							}
							RemCoin(sender.UserID, Config.Tyt)

							sender.Reply(fmt.Sprintf("推一推即将开始，已扣除%d个积分,订单编号:%d，剩余%d", Config.Tyt, no, GetCoin(sender.UserID)))
						} else {
							sender.Reply(fmt.Sprintf("推一推即将开始，已扣除%d个积分，管理员通道", Config.Tyt))
						}
						//runTask(&Task{Path: "jd_tyt.js", Envs: []Env{
						//	{Name: "tytpacketId", Value: env[1]},
						//}}, sender)
						tytlist[env[1]] = no
						go runtyt(sender, env[1])
						//return fmt.Sprintf("订单编号：%d,推一推结束", no)
					}
				}
			}
		}

		{
			if strings.Contains(msg, "pt_key") {
				logs.Info(msg + "开始CK登录")
				ptKey := FetchJdCookieValue("pt_key", msg)
				ptPin := FetchJdCookieValue("pt_pin", msg)
				if len(ptPin) > 0 && len(ptKey) > 0 {
					ck := JdCookie{
						PtKey: ptKey,
						PtPin: ptPin,
					}
					if CookieOK(&ck) {
						if sender.IsQQ() {
							ck.QQ = sender.UserID
						} else if sender.IsTG() {
							ck.Telegram = sender.UserID
						}
						if HasKey(ck.PtKey) {
							sender.Reply(fmt.Sprintf("重复提交"))
						} else {
							if nck, err := GetJdCookie(ck.PtPin); err == nil {
								nck.InPool(ck.PtKey)
								msg := fmt.Sprintf("更新账号，%s", ck.PtPin)
								if sender.IsQQ() {
									ck.Update(QQ, ck.QQ)
								}
								sender.Reply(fmt.Sprintf(msg))
								(&JdCookie{}).Push(msg)
								logs.Info(msg)
							} else {
								if Cdle {
									ck.Hack = True
								}
								NewJdCookie(&ck)
								msg := fmt.Sprintf("添加账号，账号名:%s", ck.PtPin)
								if sender.IsQQ() {
									ck.Update(QQ, ck.QQ)
								}
								sender.Reply(fmt.Sprintf(msg))
								sender.Reply(ck.Query())
								(&JdCookie{}).Push(msg)
								logs.Info(msg)
							}
						}
					} else {
						sender.Reply(fmt.Sprintf("无效"))
					}
				}
				go func() {
					Save <- &JdCookie{}
				}()
				return nil
			}
		}

		for k, v := range replies {
			if regexp.MustCompile(k).FindString(msg) != "" {
				if strings.Contains(msg, "妹") && time.Now().Unix()%10 == 0 {
					v = "https://pics4.baidu.com/feed/d833c895d143ad4bfee5f874cfdcbfa9a60f069b.jpeg?token=8a8a0e1e20d4626cd31c0b838d9e4c1a"
				}
				if regexp.MustCompile(`^https{0,1}://[^\x{4e00}-\x{9fa5}\n\r\s]{3,}$`).FindString(v) != "" {
					url := v
					rsp, err := httplib.Get(url).Response()
					if err != nil {
						return nil
					}
					ctp := rsp.Header.Get("content-type")
					if ctp == "" {
						rsp.Header.Get("Content-Type")
					}
					if strings.Contains(ctp, "text") || strings.Contains(ctp, "json") {
						data, _ := ioutil.ReadAll(rsp.Body)
						return string(data)
					}
					return rsp
				}
				return v
			}
		}

	}
	return nil
}

func runtyt(sender *Sender, code string) {
	for {
		time.Sleep(time.Duration(rand.Intn(60)))
		if tytnum < 3 {
			tytnum++
			runTask(&Task{Path: "jd_tyt.js", Envs: []Env{
				{Name: "tytpacketId", Value: code},
			}}, sender)

			no := tytlist[code]
			sender.Reply(fmt.Sprintf("订单编号：%d,推一推结束", no))
			tytnum--
			return
		}
	}
}
func startpz(invited string) (num int, flag bool) {
	logs.Info("开始膨胀助力")
	k := 0
	cks := GetJdCookies()
	for i := len(cks); i > 0; i-- {
		time.Sleep(time.Second * time.Duration(3))
		cookie := "pt_key=" + cks[i-1].PtKey + ";pt_pin=" + cks[i-1].PtPin + ";"
		sc := getScKey(cookie)
		if sc != "" {
			url := "https://api.m.jd.com/client.action?functionId=tigernian_pk_collectPkExpandScore"
			body := fmt.Sprintf(`{"ss":"{\"extraData\":{\"log\":\"\",\"sceneid\":\"HYGJZYh5\"},\"secretp\":\"%s\",\"random\":\"%d\"}","inviteId":"%s"}`, sc, rand.Intn(99999999), invited)
			req := httplib.Post(url)
			random := browser.Random()
			req.Param("clientVersion", "1.0.0")
			req.Param("client", "wh5")
			req.Param("functionId", "tigernian_pk_collectPkExpandScore")
			req.Param("body", body)
			req.Header("User-Agent", random)
			req.Header("Accept", "application/json, text/plain, */*")
			req.Header("Connection", "keep-alive")
			req.Header("Accept-Language", "zh-cn")
			req.Header("Accept-Encoding", "gzip, deflate, br")
			req.Header("Origin", "https://wbbny.m.jd.com")
			req.Header("Cookie", cookie)
			s, _ := req.String()
			bizCode, _ := jsonparser.GetInt([]byte(s), "data", "bizCode")
			bizMsg, _ := jsonparser.GetString([]byte(s), "data", "bizMsg")

			if bizCode == 0 {
				k++
				logs.Info("助力成功")

			} else {
				logs.Info("助力失败")
				logs.Info(s)
				if strings.Contains(bizMsg, "好友人气爆棚") {
					return k, true
				}
			}
		}
	}
	return k, false

}

func getScKey(ck string) (key string) {
	url := "https://api.m.jd.com/client.action?functionId=tigernian_getHomeData"
	req := httplib.Get(url)
	random := browser.Random()
	req.Param("clientVersion", "1.0.0")
	req.Param("client", "wh5")
	req.Param("functionId", "tigernian_getHomeData")
	req.Header("User-Agent", random)
	req.Header("Host", "api.m.jd.com")
	req.Header("Accept", "application/json, text/plain, */*")
	req.Header("Connection", "keep-alive")
	req.Header("Accept-Language", "zh-cn")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Origin", "https://api.m.jd.com")
	req.Header("Cookie", ck)
	data, _ := req.String()
	if strings.Contains(data, "secretp") {
		index := strings.Index(data, "\"secretp\":") + 11
		i := strings.Index(data, "shareMiniprogramSwitch") - 3
		s := data[index:i]
		return s
	}
	return ""
}

//随机slice数组
func randShuffle(slice []JdCookie) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

func starttyt(red string) (num int, f bool) {
	k := 0
	//cks := GetJdCookies(func(sb *gorm.DB) *gorm.DB {
	//	return sb.Where(fmt.Sprintf("%s != ? and %s = ? ORDER BY RAND()", Tyt, Available), False, True)
	//})
	var cks []JdCookie
	db.Where(fmt.Sprintf("%s = 'true' and %s = 'true'", Tyt, Available)).Find(&cks)
	randShuffle(cks)
	logs.Info(len(cks))
	if len(cks) < 50 {
		(&JdCookie{}).Push("推一推账号不足  注意补单")
		return k, false
	}
	for _, ck := range cks {
		time.Sleep(time.Second * 10)
		logs.Info(ck.PtPin)
		cookie := "pt_key=" + ck.PtKey + ";pt_pin=" + ck.PtPin + ";"
		sprintf := fmt.Sprintf(`https://api.m.jd.com/?functionId=helpCoinDozer&appid=station-soa-h5&client=H5&clientVersion=1.0.0&t=1641900500241&body={"actId":"49f40d2f40b3470e8d6c39aa4866c7ff","channel":"coin_dozer","referer":"-1","frontendInitStatus":"s","packetId":"%s","helperStatus":"0"}&_ste=1`, red)
		req := httplib.Post(sprintf)
		random := browser.Random()
		req.Header("User-Agent", random)
		req.Header("Host", "api.m.jd.com")
		req.Header("Accept", "application/json, text/plain, */*")
		req.Header("Origin", "https://pushgold.jd.com")
		req.Header("Cookie", cookie)
		data, _ := req.String()
		code, _ := jsonparser.GetInt([]byte(data), "code")
		logs.Info(data)
		if code == 0 {
			k++
			logs.Info(jsonparser.GetString([]byte(data), "data", "amount"))
		} else {
			if strings.Contains(data, "完成") {
				logs.Info("返回完成")
				return k, true
			} else if strings.Contains(data, "帮砍机会已用完") {
				ck.Update(Tyt, False)
			} else if strings.Contains(data, "火爆") {
				ck.Update(Tyt, False)
			} else if strings.Contains(data, "帮砍排队") {
				return k, false
			} else if strings.Contains(data, "need") {
				ck.Update(Tyt, "need verity")
			} else if strings.Contains(data, "未登录") {
				CookieOK(&ck)
			} else {
				getString, _ := jsonparser.GetString([]byte(data), "msg")
				ck.Update(Tyt, getString)
				logs.Info(getString)
			}
		}
	}
	return k, false
}

func FetchJdCookieValue(key string, cookies string) string {
	match := regexp.MustCompile(key + `=([^;]*);{0,1}`).FindStringSubmatch(cookies)
	if len(match) == 2 {
		return match[1]
	} else {
		return ""
	}
}
