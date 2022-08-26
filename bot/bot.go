package bot

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/config"
	"github.com/TheBunnies/TiktokUploaderTelegram/tiktok"
	"github.com/TheBunnies/TiktokUploaderTelegram/twitter"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"regexp"
	"strings"
)

var (
	rgxTiktok  = regexp.MustCompile(`http(s|):\/\/.*(tiktok).com[^\s]*`)
	rgxTwitter = regexp.MustCompile(`http(|s):\/\/twitter\.com\/i\/status\/[0-9]*`)

	Client *tgbotapi.BotAPI
)

func InitBot() {
	config.LoadEnv()
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}
	Client = bot
	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			go func() {
				if update.Message.Chat.IsPrivate() && (strings.HasPrefix(update.Message.Text, "/help") || strings.HasPrefix(update.Message.Text, "/start")) {
					log.Println(utils.GetTelegramUserString(update.Message.From), "just invoked the /start or /help command")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! Start using me by just typing either tiktok or twitter URL in whatever chat I'm in :)")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					return
				}

				err = twitter.Handle(update, bot)
				if err != nil {
					log.Println("Couldn't handle a twitter request", utils.GetTelegramUserString(update.Message.From), err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					return
				}
				err = tiktok.Handle(update, bot)
				if err != nil {
					log.Println("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					return
				}
			}()
		}
	}
}
