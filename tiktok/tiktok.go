package tiktok

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TheBunnies/TiktokUploaderTelegram/config"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	"github.com/google/uuid"
)

const Origin = "https://api16-normal-c-useast1a.tiktokv.com"

func Parse(id string) (uint64, error) {
	return strconv.ParseUint(id, 10, 64)
}

func NewAwemeItem(id uint64) (*AwemeItem, error) {
	req, err := http.NewRequest("GET", Origin+"/aweme/v1/feed/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36")
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("aweme_id", strconv.FormatUint(id, 10))
	q.Add("os_version", "9")
	q.Add("device_type", "ASUS_Z01QD")
	q.Add("device_platform", "android")
	q.Add("version_code", "300904")
	q.Add("app_name", "musical_ly")
	q.Add("channel", "googleplay")
	q.Add("device_id", "7318517321748022790")
	q.Add("iid", "7318518857994389254")

	req.URL.RawQuery = q.Encode()
	cookieOdin := &http.Cookie{
		Name:   "odin_tt",
		Value:  utils.RandomString(160),
		MaxAge: 300,
	}
	req.AddCookie(cookieOdin)

	var transport *http.Transport
	proxyUrl, err := url.Parse(config.ProxyUrl)
	if err != nil {
		transport = nil
	} else {
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	var detail AwemeDetail
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return nil, err
	}
	return &detail.AwemeList[0], nil
}

func (a AwemeItem) Duration() time.Duration {
	return time.Duration(a.Video.Duration) * time.Millisecond
}
func (a AwemeItem) Description() string {
	return strings.TrimSpace(a.Desc)
}

func (a AwemeItem) Time() string {
	return strings.Replace(time.Unix(a.CreateTime, 0).Format("Mon, 02 Jan 2006 15:04:05 MST"), "  ", " ", -1)
}

func (a AwemeItem) URL() (string, error) {
	if len(a.Video.Play_Addr.URL_List) == 0 {
		return "", errors.New("invalid slice")
	}
	first := a.Video.Play_Addr.URL_List[0]
	loc, err := url.Parse(first)
	if err != nil {
		return "", err
	}
	loc.RawQuery = ""
	loc.Scheme = "https"
	return loc.String(), nil
}

func GetId(uri string) (string, error) {
	url, _ := url.Parse(uri)
	url.RawQuery = ""
	url.Scheme = "https"
	resp, err := http.Get(url.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusBadGateway {
		return "", errors.New("video not found")
	}
	newUrl, _ := url.Parse(resp.Request.URL.String())
	newUrl.RawQuery = ""
	newUrl.Scheme = "http"

	return utils.FileNameWithoutExtension(filepath.Base(newUrl.String())), nil
}

func (a AwemeItem) DownloadVideo(downloadBytesLimit int64) (*os.File, error) {
	addr, err := a.URL()
	if err != nil {
		return nil, err
	}
	res, err := http.Get(addr)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusForbidden {
		return nil, errors.New("got forbidden out of a video")
	}
	defer res.Body.Close()
	size, _ := strconv.Atoi(res.Header.Get("Content-Length"))
	downloadSize := int64(size)
	if downloadSize > downloadBytesLimit {
		return nil, errors.New("too large")
	}
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("%s.%s", u.String(), strings.Split(res.Header.Get("Content-Type"), "/")[1])
	if strings.HasSuffix(filename, ".mpeg") {
		filename = strings.Replace(filename, ".mpeg", ".mp3", 1)
	}
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

func (a AwemeItem) DownloadImagesWithAudio(downloadBytesLimit int64) ([]*os.File, *os.File, error) {
	var addresses []string
	for _, address := range a.ImagePostInfo.Images {
		addresses = append(addresses, address.DisplayImage.UrlList[1])
	}
	var images []*os.File
	for _, item := range addresses {
		res, err := http.Get(item)
		if err != nil {
			res.Body.Close()
			return nil, nil, err
		}
		if res.StatusCode == http.StatusForbidden {
			res.Body.Close()
			return nil, nil, errors.New("got forbidden out of an image")
		}
		u, err := uuid.NewUUID()
		if err != nil {
			res.Body.Close()
			return nil, nil, err
		}
		filename := fmt.Sprintf("%s.%s", u.String(), strings.Split(res.Header.Get("Content-Type"), "/")[1])
		file, err := os.Create(filename)
		if err != nil {
			res.Body.Close()
			file.Close()
			return nil, nil, err
		}
		if _, err := file.ReadFrom(res.Body); err != nil {
			res.Body.Close()
			file.Close()
			return nil, nil, err
		}
		res.Body.Close()
		file.Close()
		openedFile, err := os.Open(file.Name())
		if err != nil {
			openedFile.Close()
			os.Remove(openedFile.Name())
			return nil, nil, err
		}
		images = append(images, openedFile)
	}
	file, err := a.DownloadVideo(downloadBytesLimit)
	if err != nil {
		return nil, nil, err
	}
	return images, file, nil

}
