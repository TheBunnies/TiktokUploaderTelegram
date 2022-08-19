package twitter

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"regexp"
)

var (
	rgxTwitter = regexp.MustCompile(`http(|s):\/\/twitter\.com\/i\/status\/[0-9]*`)
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) error {
	if rgxTwitter.MatchString(update.Message.Text) {
		link := utils.TrimURL(rgxTwitter.FindString(update.Message.Text))
		log.Println("Started processing twitter request by " + utils.GetTelegramUserString(update.Message.From))

		data := NewTwitterVideoDownloader(link)
		file, err := data.Download()
		if err != nil {
			return err
		}
		media := tgbotapi.FilePath(file.Name())
		video := tgbotapi.NewVideo(update.Message.From.ID, media)
		video.ReplyToMessageID = update.Message.MessageID

		_, err = api.Send(video)
		if err != nil {
			file.Close()
			os.Remove(file.Name())
			return err
		}

		file.Close()
		os.Remove(file.Name())

		log.Println("Finished processing twitter request by " + utils.GetTelegramUserString(update.Message.From))
	} else {

	}
	return nil
}
