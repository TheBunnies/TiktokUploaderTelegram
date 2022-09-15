package ttvideo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strings"
	"time"
)

const Origin = "https://tik-tok-video.com"

func NewTTVideoDetail(url string) (*TTVideoDetail, error) {
	req, err := http.NewRequest("POST", Origin+"/api/convert", strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, url)))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	var detail TTVideoDetail
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (a TTVideoDetail) Duration() string {
	return a.Meta.Duration
}

func (a TTVideoDetail) Author() string {
	return strings.Split(a.Meta.Title, " -")[0]
}
func (a TTVideoDetail) URL() (string, error) {
	if len(a.Url) == 0 {
		return "", errors.New("invalid slice")
	}
	return a.Url[0].Url, nil
}

func (a TTVideoDetail) DownloadVideo() (*os.File, error) {
	url, err := a.URL()
	if err != nil {
		return nil, err
	}
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("%s.%s", u.String(), strings.Split(res.Header.Get("Content-Type"), "/")[1])
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if _, err := file.ReadFrom(res.Body); err != nil {
		return nil, err
	}
	openedFile, err := os.Open(file.Name())
	if err != nil {
		openedFile.Close()
		os.Remove(openedFile.Name())
		return nil, err
	}
	return openedFile, nil
}
