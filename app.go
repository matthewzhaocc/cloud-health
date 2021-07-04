package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"fmt"
	"time"
	"context"
    redis "github.com/go-redis/redis/v8"
)

var config CloudHealthConfiguration
var ctx = context.Background()
func init() {
	configFile, err := os.Open("./config.json")
	if err != nil {}
	value, _ := ioutil.ReadAll(configFile)

	json.Unmarshal(value, &config)

	
}

func main() {
	for _, host := range config.Hosts {
		go func (hostinfo HealthCheckHost) {
			redisClient := redis.NewClient(&redis.Options{
				Addr: os.Getenv("REDIS_ADDR"),
				Password: os.Getenv("REDIS_PASSWORD"),
				DB: 0,
			})
			for {
				_, err := redisClient.Get(ctx, hostinfo.Url).Result()
				if (err == nil) {
					time.Sleep(time.Second * 1)
					fmt.Println(err)
					continue
				}

				redisClient.Set(ctx, hostinfo.Url, "locked", time.Second * 5)
				resp, err := http.Get(hostinfo.Url)
				if (err != nil) {
					http.Post(hostinfo.Webhook, "application/json", bytes.NewBuffer([]byte("{\"url\": \""+hostinfo.Url+"\"}")))
					continue
				}
				
				if (resp.StatusCode != 200) {
					http.Post(hostinfo.Webhook, "application/json", bytes.NewBuffer([]byte("{\"url\": \""+hostinfo.Url+"\"}")))
					continue
				}
			}
		}(host)
	}
	for {}
}

type CloudHealthConfiguration struct {
	Hosts []HealthCheckHost `json:"hosts"`
}

type HealthCheckHost struct {
	Url string `json:"hostname"`
	Webhook string `json:"onFailWebhook"`
}