package go_eureka_client

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParseUrl(t *testing.T)  {

	u,_:=url.Parse("http://127.0.0.1:8761/eureka/")

	fmt.Println(u)

}

func TestNewRegistryClient(t *testing.T) {
	client:= NewRegistryClient([]string{"http://127.0.0.1:8761/eureka/"},AppName("go.eureka.example"))
	client.Start()
	client.Stop()
}