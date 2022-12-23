package twitter

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"regexp"
)

var (
	rgxTwitter = regexp.MustCompile(`http(|s):\/\/twitter\.com\/i\/status\/[0-9]*`)
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) error {
	link := utils.TrimURL(rgxTwitter.FindString(update.Message.Text))
	db.DRIVER.LogInformation("Started processing twitter request " + link + " by " + utils.GetTelegramUserString(update.Message.From))

	data := NewTwitterVideoDownloader(link)
	file, err := data.Download(utils.DownloadBytesLimit)
	if err != nil {
		return err
	}
	media := tgbotapi.FilePath(file.Name())
	video := tgbotapi.NewVideo(update.Message.Chat.ID, media)
	video.ReplyToMessageID = update.Message.MessageID

	_, err = api.Send(video)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return err
	}

	file.Close()
	os.Remove(file.Name())

	db.DRIVER.LogInformation("Finished processing twitter request by " + utils.GetTelegramUserString(update.Message.From))
	return nil
}
