package utils

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	DownloadBytesLimit = 52428800
)

var (
	RgxTiktok  = regexp.MustCompile(`http(s|):\/\/.*(tiktok).com[^\s]*`)
	RgxYoutube = regexp.MustCompile(`http(s|):\/\/(www.|)youtube.com\/shorts\/.*`)
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

func SanitizeUrl(url string) string {
	return strings.Split(url, "%20")[0]
}

func RandomString(r int) string {
	var sb strings.Builder
	slice := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	rand.Seed(time.Now().Unix())
	for i := 0; i < r; i++ {
		rng := slice[rand.Intn(len(slice))]
		sb.Write([]byte(rng))
	}
	return sb.String()
}

func RandomDigits(r int) string {
	var sb strings.Builder
	slice := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	rand.Seed(time.Now().Unix())
	for i := 0; i < r; i++ {
		rng := slice[rand.Intn(len(slice))]
		sb.Write([]byte(rng))
	}
	return sb.String()
}

func ChunkSlice(slice []interface{}, chunkSize int) [][]interface{} {
	var chunks [][]interface{}
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}
