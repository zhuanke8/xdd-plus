package models

import (
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/buger/jsonparser"
	"io/ioutil"
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

var ListenQQPrivateMessage = func(uid int64, msg string) {
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

var pcodes = make(map[string]string)
var replies = map[string]string{}

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

func findMapKey3(str string, m map[string]string) string {
	if val, ok := m[str]; ok {
		fmt.Println("查询到", str, "手机号为：", val)
		return val
	} else {
		fmt.Println("未能检索到该数据")
	}
	return ""
}

var handleMessage = func(msgs ...interface{}) interface{} {
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
	switch msg {
	default:
		{
			regex := "^\\d{6}$"
			reg := regexp.MustCompile(regex)
			if Config.VIP {
				if reg.MatchString(msg) {
					logs.Info("进入验证码阶段")
					addr := Config.Jdcurl
					phone := findMapKey3(string(sender.UserID), pcodes)
					if phone != "" {
						req := httplib.Post(addr + "/api/VerifyCode")
						req.Header("content-type", "application/json")
						data, _ := req.Body(`{"Phone":"` + phone + `","QQ":"` + strconv.Itoa(sender.UserID) + `","qlkey":0,"Code":"` + msg + `"}`).Bytes()
						message, _ := jsonparser.GetString(data, "message")
						if strings.Contains(string(data), "pt_pin=") {
							sender.Reply("登录成功。可以继续登录下一个账号")
							if strings.Contains(msg, "pt_key") {
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
						} else if strings.Contains(message, "添加xdd成功") {
							sender.Reply("登录成功。可以继续登录下一个账号")
						} else {
							if message != "" {
								sender.Reply(message)
							} else {
								sender.Reply("登录失败。请重新登录")
							}
						}
					}
				}
			}
		}
		{
			if Config.VIP {
				ist := findMapKey3(string(sender.UserID), pcodes)
				if strings.EqualFold(ist, "true") {
					regular := `^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\d{8}$`
					reg := regexp.MustCompile(regular)
					if reg.MatchString(msg) {
						addr := Config.Jdcurl
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
						if !success && status == 666 && i < 5 {

							sender.Reply("正在进行滑块验证...")
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
								if i > 5 {
									sender.Reply("滑块验证失败,请联系管理员或者手动登录")
									break
								}
								if status == 666 {
									i++
									sender.Reply(fmt.Sprintf("正在进行第%d次滑块验证...", i))
									continue
								}
								if success {
									pcodes[string(sender.UserID)] = msg
									sender.Reply("请输入6位验证码：")
									break
								}
								if strings.Contains(message, "上限") {
									i = 6
									sender.Reply(message)
								}
								//sender.Reply(message)
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
			if Config.VIP {
				if strings.Contains(msg, "登录") || strings.Contains(msg, "登陆") {
					var tabcount int64
					addr := Config.Jdcurl
					if addr == "" {
						return "若兰很忙，请稍后再试。"
					}
					logs.Info(addr + "/api/Config")
					if addr != "" {
						data, _ := httplib.Get(addr + "/api/Config").Bytes()
						tabcount, _ = jsonparser.GetInt(data, "data", "tabcount")
						if tabcount != 0 {

						} else {
							sender.Reply("服务忙，请稍后再试。")
						}
					}
					pcodes[string(sender.UserID)] = "true"
					sender.Reply("若兰为您服务，请输入11位手机号：")

				}
			}
		}
		{
			//发财挖宝
			if Config.VIP {
				//dyj
				inviterId := regexp.MustCompile(`inviterId=(\S+)(&|&amp;)inviterCode`).FindStringSubmatch(msg)
				inviterCode := regexp.MustCompile(`inviterCode=(\S+)(&|&amp;)utm_user`).FindStringSubmatch(msg)
				if len(inviterCode) == 0 {
					inviterCode = regexp.MustCompile(`inviterCode=(\S+)(&|&amp;)utm_source`).FindStringSubmatch(msg)
				}
				if len(inviterId) > 0 && len(inviterCode) > 0 {
					if !sender.IsAdmin {
						sender.Reply("仅管理员可用")
					} else {
						sender.Reply(fmt.Sprintf("发财挖宝开始，管理员通道"))
						num, num1, f := startfcwb(inviterId[1], inviterCode[1])
						if f {
							sender.Reply(fmt.Sprintf("助力完成，助力成功：%d个,无效助力账号:%d个", num, num1))
						}
					}
					return nil
				}
			}
		}
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
			if Config.VIP {
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
		}
		{
			ss := regexp.MustCompile(`pin=([^;=\s]+);wskey=([^;=\s]+)`).FindAllStringSubmatch(msg, -1)
			if len(ss) > 0 {
				for _, s := range ss {
					wkey := "pin=" + s[1] + ";wskey=" + s[2] + ";"
					//rsp := cmd(fmt.Sprintf(`python3 test.py "%s"`, wkey), &Sender{})
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
							WsKey: s[2],
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
		{
			//dyj
			inviterId := regexp.MustCompile(`inviterId=(\S+)(&|&amp;)helpType`).FindStringSubmatch(msg)
			redEnvelopeId := regexp.MustCompile(`redEnvelopeId=(\S+)(&|&amp;)inviterId`).FindStringSubmatch(msg)
			if len(inviterId) > 0 && len(redEnvelopeId) > 0 {
				if !sender.IsAdmin {
					sender.Reply("仅管理员可用")
				} else {
					sender.Reply(fmt.Sprintf("大赢家开始，管理员通道"))
					num, num1, f, f1 := startdyj(inviterId[1], redEnvelopeId[1], 1)
					if f {
						sender.Reply(fmt.Sprintf("助力完成，助力成功：%d个,火爆账号:%d个", num, num1))
						if f1 {
							sender.Reply("满足提现条件，开始自动提现助力")
							n, i, _, f12 := startdyj(inviterId[1], redEnvelopeId[1], 2)
							if f12 {
								sender.Reply(fmt.Sprintf("提现助力完成，助力成功：%d个,火爆账号:%d个", n, i))
							}
						}
					} else {
						sender.Reply(fmt.Sprintf("你已经黑IP拉！，助力成功：%d个,火爆账号:%d个", num, num1))
					}

				}
				return nil
			}

		}
		{
			//k1k
			ss := regexp.MustCompile(`launchid=(\S+)(&|&amp;)ptag`).FindStringSubmatch(msg)
			if len(ss) > 0 {
				if !sender.IsAdmin {
					sender.Reply("仅管理员可用")
				} else {
					sender.Reply(fmt.Sprintf("砍价开始，管理员通道"))
					runTask(&Task{Path: "jd_kanjia.js", Envs: []Env{
						{Name: "launchid", Value: ss[1]},
					}}, sender)
				}
				return nil
			}
		}
		{ //tyt
			ss := regexp.MustCompile(`packetId=(\S+)(&|&amp;)currentActId`).FindStringSubmatch(msg)
			if len(ss) > 0 {
				if !sender.IsAdmin {
					coin := GetCoin(sender.UserID)
					if coin < Config.Tyt {
						return fmt.Sprintf("推一推需要%d个互助值", Config.Tyt)
					}
					RemCoin(sender.UserID, Config.Tyt)
					sender.Reply(fmt.Sprintf("推一推即将开始，已扣除%d个互助值", Config.Tyt))
				} else {
					sender.Reply(fmt.Sprintf("推一推即将开始，已扣除%d个互助值，管理员通道", Config.Tyt))
				}

				runTask(&Task{Path: "jd_tyt.js", Envs: []Env{
					{Name: "tytpacketId", Value: ss[1]},
				}}, sender)
				return nil
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
		{
			o := findShareCode(msg)
			if o != "" {
				return "导入互助码成功"
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

func startdyj(ine string, red string, type1 int) (num int, num1 int, f bool, f1 bool) {
	k := 0
	n := 0
	cks := GetJdCookies()
	for i := range cks {
		time.Sleep(time.Second * time.Duration(5))
		cookie := "pt_key=" + cks[i].PtKey + ";pt_pin=" + cks[i].PtPin + ";"
		sprintf := fmt.Sprintf(`https://api.m.jd.com/client.action?functionId=openRedEnvelopeInteract&body={"linkId":"PFbUR7wtwUcQ860Sn8WRfw","redEnvelopeId":"%s","inviter":"%s","helpType":"%d"}&t=1626363029817&appid=activities_platform&clientVersion=3.5.0`, red, ine, type1)
		req := httplib.Get(sprintf)
		random := browser.Random()
		req.Header("User-Agent", random)
		req.Header("Host", "api.m.jd.com")
		req.Header("Accept", "application/json, text/plain, */*")
		req.Header("Connection", "keep-alive")
		req.Header("Accept-Language", "zh-cn")
		req.Header("Accept-Encoding", "gzip, deflate, br")
		req.Header("Origin", "https://618redpacket.jd.com")
		req.Header("Cookie", cookie)
		data, _ := req.String()
		if strings.Contains(data, "助力成功") {
			logs.Info("助力成功")
			k++
		} else if strings.Contains(data, "火爆") {
			logs.Info("火爆了")
			n++
		} else if strings.EqualFold(data, "") {
			return i, n, false, false
		} else if strings.Contains(data, "今日帮好友拆红包次数已达上限") {
			logs.Info("助力上限")
		} else if strings.Contains(data, "已成功提现") {
			return i, n, true, true
		} else {
			logs.Info("要么助力过了，要么没登录")
		}
	}
	return k, n, true, false
}

func startfcwb(ine string, red string) (num int, num1 int, f bool) {
	logs.Info("开始发财挖宝")
	k := 0
	n := 0
	cks := GetJdCookies()
	for i := len(cks); i > 0; i-- {
		if k > 125 {
			return k, n, true
		}
		time.Sleep(time.Second * time.Duration(3))
		cookie := "pt_key=" + cks[i-1].PtKey + ";pt_pin=" + cks[i-1].PtPin + ";"
		//https://api.m.jd.com/?functionId=happyDigHelp&body={"linkId":"pTTvJeSTrpthgk9ASBVGsw","inviter":"-ftyyGV7YwjPJ63tnKLwjw","inviteCode":"7476e3bed5d74f74b0a547b7e4d1e07225061636959196596"}&t=1635561607124&appid=activities_platform&client=H5&clientVersion=1.0.0
		sprintf := fmt.Sprintf(`https://api.m.jd.com/?functionId=happyDigHelp&body={"linkId":"pTTvJeSTrpthgk9ASBVGsw","inviter":"%s","inviteCode":"%s"}&t=1635561607124&appid=activities_platform&client=H5&clientVersion=1.0.0`, ine, red)
		logs.Info(sprintf)
		req := httplib.Get(sprintf)
		random := browser.Random()
		req.Header("User-Agent", random)
		//req.Header("Host", "api.m.jd.com")
		req.Header("Accept", "application/json, text/plain, */*")
		req.Header("Connection", "keep-alive")
		req.Header("Accept-Language", "zh-cn")
		req.Header("Accept-Encoding", "gzip, deflate, br")
		req.Header("Origin", "https://api.m.jd.com")
		req.Header("Cookie", cookie)

		data, _ := req.String()
		logs.Info(data)
		if strings.Contains(data, "true") {
			logs.Info("助力成功")
			k++
		} else if strings.Contains(data, "已经邀请过") {
			logs.Info("已经邀请过")
		} else if strings.Contains(data, "火爆") {
			logs.Info("火爆了")
			n++
		} else {
			logs.Info("要么助力过了，要么没登录,要么火爆")
		}
	}

	return k, n, true
}

func FetchJdCookieValue(key string, cookies string) string {
	match := regexp.MustCompile(key + `=([^;]*);{0,1}`).FindStringSubmatch(cookies)
	if len(match) == 2 {
		return match[1]
	} else {
		return ""
	}
}
