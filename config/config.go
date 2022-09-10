package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var (
	Token    = ""
	User     = ""
	Password = ""
	Ip       = ""
	Port     = ""
	ConStr   = ""

	ProxyUrl = ""
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Token = os.Getenv("TOKEN")
	User = os.Getenv("USER")
	Password = os.Getenv("PASSWORD")
	Ip = os.Getenv("IP")
	Port = os.Getenv("PORT")
	ConStr = os.Getenv("CONSTR")

	if User != "" && Password != "" && Ip != "" && Port != "" {
		ProxyUrl = "http://" + User + ":" + Password + "@" + Ip + ":" + Port
	}
}
