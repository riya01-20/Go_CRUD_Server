package config

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db *gorm.DB
	rdb *redis.Client  
	ctx = context.Background();
)

func Connect() {
	dsn := "root:Riya@123@tcp(127.0.0.1:3306)/go_gin_crud?parseTime=true"
	d, err := gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db = d

	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password
		DB:       0,                // Use default DB
	})

	//pinging redis to check connection 
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	} else {
		fmt.Println("Redis connected")
	}
}

func GetDB() *gorm.DB {
	fmt.Println("..DB connected Successfully....")
	return db
}

func GetRedis() *redis.Client {
	fmt.Println("..Redis connected Successfully....")
	return rdb
}