package youtube

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) error {
	link := utils.SanitizeUrl(utils.RgxYoutube.FindString(update.Message.Text))

	db.DRIVER.LogInformation("Started processing youtube request " + link + " by " + utils.GetTelegramUserString(update.Message.From))

	file, err := DownloadVideo(link)
	if err != nil {
		return err
	}
	video := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FilePath(file.Name()))
	video.ReplyToMessageID = update.Message.MessageID
	_, err = api.Send(video)

	defer file.Close()
	defer os.Remove(file.Name())

	db.DRIVER.LogInformation("Finished processing youtube request by " + utils.GetTelegramUserString(update.Message.From))

	return nil
}
