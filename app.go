package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var config CloudHealthConfiguration

func init() {
	configFile, err := os.Open("./config.json")
	if err != nil {}
	value, _ := ioutil.ReadAll(configFile)

	json.Unmarshal(value, &config)
}

func main() {
	for {
		for _, host := range config.Hosts {
			go func (hostinfo HealthCheckHost) {
				resp, err := http.Get(hostinfo.Url)
				if (err != nil) {
					http.Post(hostinfo.Webhook, "application/json", bytes.NewBuffer([]byte("{\"url\": \""+hostinfo.Url+"\"}")))
					return
				}
				
				if (resp.StatusCode != 200) {
					http.Post(hostinfo.Webhook, "application/json", bytes.NewBuffer([]byte("{\"url\": \""+hostinfo.Url+"\"}")))
					return
				}
			}(host)

			
		}
		time.Sleep(time.Duration(time.Second * 300))
	}
}

type CloudHealthConfiguration struct {
	Hosts []HealthCheckHost `json:"hosts"`
}

type HealthCheckHost struct {
	Url string `json:"hostname"`
	Webhook string `json:"onFailWebhook"`
}