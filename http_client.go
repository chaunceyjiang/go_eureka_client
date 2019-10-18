package go_eureka_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

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
	app:=&Application{*instance}
	data, err := json.Marshal(app)
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
