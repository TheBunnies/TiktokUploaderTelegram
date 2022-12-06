package tiktok

import (
	"fmt"
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"regexp"
)

var (
	rgxTiktok = regexp.MustCompile(`http(s|):\/\/.*(tiktok)\.com.*`)
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) {
	link := utils.TrimURL(rgxTiktok.FindString(update.Message.Text))
	link = utils.SanitizeTiktokUrl(link)

	db.DRIVER.LogInformation("Started processing tiktok request " + link + " by " + utils.GetTelegramUserString(update.Message.From))

	id, err := GetId(link)
	if err != nil {
		db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
		msg.ReplyToMessageID = update.Message.MessageID
		api.Send(msg)
		return
	}
	parsedId, err := Parse(id)
	if err != nil {
		db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
		msg.ReplyToMessageID = update.Message.MessageID
		api.Send(msg)
		return
	}
	data, err := NewAwemeItem(parsedId)
	if err != nil {
		db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
		msg.ReplyToMessageID = update.Message.MessageID
		api.Send(msg)
		return
	}
	if data.ImagePostInfo.Images == nil {
		file, err := data.DownloadVideo(utils.DownloadBytesLimit)
		if err != nil {
			if err.Error() == "too large" {
				db.DRIVER.LogInformation("A requested video exceeded it's upload limit for " + utils.GetTelegramUserString(update.Message.From))
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your requested tiktok video is too large for me to handle! I can only upload videos up to 50MB")
				msg.ReplyToMessageID = update.Message.MessageID
				api.Send(msg)
				return
			}
			db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err.Error())
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
			msg.ReplyToMessageID = update.Message.MessageID
			api.Send(msg)
			return
		}
		message := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Author: %s \nDuration: %s\nCreation time: %s \nDescription: %s \n",
			data.Author.Nickname,
			data.Duration(),
			data.Time(),
			data.Description()))
		message.ReplyToMessageID = update.Message.MessageID

		api.Send(message)

		video := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FilePath(file.Name()))

		_, err = api.Send(video)
		if err != nil {
			file.Close()
			os.Remove(file.Name())
			db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err.Error())
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
			msg.ReplyToMessageID = update.Message.MessageID
			api.Send(msg)
			return
		}

		file.Close()
		os.Remove(file.Name())
	} else {
		images, audio, err := data.DownloadImagesWithAudio(utils.DownloadBytesLimit)
		if err != nil {
			if err.Error() == "too large" {
				db.DRIVER.LogInformation("A requested video exceeded it's upload limit for " + utils.GetTelegramUserString(update.Message.From))
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your requested tiktok video is too large for me to handle! I can only upload videos up to 50MB")
				msg.ReplyToMessageID = update.Message.MessageID
				api.Send(msg)
				return
			}
			db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err.Error())
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
			msg.ReplyToMessageID = update.Message.MessageID
			api.Send(msg)
			return
		}
		message := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Author: %s \nDuration: %s\nCreation time: %s \nDescription: %s \n",
			data.Author.Nickname,
			data.Duration(),
			data.Time(),
			data.Description()))
		message.ReplyToMessageID = update.Message.MessageID

		api.Send(message)

		var photos []interface{}
		for _, image := range images {
			photos = append(photos, tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(image.Name())))
		}
		chuncks := utils.ChunkSlice(photos, 10)
		for _, chunck := range chuncks {
			mediaGroup := tgbotapi.NewMediaGroup(update.Message.Chat.ID, chunck)
			api.Send(mediaGroup)
		}

		video := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FilePath(audio.Name()))
		_, err = api.Send(video)
		if err != nil {
			audio.Close()
			os.Remove(audio.Name())
			closeAndDeleteFiles(images)
			db.DRIVER.LogError("Couldn't handle a tiktok request", utils.GetTelegramUserString(update.Message.From), err.Error())
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
			msg.ReplyToMessageID = update.Message.MessageID
			api.Send(msg)
			return
		}

		audio.Close()
		os.Remove(audio.Name())
		closeAndDeleteFiles(images)
	}

	db.DRIVER.LogInformation("Finished processing tiktok request by " + utils.GetTelegramUserString(update.Message.From))
}

func closeAndDeleteFiles(files []*os.File) {
	for _, file := range files {
		file.Close()
		os.Remove(file.Name())
	}
}
