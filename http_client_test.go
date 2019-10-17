package go_eureka_client

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

func TestParseUrl(t *testing.T) {

	u, _ := url.Parse("http://127.0.0.1:8761/eureka/")

	fmt.Println(u)
	fmt.Println(time.Now().UnixNano() / 1e6)
}

func TestNewRegistryClient(t *testing.T) {
	client := NewRegistryClient([]string{"http://localhost:8761/eureka/"}, AppName("go"), InstanceHost("192.168.5.221"), InstancePort(9090))
	client.Start()
	//client.Stop()
	time.Sleep(60 * time.Second)
	client.Stop()
}
