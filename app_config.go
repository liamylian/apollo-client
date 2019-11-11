package apollo_client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	AppConfigFileName = "app.properties"
	ConfigFilePath    = "AGOLLO_CONF"
)

var (
	longPollInterval       = 2 * time.Second //2s
	longPollConnectTimeout = 1 * time.Minute //1m

	connectTimeout = 1 * time.Second //1s
	//notify timeout
	notifyConnectTimeout = 10 * time.Minute //10m
	//for on error retry
	onErrorRetryInterval = 1 * time.Second //1s
	//for typed config agcache of parser result, e.g. integer, double, long, etc.
	//max_config_cache_size    = 500             //500 agcache key
	//config_cache_expire_time = 1 * time.Minute //1 minute

	//max retries connect apollo
	maxRetries = 5

	//refresh ip list
	refreshIpListInterval = 20 * time.Minute //20m

	//appconfig
	appConfig *AppConfig

	//real servers ip
	servers sync.Map

	//next try connect period - 60 second
	nextTryConnectPeriod int64 = 60
)

type AppConfig struct {
	AppId            string `json:"appId"`
	Cluster          string `json:"cluster"`
	NamespaceName    string `json:"namespaceName"`
	Ip               string `json:"ip"`
	NextTryConnTime  int64  `json:"-"`
	BackupConfigPath string `json:"backupConfigPath"`
}

func (c *AppConfig) getBackupConfigPath() string {
	return c.BackupConfigPath
}

func (c *AppConfig) getHost() string {
	if strings.HasPrefix(c.Ip, "http") {
		if !strings.HasSuffix(c.Ip, "/") {
			return c.Ip + "/"
		}
		return c.Ip
	}
	return "http://" + c.Ip + "/"
}

//if this connect is fail will set this time
func (c *AppConfig) setNextTryConnTime(nextTryConnectPeriod int64) {
	c.NextTryConnTime = time.Now().Unix() + nextTryConnectPeriod
}

//is connect by ip directly
//false : no
//true : yes
func (c *AppConfig) isConnectDirectly() bool {
	if c.NextTryConnTime >= 0 && c.NextTryConnTime > time.Now().Unix() {
		return true
	}

	return false
}

func (c *AppConfig) selectHost() string {
	if !c.isConnectDirectly() {
		return c.getHost()
	}

	host := ""

	servers.Range(func(k, v interface{}) bool {
		server := v.(*serverInfo)
		// if some node has down then select next node
		if server.IsDown {
			return true
		}
		host = k.(string)
		return false
	})

	return host
}

func setDownNode(host string) {
	if host == "" || appConfig == nil {
		return
	}

	if host == appConfig.getHost() {
		appConfig.setNextTryConnTime(nextTryConnectPeriod)
	}

	servers.Range(func(k, v interface{}) bool {
		server := v.(*serverInfo)
		// if some node has down then select next node
		if k.(string) == host {
			server.IsDown = true
			return false
		}
		return true
	})
}

type serverInfo struct {
	AppName     string `json:"appName"`
	InstanceId  string `json:"instanceId"`
	HomepageUrl string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

func initFileConfig() {
	// default use application.properties
	initConfig(nil)
}

func initConfig(loadAppConfig func() (*AppConfig, error)) {
	var err error
	//init config file
	appConfig, err = getLoadAppConfig(loadAppConfig)

	if err != nil {
		return
	}

	func(appConfig *AppConfig) {
		splitNamespaces(appConfig.NamespaceName, func(namespace string) {
			apolloConfig := &ApolloConfig{}
			apolloConfig.init(appConfig, namespace)

			updateApolloConfig(apolloConfig, false)
		})
	}(appConfig)
}

// set load app config's function
func getLoadAppConfig(loadAppConfig func() (*AppConfig, error)) (*AppConfig, error) {
	if loadAppConfig != nil {
		return loadAppConfig()
	}
	configPath := os.Getenv(ConfigFilePath)
	if configPath == "" {
		configPath = AppConfigFileName
	}
	return loadJsonConfig(configPath)
}

//set timer for update ip list
//interval : 20m
func initServerIpList() {
	syncServerIpList(nil)
	logger.Debug("syncServerIpList started")

	t2 := time.NewTimer(refreshIpListInterval)
	for {
		select {
		case <-t2.C:
			syncServerIpList(nil)
			t2.Reset(refreshIpListInterval)
		}
	}
}

func syncServerIpListSuccessCallBack(responseBody []byte) (o interface{}, err error) {
	logger.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*serverInfo, 0)

	err = json.Unmarshal(responseBody, &tmpServerInfo)

	if err != nil {
		logger.Error("Unmarshal json Fail,Error:", err)
		return
	}

	if len(tmpServerInfo) == 0 {
		logger.Info("get no real server!")
		return
	}

	for _, server := range tmpServerInfo {
		if server == nil {
			continue
		}
		servers.Store(server.HomepageUrl, server)
	}
	return
}

//sync ip list from server
//then
//1.update agcache
//2.store in disk
func syncServerIpList(newAppConfig *AppConfig) error {
	appConfig := GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	_, err := request(getServicesConfigUrl(appConfig), &ConnectConfig{}, &CallBack{
		SuccessCallBack: syncServerIpListSuccessCallBack,
	})

	return err
}

func GetAppConfig(newAppConfig *AppConfig) *AppConfig {
	if newAppConfig != nil {
		return newAppConfig
	}
	return appConfig
}

func getConfigUrl(config *AppConfig) string {
	return getConfigUrlByHost(config, config.getHost())
}

func getConfigUrlByHost(config *AppConfig, host string) string {
	return fmt.Sprintf("%sconfigs/%s/%s/%s?releaseKey=%s&ip=%s",
		host,
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(config.NamespaceName),
		url.QueryEscape(getCurrentApolloConfigReleaseKey(config.NamespaceName)),
		getInternalIp())
}

func getConfigURLSuffix(config *AppConfig, namespaceName string) string {
	if config == nil {
		return ""
	}
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(getCurrentApolloConfigReleaseKey(namespaceName)),
		getInternalIp())
}

func getNotifyUrlSuffix(notifications string, config *AppConfig, newConfig *AppConfig) string {
	if newConfig != nil {
		return ""
	}
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(notifications))
}

func getServicesConfigUrl(config *AppConfig) string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		config.getHost(),
		url.QueryEscape(config.AppId),
		getInternalIp())
}
