package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
	"github.com/buger/jsonparser"
)

const (
	QL = "ql"
	V4 = "v4"
	LI = "li"
)

type Container struct {
	Type      string
	Name      string
	Default   bool
	Address   string
	Username  string
	Password  string
	Cid       string
	Secret    string
	Path      string
	Version   string
	Token     string
	Available bool
	Delete    []string
	Weigth    int
	Mode      string
	Reader    *bufio.Reader
	Config    string
	Limit     int
	cks       []JdCookie
	Zhu       int
	Ci        int
	Resident  string
}

func initContainer() {
	var ql = 0
	for i := range Config.Containers {
		if Config.Containers[i].Weigth == 0 {
			Config.Containers[i].Weigth = 1
		}
		Config.Containers[i].Type = ""
		if Config.Containers[i].Address != "" {
			vv := regexp.MustCompile(`^(https?://[\.\w]+:?\d*)`).FindStringSubmatch(Config.Containers[i].Address)
			if len(vv) == 2 {
				Config.Containers[i].Address = vv[1]
			} else {
				logs.Warn("%s地址错误", Config.Containers[i].Type)
			}
			version, err := GetQlVersion(Config.Containers[i].Address)
			if err == nil {
				if Config.Containers[i].getToken() == nil {
					logs.Info("青龙" + version + "通道登录成功")
					ql++
				} else {
					logs.Warn("青龙" + version + "通道登录失败")
				}
				Config.Containers[i].Type = "ql"
				Config.Containers[i].Version = version
			} else {
				if err := Config.Containers[i].getSession(); err == nil {
					logs.Info("v系登录成功")
				} else {
					(&JdCookie{}).Push("容器出现错误连接" + Config.Containers[i].Address)
					logs.Info("v系登录失败")
				}
				Config.Containers[i].Type = "v4"
			}
		} else if Config.Containers[i].Path != "" {
			f, err := os.Open(Config.Containers[i].Path)
			if err != nil {
				logs.Warn("无法打开%s，请检查路径是否正确", Config.Containers[i].Path)
			} else {
				rd := bufio.NewReader(f)
				for {
					line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
					if err != nil || io.EOF == err {
						break
					}
					if pt := regexp.MustCompile(`^pt_key=`).FindString(line); pt != "" {
						Config.Containers[i].Type = "li"
						break
					}
					if pt := regexp.MustCompile(`^Cookie\d+`).FindString(line); pt != "" {
						Config.Containers[i].Type = "v4"
						break
					}
					if strings.Contains(line, "TempBlockCookie") {
						Config.Containers[i].Type = "v4"
						break
					}
					if strings.Contains(line, "QYWX_KEY") {
						Config.Containers[i].Type = "v4"
						break
					}
				}
				if Config.Containers[i].Type == "" {
					Config.Containers[i].Type = "li"
				}
				f.Close()
				logs.Info(Config.Containers[i].Type + "配置文件正确")
			}
		}
	}

	(&JdCookie{}).Push(fmt.Sprintf("容器加载完毕，共%d个", ql))
}

func (c *Container) write(cks []JdCookie) error {
	switch c.Type {
	case "ql":
		if c.Version == "openapi" {
			if len(c.Delete) > 0 {
				c.request("/api/envs", DELETE, fmt.Sprintf(`[%s]`, strings.Join(c.Delete, ",")))
			}
			hh := []string{}
			if len(cks) != 0 {
				for _, ck := range cks {
					if ck.Available == True {
						if Config.Note == "pin" {
							hh = append(hh, fmt.Sprintf(`{"name":"JD_COOKIE","value":"pt_key=%s;pt_pin=%s;","remarks":"%s"}`, ck.PtKey, ck.PtPin, ck.Nickname))
						} else {
							hh = append(hh, fmt.Sprintf(`{"name":"JD_COOKIE","value":"pt_key=%s;pt_pin=%s;","remarks":"%s"}`, ck.PtKey, ck.PtPin, ck.Note))
						}
					}
				}
				c.request("/api/envs", POST, fmt.Sprintf(`[%s]`, strings.Join(hh, ",")))
				if Config.IsHelp {
					type AutoGenerated struct {
						Code int `json:"code"`
						Data []struct {
							Value     string  `json:"value"`
							ID        string  `json:"_id"`
							IntID     int     `json:"id"`
							Created   int64   `json:"created"`
							Status    int     `json:"status"`
							Timestamp string  `json:"timestamp"`
							Position  float64 `json:"position"`
							Name      string  `json:"name"`
							Remarks   string  `json:"remarks,omitempty"`
						} `json:"data"`
					}
					help := getQLHelp(len(cks))
					for k := range help {
						var data, err = c.request("/api/envs?searchValue=" + k)
						a := AutoGenerated{}
						err = json.Unmarshal(data, &a)
						if err != nil {
							continue
						}
						toDelete := []string{}
						for _, env := range a.Data {
							if env.IntID == 0 {
								toDelete = append(toDelete, fmt.Sprintf("\"%s\"", env.ID))
							} else {
								toDelete = append(toDelete, fmt.Sprintf("\"%d\"", env.IntID))
							}
							//toDelete = append(toDelete, fmt.Sprintf("\"%s\"", env.ID))
						}
						logs.Info(toDelete)
						if len(toDelete) > 0 {
							c.request("/api/envs", DELETE, fmt.Sprintf(`[%s]`, strings.Join(toDelete, ",")))
						}
					}
					for k, v := range help {
						if v == "" {
							v = "&"
						}
						r := map[string]string{
							"name":  k,
							"value": v,
						}
						d, _ := json.Marshal(r)
						c.request("/api/envs", POST, fmt.Sprintf(`[%s]`, string(d)))
					}
				}
			}
		} else {
			if len(c.Delete) > 0 {
				c.request("/api/cookies", DELETE, fmt.Sprintf(`[%s]`, strings.Join(c.Delete, ",")))
			}
			d := []string{}
			for _, ck := range cks {
				if ck.Available == True {
					d = append(d, fmt.Sprintf("\"pt_key=%s;pt_pin=%s;\"", ck.PtKey, ck.PtPin))
				}
			}
			if len(d) != 0 {
				c.request("/api/cookies", POST, fmt.Sprintf(`[%s]`, strings.Join(d, ",")))
			}
		}
	case "v4":
		return c.postConfig(func(config string) string {
			TempBlockCookie := ""
			cookies := ""
			for i, ck := range cks {
				if ck.Available == False {
					TempBlockCookie += fmt.Sprintf("%d ", i+1)
				}
				cookies += fmt.Sprintf("Cookie%d=\"pt_key=%s;pt_pin=%s;\"\n", i+1, ck.PtKey, ck.PtPin)
			}
			config = fmt.Sprintf(`TempBlockCookie="%s"`, TempBlockCookie) + "\n" + cookies + getVhelpRule(len(cks)) + config
			return config
		})
	case "li":
		config := ""
		f, err := os.OpenFile(c.Path, os.O_RDWR|os.O_CREATE, 0777) //打开文件 |os.O_RDWR
		if err != nil {
			return err
		}
		defer f.Close()
		rd := bufio.NewReader(f)
		for {
			line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
			if err != nil || io.EOF == err {
				break
			}
			if pt := regexp.MustCompile(`^pt_key=(.*);pt_pin=([^'";\s]+);?`).FindStringSubmatch(line); len(pt) != 0 {
				continue
			}
			if pt := regexp.MustCompile(`^pt_key=(.*)`).FindStringSubmatch(line); len(pt) != 0 {
				continue
			}
			config += line
		}
		for _, ck := range cks {
			if ck.PtPin == "" || ck.PtKey == "" {
				continue
			}
			if ck.Available == True {
				config += fmt.Sprintf("pt_key=%s;pt_pin=%s\n", ck.PtKey, ck.PtPin)
			}
		}
		f.Truncate(0)
		f.Seek(0, 0)
		if _, err := io.WriteString(f, config); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (c *Container) read() error {
	c.Available = true
	switch c.Type {
	case "ql":
		if c.Version == "openapi" {
			type AutoGenerated struct {
				Code int `json:"code"`
				Data []struct {
					Value     string  `json:"value"`
					ID        string  `json:"_id"`
					IntID     int     `json:"id"`
					Created   int64   `json:"created"`
					Status    int     `json:"status"`
					Timestamp string  `json:"timestamp"`
					Position  float64 `json:"position"`
					Name      string  `json:"name"`
					Remarks   string  `json:"remarks,omitempty"`
				} `json:"data"`
			}
			var data, err = c.request("/api/envs?searchValue=JD_COOKIE")
			a := AutoGenerated{}
			err = json.Unmarshal(data, &a)
			if err != nil {
				c.Available = false
				logs.Info(c.Address + "地址错误")
				return err
			}
			c.Delete = []string{}

			for _, env := range a.Data {
				if env.IntID == 0 {
					c.Delete = append(c.Delete, fmt.Sprintf("\"%s\"", env.ID))
				} else {
					c.Delete = append(c.Delete, fmt.Sprintf("\"%d\"", env.IntID))
				}
				//c.Delete = append(c.Delete, fmt.Sprintf("\"%s\"", env.ID))
				res := regexp.MustCompile(`pt_key=(\S+);pt_pin=([^\s;]+);?`).FindAllStringSubmatch(env.Value, -1)
				for _, v := range res {
					CheckIn(v[2], v[1])
				}
			}
			return nil
		} else {
			var data, err = c.request("/api/cookies")
			if err != nil {
				c.Available = false
				return err
			}
			type AutoGenerated struct {
				Code int `json:"code"`
				Data []struct {
					Value     string  `json:"value"`
					ID        string  `json:"_id"`
					Created   int64   `json:"created"`
					Status    int     `json:"status"`
					Timestamp string  `json:"timestamp"`
					Position  float64 `json:"position"`
					Nickname  string  `json:"nickname"`
				} `json:"data"`
			}
			var a = AutoGenerated{}
			c.Delete = []string{}
			json.Unmarshal(data, &a)
			for _, vv := range a.Data {
				c.Delete = append(c.Delete, fmt.Sprintf("\"%s\"", vv.ID))
				res := regexp.MustCompile(`pt_key=(\S+);pt_pin=([^\s;]+);?`).FindStringSubmatch(vv.Value)
				if len(res) == 3 {
					CheckIn(res[2], res[1])
				}
			}
		}
	case "v4":
		return c.getConfig(func(rd *bufio.Reader) string {
			config := ""
			for {
				line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
				if err != nil || io.EOF == err {
					config += line
					break
				}
				if pt := regexp.MustCompile(`^#?\s?Cookie(\d+)=\S+pt_key=(.+);pt_pin=([^'";\s]+);?`).FindStringSubmatch(line); len(pt) != 0 {
					CheckIn(pt[3], pt[2])
					continue
				}
				if pt := regexp.MustCompile(`^ForOther`).FindString(line); pt != "" {
					continue
				}
				if pt := regexp.MustCompile(`^My.*\d+=`).FindString(line); pt != "" {
					continue
				}
				if pt := regexp.MustCompile(`^Cookie\d+`).FindString(line); pt != "" {
					continue
				}
				if pt := regexp.MustCompile(`^TempBlockCookie`).FindString(line); pt != "" {
					continue
				}
				config += line
			}
			return config
		})
	case "li":
		f, err := os.OpenFile(c.Path, os.O_RDWR|os.O_CREATE, 0777) //打开文件 |os.O_RDWR
		if err != nil {
			c.Available = false
			return err
		}
		defer f.Close()
		rd := bufio.NewReader(f)
		for {
			line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
			if err != nil || io.EOF == err {
				break
			}
			if pt := regexp.MustCompile(`^pt_key=(.+);pt_pin=([^'";\s]+);?`).FindStringSubmatch(line); len(pt) != 0 {
				CheckIn(pt[2], pt[1])
				continue
			}
		}
	}
	return nil
}

func (c *Container) getToken() error {
	version, _ := GetQlVersion(c.Address)
	if version == "openapi" {
		token := &Token{}
		err, b2 := getT(c, token)
		if b2 {
			c.Token = token.Token
		}
		return err
	} else {
		req := httplib.Post(c.Address + "/api/login")
		req.Header("Content-Type", "application/json;charset=UTF-8")
		req.Body(fmt.Sprintf(`{"username":"%s","password":"%s"}`, c.Username, c.Password))
		if rsp, err := req.Response(); err == nil {
			data, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				return err
			}
			c.Token, _ = jsonparser.GetString(data, "token")
			if c.Token == "" {
				c.Token, _ = jsonparser.GetString(data, "data", "token")
			}
		} else {
			return err
		}
	}
	return nil
}

func getT(c *Container, token *Token) (error, bool) {
	req := httplib.Get(c.Address + fmt.Sprintf(`/open/auth/token?client_id=%s&client_secret=%s`, c.Cid, c.Secret))
	req.Header("Content-Type", "application/json;charset=UTF-8")
	if rsp, err := req.Response(); err == nil {
		data, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return err, true
		}
		c.Token, _ = jsonparser.GetString(data, "data", "token")
		token.Token, _ = jsonparser.GetString(data, "data", "token")
		zero, _ := time.ParseInLocation("2006-01-02", time.Now().Local().Format("2006-01-02"), time.Local)
		token.Expiration = zero
		token.Address = c.Address
	} else {
		return err, true
	}
	return nil, false
}

func (c *Container) request(ss ...string) ([]byte, error) {
	var api, method, body string
	for _, s := range ss {
		if s == GET || s == POST || s == PUT || s == DELETE {
			method = s
		} else if strings.Contains(s, "/api/") {
			if c.Version == "openapi" {
				api = strings.ReplaceAll(s, "api", "open")
			} else {
				api = s
			}
		} else {
			body = s
		}
	}
	var req *httplib.BeegoHTTPRequest
	var i = 0
	for {
		i++
		switch method {
		case POST:
			req = httplib.Post(c.Address + api)
		case PUT:
			req = httplib.Put(c.Address + api)
		case DELETE:
			req = httplib.Delete(c.Address + api)
		default:
			req = httplib.Get(c.Address + api)
		}
		req.Header("Authorization", "Bearer "+c.Token)
		if body != "" {
			req.Header("Content-Type", "application/json;charset=UTF-8")
			req.Body(body)
		}
		if data, err := req.Bytes(); err == nil {
			code, _ := jsonparser.GetInt(data, "code")
			if code == 200 {
				return data, nil
			} else {
				logs.Warn(string(data))
				logs.Info(c.Address + "异常")
				if i >= 5 {
					return nil, errors.New("异常")
				}
				c.getToken()
			}
		}
	}
	return []byte{}, nil
}

func GetQlVersion(address string) (string, error) {
	data, err := httplib.Get(address).String()
	if err != nil {
		return "", err
	}
	js := regexp.MustCompile(`/umi\.\w+\.js`).FindString(data)
	if js == "" {
		return "", errors.New("好像不是青龙面板")
	}
	data, err = httplib.Get(address + js).String()
	if err != nil {
		return "", err
	}
	v := ""
	//logs.Info(data)
	if strings.Contains(data, "v2.2") {
		v = "api"
	} else {
		v = "openapi"
	}
	return v, nil
}

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

func (c *Container) getConfig(handle func(*bufio.Reader) string) error {
	if c.Address == "" {
		f, err := os.OpenFile(c.Path, os.O_RDWR|os.O_CREATE, 0777) //打开文件 |os.O_RDWR
		if err != nil {
			return err
		}
		defer f.Close()
		c.Config = handle(bufio.NewReader(f))
	} else {
		err := c.getSession()
		if err != nil {
			return err
		}
		req := httplib.Get(c.Address + "/api/config/config")
		req.Header("Cookie", c.Token)
		rsp, err := req.Response()
		if err != nil {
			return err
		}
		c.Config = handle(bufio.NewReader(rsp.Body))
	}
	return nil
}

func (c *Container) postConfig(handle func(config string) string) error {
	if c.Address == "" {
		f, err := os.OpenFile(c.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			return err
		}
		defer f.Close()
		f.WriteString(handle(c.Config))
	} else {
		req := httplib.Post(c.Address + "/api/save")
		req.Header("Cookie", c.Token)
		req.Param("content", handle(c.Config))
		req.Param("name", "config.sh")
		_, err := req.Bytes()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) getSession() error {

	if Config.IsOldV4 {
		req := httplib.Post(c.Address + "/api/auth")
		req.Param("username", c.Username)
		req.Param("password", c.Password)
		rsp, err := req.Response()
		if err != nil {
			return err
		}
		c.Token = rsp.Header.Get("Set-Cookie")
		if data, err := ioutil.ReadAll(rsp.Body); err != nil {
			return err
		} else {
			err, _ := jsonparser.GetInt(data, "err")
			if err != 0 {
				return errors.New(string(data))
			}
		}
		return nil
	} else {
		req := httplib.Post(c.Address + "/auth")
		req.Param("username", c.Username)
		req.Param("password", c.Password)
		rsp, err := req.Response()
		if err != nil {
			return err
		}
		c.Token = rsp.Header.Get("Set-Cookie")
		if data, err := ioutil.ReadAll(rsp.Body); err != nil {
			return err
		} else {
			err, _ := jsonparser.GetInt(data, "err")
			if err != 0 {
				return errors.New(string(data))
			}
		}
		return nil
	}

}
