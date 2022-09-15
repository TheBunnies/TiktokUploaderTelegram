package bot

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/config"
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
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
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}
	Client = bot
	bot.Debug = false

	db.DRIVER.LogInformation("Authorized on account", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Chat.IsPrivate() && (strings.HasPrefix(update.Message.Text, "/help") || strings.HasPrefix(update.Message.Text, "/start")) {
			db.DRIVER.LogInformation(utils.GetTelegramUserString(update.Message.From), "just invoked the /start or /help command")
			err = TryCreateUser(update.Message.From)
			if err != nil {
				db.DRIVER.LogError("Error while creating a user", err.Error())
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! Start using me by just typing either tiktok or twitter URL in whatever chat I'm in :)")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue
		}
		if rgxTwitter.MatchString(update.Message.Text) {
			go func() {
				err = TryCreateUser(update.Message.From)
				if err != nil {
					db.DRIVER.LogError("Error while creating a user", err.Error())
				}
				err = twitter.Handle(update, bot)
				if err != nil {
					db.DRIVER.LogError("Couldn't handle a twitter request", utils.GetTelegramUserString(update.Message.From), err.Error())
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					return
				}
			}()
		}
		if rgxTiktok.MatchString(update.Message.Text) {
			go func() {
				err = TryCreateUser(update.Message.From)
				if err != nil {
					db.DRIVER.LogError("Error while creating a user", err.Error())
				}
				err = tiktok.Handle(update, bot)
				if err != nil {
					db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err.Error())
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Tiktok requests are currently unavailable in the bot, we're investigating the issue and will inform you when it's finally going back. Thank you for your patience!")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					return
				}
			}()
		}
	}
}

func TryCreateUser(user *tgbotapi.User) error {
	exists, err := db.DRIVER.IsUserExists(user.ID)
	if err != nil {
		return err
	}
	if !exists {
		return db.DRIVER.CreateUser(user.ID, user.FirstName, user.LastName, user.UserName)
	}
	return nil
}
