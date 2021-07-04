package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	redis "github.com/go-redis/redis/v8"
	fiber "github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"os"
	"time"
)

var ctx = context.Background()
var db *gorm.DB

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	db, _ = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	db.AutoMigrate(&HealthCheckHost{})
	/**db.Create(&HealthCheckHost{
		Url: "https://googlexxxxxxxxxxxxxxxxxxxxxxxx.com",
		Webhook: "http://localhost:3000/webhook",
		WaitTime: 3,
	})*/

	fmt.Println("sup bitch")
}

func main() {
	app := fiber.New()
	app.Post("/new", func(c *fiber.Ctx) error {
		host := new(HealthCheckHost)
		if err := c.BodyParser(host); err != nil {
			return err
		}
		db.Create(host)
		c.SendString("success")
		return nil
	})
	go app.Listen(":" + os.Getenv("PORT"))
	for {
		var hosts []HealthCheckHost
		db.Find(&hosts)
	
		for _, host := range hosts {
			go HealthCheckFunc(host)
		}
		time.Sleep(time.Second)
	}
	
}

func HealthCheckFunc(hostinfo HealthCheckHost) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	defer redisClient.Close()
	_, err := redisClient.Get(ctx, hostinfo.Url).Result()
	if err == nil {
		return
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
		return
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
		return
	}
}

type HealthCheckHost struct {
	gorm.Model
	Url      string `json:"hostname"`
	Webhook  string `json:"onFailWebhook"`
	WaitTime int    `json:"waitTime"`
}

type HealthCheckFailResponse struct {
	Url       string `json:"hostname"`
	Timestamp string `json:"timestamp"`
}
