package bot

import (
	"fmt"
	"github.com/TheBunnies/TiktokUploaderTelegram/config"
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
	"github.com/TheBunnies/TiktokUploaderTelegram/tiktok"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	"github.com/TheBunnies/TiktokUploaderTelegram/youtube"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

var (
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
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}
		go func(upd tgbotapi.Update) {
			if upd.CallbackQuery != nil {
				if upd.CallbackQuery.Message.Caption != "" {
					return
				}
				parsedId, err := tiktok.Parse(upd.CallbackQuery.Data)
				if err != nil {
					db.DRIVER.LogError("Couldn't handle a callback request", utils.GetTelegramUserString(upd.CallbackQuery.From), err.Error())
					return
				}
				data, err := tiktok.NewAwemeItem(parsedId)
				if err != nil {
					db.DRIVER.LogError("Couldn't handle a callback request", utils.GetTelegramUserString(upd.CallbackQuery.From), err.Error())
					return
				}

				msg := tgbotapi.NewEditMessageCaption(upd.CallbackQuery.Message.Chat.ID, upd.CallbackQuery.Message.MessageID, fmt.Sprintf("Author: %s \nDuration: %s\nCreation time: %s \nDescription: %s \n",
					data.Author.Nickname,
					data.Duration(),
					data.Time(),
					data.Description()))
				if _, err := bot.Send(msg); err != nil {
					db.DRIVER.LogError("Couldn't handle a callback request", utils.GetTelegramUserString(upd.CallbackQuery.From), err.Error())
				}
				return
			}

			if upd.Message.Chat.IsPrivate() && (strings.HasPrefix(upd.Message.Text, "/help") || strings.HasPrefix(upd.Message.Text, "/start")) {
				db.DRIVER.LogInformation(utils.GetTelegramUserString(upd.Message.From), "just invoked the /start or /help command")
				err = TryCreateUser(upd.Message.From)
				if err != nil {
					db.DRIVER.LogError("Error while creating a user", err.Error())
				}
				msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Hello! Start using me by just typing either tiktok or twitter URL in whatever chat I'm in :)")
				msg.ReplyToMessageID = upd.Message.MessageID
				bot.Send(msg)
				return
			}
			if utils.RgxTiktok.MatchString(upd.Message.Text) {
				err = TryCreateUser(upd.Message.From)
				if err != nil {
					db.DRIVER.LogError("Error while creating a user", err.Error())
				}
				action := tgbotapi.NewChatAction(upd.Message.Chat.ID, tgbotapi.ChatTyping)
				bot.Send(action)
				err = tiktok.Handle(upd, bot)
				if err != nil {
					if err.Error() == "too large" {
						db.DRIVER.LogInformation("A requested video exceeded it's upload limit for " + utils.GetTelegramUserString(upd.Message.From))
						msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Your requested tiktok video is too large for me to handle! I can only upload videos up to 50MB")
						msg.ReplyToMessageID = upd.Message.MessageID
						bot.Send(msg)
						return
					}
					db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(upd.Message.From), err.Error())
					msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
					msg.ReplyToMessageID = upd.Message.MessageID
					bot.Send(msg)
				}
			}
			if utils.RgxYoutube.MatchString(upd.Message.Text) {
				err = TryCreateUser(upd.Message.From)
				if err != nil {
					db.DRIVER.LogError("Error while creating a user", err.Error())
				}
				action := tgbotapi.NewChatAction(upd.Message.Chat.ID, tgbotapi.ChatTyping)
				bot.Send(action)
				err = youtube.Handle(upd, bot)
				if err != nil {
					db.DRIVER.LogError("Couldn't handle a youtube request", utils.GetTelegramUserString(upd.Message.From), err.Error())
					msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
					msg.ReplyToMessageID = upd.Message.MessageID
					bot.Send(msg)
				}
			}
		}(update)
	}
}

func TryCreateUser(user *tgbotapi.User) error {
	exists, err := db.DRIVER.IsUserExists(user.ID)
	if err != nil {
		return nil
	}
	if exists {
		dbUser, _ := db.DRIVER.GetUser(user.ID)
		return db.DRIVER.UpdateUser(*dbUser, db.User{FirstName: user.FirstName, LastName: user.LastName, Username: user.UserName})
	} else {
		return db.DRIVER.CreateUser(user.ID, user.FirstName, user.LastName, user.UserName)
	}

}
