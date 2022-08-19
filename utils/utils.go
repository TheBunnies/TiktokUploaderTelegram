package utils

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/url"
	"strings"
)

func FileNameWithoutExtension(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}

func TrimURL(uri string) string {
	loc, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	loc.RawQuery = ""
	loc.Scheme = "https"
	return loc.String()
}

func GetTelegramUserString(user *tgbotapi.User) string {
	return fmt.Sprintf("[%d] - %s %s (%s)", user.ID, user.FirstName, user.LastName, user.UserName)
}

func SanitizeTiktokUrl(url string) string {
	return strings.Split(url, "%20")[0]
}
