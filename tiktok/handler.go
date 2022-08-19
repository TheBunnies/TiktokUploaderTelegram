package tiktok

import (
	"fmt"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"regexp"
)

var (
	rgxTiktok = regexp.MustCompile(`http(s|):\/\/.*(tiktok)\.com.*`)
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) error {
	if rgxTiktok.MatchString(update.Message.Text) {
		link := utils.TrimURL(update.Message.Text)
		link = utils.SanitizeTiktokUrl(link)

		log.Println("Started processing tiktok request by " + utils.GetTelegramUserString(update.Message.From))

		id, err := GetId(link)
		if err != nil {
			return err
		}
		parsedId, err := Parse(id)
		if err != nil {
			return err
		}
		data, err := NewAwemeDetail(parsedId)
		if err != nil {
			return err
		}
		file, err := data.DownloadVideo()
		if err != nil {
			return err
		}
		message := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Author: %s \nDuration: %s\nCreation time: %s \nDescription: %s \n",
			data.Author,
			data.Duration(),
			data.Time(),
			data.Description()))
		message.ReplyToMessageID = update.Message.MessageID

		api.Send(message)

		media := tgbotapi.FilePath(file.Name())
		video := tgbotapi.NewVideo(update.Message.Chat.ID, media)

		_, err = api.Send(video)
		if err != nil {
			file.Close()
			os.Remove(file.Name())
			return err
		}

		file.Close()
		os.Remove(file.Name())
		log.Println("Finished processing tiktok request by " + utils.GetTelegramUserString(update.Message.From))
	}
	return nil
}
