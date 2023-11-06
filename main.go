package main

import (
	"log"
	"os"

	"github.com/kataras/iris/v12"
)

type Config struct {
	MongoURI string
	RedisURI string
	Host     string
	Port     string
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("%s not set, using default: %s\n", key, defaultValue)
		return defaultValue
	}
	return value
}

func NewConfig() *Config {
	return &Config{
		MongoURI: getEnvWithDefault("MONGO_URI", "mongodb://localhost:27017"),
		RedisURI: getEnvWithDefault("REDIS_URI", "redis://localhost:6379"),
		Host:     getEnvWithDefault("HOST", "0.0.0.0"),
		Port:     getEnvWithDefault("PORT", "8080"),
	}
}

func main() {
	config := NewConfig()
	api := NewAPI(config)

	log.Printf("Listening on %s:%s\n", config.Host, config.Port)

	go api.StartCron()

	api.Run(iris.Addr(config.Host+":"+config.Port), iris.WithoutServerError(iris.ErrServerClosed))
}
