package go_eureka_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Port struct {
	Attr    int  `json:"$"`
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
	Attr           int `json:"$"`
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

type Option func(*Instance)

func newOptions(opts ...Option) *Instance {
	opt := &Instance{
		InstanceID:                    "",
		HostName:                      "",
		App:                           "",
		IPAddr:                        "",
		Status:                        "",
		Overriddenstatus:              "",
		Port:                          Port{Attr: _DEFAULT_INSTNACE_PORT, Enabled: true},
		SecurePort:                    Port{Attr: _DEFAULT_INSTNACE_SECURE_PORT, Enabled: false},
		CountryID:                     1,
		DataCenterInfo:                DataCenterInfo{Class: _DEFAULT_DATA_CENTER_INFO_CLASS, Name: _DEFAULT_DATA_CENTER_INFO},
		LeaseInfo:                     LeaseInfo{RenewalIntervalInSecs: _RENEWAL_INTERVAL_IN_SECS, DurationInSecs: _DURATION_IN_SECS},
		Metadata:                      Metadata{Attr: _DEFAULT_INSTNACE_PORT},
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

	if opt.HomePageURL == "" {

	}
	if opt.StatusPageURL == "" {

	}

	if opt.HealthCheckURL == "" {

	}
	if opt.VipAddress == "" {
		opt.VipAddress = strings.ToLower(opt.App)
	}
	if opt.SecureVipAddress == "" {
		opt.SecureVipAddress = strings.ToLower(opt.App)
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
	if status == "" {
		status = INSTANCE_STATUS_UP
	}
	if overriddenstatus == "" {
		overriddenstatus = INSTANCE_STATUS_UNKNOWN
	}
	client.callAllEurekaServer(func(url string) error {
		return client.httpClient.Register(url, client.instance)
	})
	client.instance.Status = string(status)
	client.instance.Overriddenstatus = string(overriddenstatus)
	client.lock.Lock()
	client.alive = true
	client.lock.Unlock()
}

func (client *RegistryClient) callAllEurekaServer(f func(url string) error)  {
	for _, url := range client.eurekaSevers {
		err := f(url)
		if err != nil {
			log.Printf("Error Eureka server [%s] is down. %s", url, err.Error())
		}
	}

}
func (client *RegistryClient) sendHeartbeat() {

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

type HttpClient struct {
}

type IHttpClient interface {
	Register(eurekaServer string, instance *Instance) error
	Cancel(rawUrl, appName, instanceId string) error
	SendHeartBeat(eurekaServer string, appName string, instanceId string, lastDirtyTimestamp int64, status instanceStatus, overriddenstatus instanceStatus) error
	StatusUpdate(eurekaServer string, appName string, instanceId string, lastDirtyTimestamp int64, status instanceStatus) error
	DeleteStatusOverride(eurekaServer string, appName string, instanceId string, lastDirtyTimestamp int64) error
}

func (h *HttpClient) Register(eurekaServer string, instance *Instance) error {
	client := &http.Client{Timeout: time.Duration(_DEFAULT_TIME_OUT) * time.Second}
	data, err := json.Marshal(instance)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", eurekaServer+fmt.Sprintf("apps/%s", instance.App), reader)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-encoding", "gzip")
	_, err = client.Do(req)
	return err
}

func (h *HttpClient) Cancel(eurekaServer, appName, instanceId string) error {
	//u, err := url.Parse(rawUrl)
	//if err != nil {
	//	return err
	//}
	client := &http.Client{Timeout: time.Duration(_DEFAULT_TIME_OUT) * time.Second}

	req, err := http.NewRequest("DELETE", eurekaServer+fmt.Sprintf("apps/%s/%s", appName, instanceId), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept-encoding", "gzip")
	_, err = client.Do(req)
	return err
}
func (h *HttpClient) SendHeartBeat(eurekaServer string, appName string, instanceId string, lastDirtyTimestamp int64, status instanceStatus, overriddenstatus instanceStatus) error {
	heartBeatUrl := eurekaServer + fmt.Sprintf("apps/%s/%s?status=%s&lastDirtyTimestamp=%d", appName, instanceId, status, lastDirtyTimestamp)
	if overriddenstatus != "" {
		heartBeatUrl += fmt.Sprintf("&overriddenstatus=%s", overriddenstatus)
	}
	log.Println("heartBeatUrl " + heartBeatUrl)

	client := &http.Client{Timeout: time.Duration(_DEFAULT_TIME_OUT) * time.Second}
	req, err := http.NewRequest("PUT", heartBeatUrl, nil)
	req.Header.Add("Accept-encoding", "gzip")
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	return err
}
func (h *HttpClient) StatusUpdate(eurekaServer string, appName string, instanceId string, lastDirtyTimestamp int64, status instanceStatus) error {
	statusUrl := eurekaServer + fmt.Sprintf("apps/%s/%s?status=%s&lastDirtyTimestamp=%d", appName, instanceId, status, lastDirtyTimestamp)
	log.Println("statusUrl ", statusUrl)
	client := http.Client{Timeout: time.Duration(_DEFAULT_TIME_OUT) * time.Second}

	req, err := http.NewRequest("PUT", statusUrl, nil)
	req.Header.Add("Accept-encoding", "gzip")
	if err != nil {
		return nil
	}

	_, err = client.Do(req)
	return err
}

func (h HttpClient) DeleteStatusOverride(eurekaServer string, appName string, instanceId string, lastDirtyTimestamp int64) error {
	delUrl := eurekaServer + fmt.Sprintf("apps/%s/%s/status?lastDirtyTimestamp=%d", appName, instanceId, lastDirtyTimestamp)
	log.Println("DeleteStatusOverride ", delUrl)
	client := &http.Client{Timeout: time.Duration(_DEFAULT_TIME_OUT) * time.Second}
	req, err := http.NewRequest("DELETE", delUrl, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	return err
}
