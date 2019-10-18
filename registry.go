package go_eureka_client

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Port struct {
	Value   int  `json:"$"`
	Enabled bool `json:"@enabled"`
}

type DataCenterInfo struct {
	Class string `json:"@class"`
	Name  string `json:"name"`
}
type LeaseInfo struct {
	RenewalIntervalInSecs int64 `json:"renewalIntervalInSecs"`
	DurationInSecs        int64 `json:"durationInSecs"`
	RegistrationTimestamp int64 `json:"registrationTimestamp"`
	LastRenewalTimestamp  int64 `json:"lastRenewalTimestamp"`
	EvictionTimestamp     int64 `json:"evictionTimestamp"`
	ServiceUpTimestamp    int64 `json:"serviceUpTimestamp"`
}
type Metadata struct {
	//Attr           int `json:"$"`
	ManagementPort int `json:"management.port"`
}
type Instance struct {
	InstanceID                    string         `json:"instanceId"`
	HostName                      string         `json:"hostName"`
	App                           string         `json:"app"`
	IPAddr                        string         `json:"ipAddr"`
	Status                        string         `json:"instanceStatus"`
	Overriddenstatus              string         `json:"overriddenstatus"`
	Port                          Port           `json:"port"`
	SecurePort                    Port           `json:"securePort"`
	CountryID                     int            `json:"countryId"`
	DataCenterInfo                DataCenterInfo `json:"dataCenterInfo"`
	LeaseInfo                     LeaseInfo      `json:"leaseInfo"`
	Metadata                      Metadata       `json:"metadata"`
	HomePageURL                   string         `json:"homePageUrl"`
	StatusPageURL                 string         `json:"statusPageUrl"`
	HealthCheckURL                string         `json:"healthCheckUrl"`
	VipAddress                    string         `json:"vipAddress"`
	SecureVipAddress              string         `json:"secureVipAddress"`
	IsCoordinatingDiscoveryServer bool           `json:"isCoordinatingDiscoveryServer"`
	LastUpdatedTimestamp          int64          `json:"lastUpdatedTimestamp"`
	LastDirtyTimestamp            int64          `json:"lastDirtyTimestamp"`
	ActionType                    string         `json:"actionType"`
}

type Application struct {
	Instance Instance `json:"instance"`
}

type Option func(*Instance)

func newOptions(opts ...Option) *Instance {
	opt := &Instance{
		InstanceID:                    "",
		HostName:                      "",
		App:                           "",
		IPAddr:                        "",
		Status:                        "",
		Overriddenstatus:              "",
		Port:                          Port{Value: _DEFAULT_INSTNACE_PORT, Enabled: true},
		SecurePort:                    Port{Value: _DEFAULT_INSTNACE_SECURE_PORT, Enabled: false},
		CountryID:                     1,
		DataCenterInfo:                DataCenterInfo{Class: _DEFAULT_DATA_CENTER_INFO_CLASS, Name: _DEFAULT_DATA_CENTER_INFO},
		LeaseInfo:                     LeaseInfo{RenewalIntervalInSecs: _RENEWAL_INTERVAL_IN_SECS, DurationInSecs: _DURATION_IN_SECS},
		Metadata:                      Metadata{ManagementPort: _DEFAULT_INSTNACE_PORT},
		HomePageURL:                   "",
		StatusPageURL:                 "",
		HealthCheckURL:                "",
		VipAddress:                    "",
		SecureVipAddress:              "",
		IsCoordinatingDiscoveryServer: false,
		LastUpdatedTimestamp:          0,
		LastDirtyTimestamp:            0,
		ActionType:                    "",
	}
	for _, o := range opts {
		o(opt)
	}

	return opt
}
func newInstance(opts ...Option) *Instance {
	instance := newOptions(opts...)
	return instance

}

type RegistryClient struct {
	lock         sync.RWMutex
	eurekaSevers []string
	instance     *Instance
	logger       *log.Logger
	done         chan struct{}
	alive        bool
	httpClient   IHttpClient
}

func (client *RegistryClient) Register(status, overriddenstatus instanceStatus) {

	client.callAllEurekaServer(func(url string) error {
		return client.setClientDefault(url)
	})

	client.callAllEurekaServer(func(url string) error {
		return client.httpClient.Register(url, client.instance)
	})

	client.lock.Lock()
	client.alive = true
	client.lock.Unlock()
}
func (client *RegistryClient) setClientDefault(urls string) error {
	u, err := url.Parse(client.eurekaSevers[0])
	if err != nil {
		return err
	}
	if client.instance.Status == "" {
		client.instance.Status = string(INSTANCE_STATUS_UP)
	}
	if client.instance.Overriddenstatus == "" {
		client.instance.Overriddenstatus = string(INSTANCE_STATUS_UNKNOWN)
	}
	if client.instance.IPAddr == "" {
		client.instance.IPAddr = u.Hostname()
	}
	if client.instance.HostName == "" {
		client.instance.HostName = u.Hostname()
	}
	if client.instance.InstanceID == "" {
		client.instance.InstanceID = fmt.Sprintf("%s:%s:%d", client.instance.HostName, client.instance.App, client.instance.Port.Value)
	}
	if client.instance.VipAddress == "" {
		client.instance.VipAddress = strings.ToLower(client.instance.App)
	}
	if client.instance.SecureVipAddress == "" {
		client.instance.SecureVipAddress = strings.ToLower(client.instance.App)
	}
	if client.instance.HomePageURL == "" {
		u.Path = ""
		client.instance.HomePageURL = u.String()
	}
	if client.instance.StatusPageURL == "" {
		u.Path = "info"
		client.instance.StatusPageURL = u.String()
	}
	if client.instance.HealthCheckURL == "" {
		u.Path = "health"
		client.instance.HealthCheckURL = u.String()
	}
	if client.instance.LastDirtyTimestamp == 0 {
		client.instance.LastDirtyTimestamp = time.Now().UnixNano() / 1e6
	}
	if client.instance.LastUpdatedTimestamp == 0 {
		client.instance.LastUpdatedTimestamp = time.Now().UnixNano() / 1e6
	}
	if client.instance.ActionType == "" {
		client.instance.ActionType = string(ACTION_TYPE_ADDED)
	}
	return nil
}
func (client *RegistryClient) callAllEurekaServer(f func(url string) error) {
	for _, url := range client.eurekaSevers {
		err := f(url)
		if err != nil {
			log.Printf("Error Eureka server [%s] is down. %s", url, err.Error())
		}
	}

}
func (client *RegistryClient) sendHeartbeat() {
	client.callAllEurekaServer(func(url string) error {
		return client.httpClient.SendHeartBeat(url, client.instance.App,
			client.instance.InstanceID, client.instance.LastDirtyTimestamp, instanceStatus(client.instance.Status), "")
	})
}
func (client *RegistryClient) heartbeat() {
	ticker := time.NewTicker(time.Duration(client.instance.LeaseInfo.RenewalIntervalInSecs) * time.Second)

	for {
		select {
		case <-ticker.C:
			client.logger.Println("sending heart beat to spring cloud server ")
			client.sendHeartbeat()

		case <-client.done:
			client.logger.Println("stopping client...", len(client.done))
			return
		}
	}
}
func (client *RegistryClient) Start() {
	client.logger.Println("start to registry client...")
	client.Register(INSTANCE_STATUS_UP, INSTANCE_STATUS_UNKNOWN)
	go client.heartbeat()
}
func (client *RegistryClient) Cancel() {

	client.callAllEurekaServer(func(url string) error {
		return client.httpClient.Cancel(url, client.instance.App, client.instance.InstanceID)
	})

	client.lock.Lock()
	client.alive = false
	client.lock.Unlock()
}
func (client *RegistryClient) Stop() {
	client.lock.RLock()
	if client.alive {
		client.lock.RUnlock()
		client.done <- struct{}{}
		close(client.done)
		client.Register(INSTANCE_STATUS_DOWN, "")
		client.Cancel()
	} else {
		client.lock.RUnlock()
		client.logger.Println("client already stop")
	}
}
func NewRegistryClient(eurekaServer []string, appName Option, opts ...Option) *RegistryClient {
	if len(eurekaServer) == 0 {
		panic("eurekaServer 为空")
	}
	opts = append(opts, appName)
	instance := newInstance(opts...)
	return &RegistryClient{
		lock:         sync.RWMutex{},
		eurekaSevers: eurekaServer,
		instance:     instance,
		done:         make(chan struct{}),
		logger:       log.New(os.Stdout, "[eurekaClient] ", log.Lshortfile|log.Ldate|log.Ltime),
		httpClient:   &HttpClient{},
	}
}
