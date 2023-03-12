package main

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/bot"
	"github.com/TheBunnies/TiktokUploaderTelegram/config"
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
)

func main() {
	config.LoadEnv()
	db.Setup()
	bot.InitBot()
}
