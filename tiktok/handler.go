package tiktok

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"strings"
	"time"
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) error {
	link := utils.TrimURL(utils.RgxTiktok.FindString(update.Message.Text))
	link = utils.SanitizeUrl(link)

	db.DRIVER.LogInformation("Started processing tiktok request " + link + " by " + utils.GetTelegramUserString(update.Message.From))

	id, err := GetId(link)
	if err != nil {
		return err
	}
	parsedId, err := Parse(id)
	if err != nil {
		return err
	}
	data, err := NewAwemeItem(parsedId)
	if err != nil {
		return err
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Request additional info", id)))
	if data.ImagePostInfo.Images == nil {
		file, err := data.DownloadVideo(utils.DownloadBytesLimit)
		if err != nil {
			return err
		}

		video := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FilePath(file.Name()))
		video.ReplyMarkup = keyboard
		video.ReplyToMessageID = update.Message.MessageID
		_, err = api.Send(video)

		defer file.Close()
		defer os.Remove(file.Name())

		if err != nil {
			return err
		}

	} else {
		images, audio, err := data.DownloadImagesWithAudio(utils.DownloadBytesLimit)
		if err != nil {
			return err
		}

		var photos []interface{}
		for _, image := range images {
			photos = append(photos, tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(image.Name())))
		}
		chunks := utils.ChunkSlice(photos, 10)
		for _, chunk := range chunks {
			mediaGroup := tgbotapi.NewMediaGroup(update.Message.Chat.ID, chunk)
			api.Send(mediaGroup)
			time.Sleep(time.Second * 2)
		}

		time.Sleep(time.Second * 1)

		defer audio.Close()
		defer os.Remove(audio.Name())
		defer closeAndDeleteFiles(images)

		name := audio.Name()
		if strings.HasSuffix(name, ".mp4") {
			c := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FilePath(name))
			c.ReplyMarkup = keyboard
			c.ReplyToMessageID = update.Message.MessageID
			_, err = api.Send(c)
		} else {
			c := tgbotapi.NewAudio(update.Message.Chat.ID, tgbotapi.FilePath(name))
			c.ReplyMarkup = keyboard
			c.ReplyToMessageID = update.Message.MessageID
			_, err = api.Send(c)
		}
		if err != nil {
			return err
		}
	}

	db.DRIVER.LogInformation("Finished processing tiktok request by " + utils.GetTelegramUserString(update.Message.From))
	return nil
}

func closeAndDeleteFiles(files []*os.File) {
	for _, file := range files {
		file.Close()
		os.Remove(file.Name())
	}
}
