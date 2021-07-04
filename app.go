package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	redis "github.com/go-redis/redis/v8"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var config CloudHealthConfiguration
var ctx = context.Background()
func init() {
	configFile, err := filepath.Abs("./config.json")
	if err != nil {
		panic(err.Error())
	}
	value, err := ioutil.ReadFile(configFile)

	err = json.Unmarshal(value, &config)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("sup bitch")
}

func main() {
	waitChan := make(chan bool)
	for _, host := range config.Hosts {
		go func (hostinfo HealthCheckHost) {
			redisClient := redis.NewClient(&redis.Options{
				Addr: os.Getenv("REDIS_ADDR"),
				Password: os.Getenv("REDIS_PASSWORD"),
				DB: 0,
			})
			fmt.Println(redisClient)
			for {
				_, err := redisClient.Get(ctx, hostinfo.Url).Result()
				if err != nil {
					time.Sleep(time.Second * 1)
					fmt.Println(err.Error())
					continue
				}

				redisClient.Set(ctx, hostinfo.Url, "locked", time.Second * 5)
				resp, err := http.Get(hostinfo.Url)
				if err != nil {
					log.Printf("Error reaching HTTP endpoint: %s\n", err.Error())
					_, err = http.Post(hostinfo.Webhook, "application/json", bytes.NewBuffer([]byte("{\"url\": \""+hostinfo.Url+"\"}")))
					if err != nil {
						log.Printf("Error calling webhook %s: %s\n", hostinfo.Webhook, err.Error())
						continue
					}
					continue
				}

				if resp.StatusCode != 200 {
					_, err = http.Post(hostinfo.Webhook, "application/json", bytes.NewBuffer([]byte("{\"url\": \""+hostinfo.Url+"\"}")))
					if err != nil {
						log.Printf("Invalid response code: %s", resp.Status)
						log.Printf("Error calling webhook %s: %s\n", hostinfo.Webhook, err.Error())
					}
					continue
				}
			}
		}(host)
	}
	<- waitChan
}

type CloudHealthConfiguration struct {
	Hosts []HealthCheckHost `json:"hosts"`
}

type HealthCheckHost struct {
	Url string `json:"hostname"`
	Webhook string `json:"onFailWebhook"`
}
