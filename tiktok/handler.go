package tiktok

import (
	"fmt"
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	rgxTiktok = regexp.MustCompile(`http(s|):\/\/.*(tiktok)\.com.*`)
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) error {
	link := utils.TrimURL(rgxTiktok.FindString(update.Message.Text))
	link = utils.SanitizeTiktokUrl(link)

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
	if data.ImagePostInfo.Images == nil {
		file, err := data.DownloadVideo(utils.DownloadBytesLimit)
		if err != nil {
			return err
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
			return err
		}

		file.Close()
		os.Remove(file.Name())
	} else {
		images, audio, err := data.DownloadImagesWithAudio(utils.DownloadBytesLimit)
		if err != nil {
			return err
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
			time.Sleep(time.Second * 2)
		}
		var c tgbotapi.Chattable
		name := audio.Name()
		if strings.HasSuffix(name, ".mp4") {
			c = tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FilePath(name))
		} else {
			c = tgbotapi.NewAudio(update.Message.Chat.ID, tgbotapi.FilePath(audio.Name()))
		}
		time.Sleep(time.Second * 1)
		_, err = api.Send(c)
		if err != nil {
			audio.Close()
			os.Remove(audio.Name())
			closeAndDeleteFiles(images)
			return err
		}

		audio.Close()
		os.Remove(audio.Name())
		closeAndDeleteFiles(images)
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
