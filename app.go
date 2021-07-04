package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	redis "github.com/go-redis/redis/v8"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
	log "github.com/sirupsen/logrus"
)

var config CloudHealthConfiguration
var ctx = context.Background()

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	configFile, err := filepath.Abs("./config.json")
	if err != nil {
		panic(err.Error())
	}
	value, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(value, &config)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("sup bitch")
}

func main() {
	waitChan := make(chan bool)
	for _, host := range config.Hosts {
		go func(hostinfo HealthCheckHost) {
			redisClient := redis.NewClient(&redis.Options{
				Addr:     os.Getenv("REDIS_ADDR"),
				Password: os.Getenv("REDIS_PASSWORD"),
				DB:       0,
			})
			for {
				_, err := redisClient.Get(ctx, hostinfo.Url).Result()
				if err == nil {
					time.Sleep(time.Second * 1)
					continue
				}
				redisClient.Set(ctx, hostinfo.Url, "locked", time.Second*time.Duration(hostinfo.WaitTime))

				resp, err := http.Get(hostinfo.Url)
				if err != nil {
					pl, _ := json.Marshal(HealthCheckFailResponse{
						Url:       hostinfo.Url,
						Timestamp: time.Now().String(),
					})
					_, err = http.Post(hostinfo.Webhook, "application/json", bytes.NewBuffer(pl))
					if err != nil {
						log.Printf("Error calling webhook %s: %s\n", hostinfo.Webhook, err.Error())
					}
					continue
				}

				if resp.StatusCode != 200 {
					pl, _ := json.Marshal(HealthCheckFailResponse{
						Url:       hostinfo.Url,
						Timestamp: time.Now().String(),
					})
					_, err = http.Post(hostinfo.Webhook, "application/json", bytes.NewBuffer(pl))
					if err != nil {
						log.Printf("Error calling webhook %s: %s\n", hostinfo.Webhook, err.Error())
					}
					continue
				}
			}
		}(host)
	}
	<-waitChan
}

type CloudHealthConfiguration struct {
	Hosts []HealthCheckHost `json:"hosts"`
}

type HealthCheckHost struct {
	Url      string `json:"hostname"`
	Webhook  string `json:"onFailWebhook"`
	WaitTime int    `json:"waitTime"`
}

type HealthCheckFailResponse struct {
	Url       string `json:"hostname"`
	Timestamp string `json:"timestamp"`
}
