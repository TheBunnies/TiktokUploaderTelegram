package tiktok

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TheBunnies/TiktokUploaderTelegram/config"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const Origin = "http://api2.musical.ly"

func Parse(id string) (uint64, error) {
	return strconv.ParseUint(id, 10, 64)
}

func NewAwemeDetail(id uint64) (*AwemeDetail, error) {
	req, err := http.NewRequest("GET", Origin+"/aweme/v1/aweme/detail/", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = "aweme_id=" + strconv.FormatUint(id, 10)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36")

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
		Timeout:   10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	var detail struct {
		Aweme_Detail AwemeDetail
	}
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return nil, err
	}
	return &detail.Aweme_Detail, nil
}

func (a AwemeDetail) Duration() time.Duration {
	return time.Duration(a.Video.Duration) * time.Millisecond
}
func (a AwemeDetail) Description() string {
	return strings.TrimSpace(a.Desc)
}

func (a AwemeDetail) Time() string {
	return strings.Replace(time.Unix(a.Create_Time, 0).Format("Mon, 02 Jan 2006 15:04:05 MST"), "  ", " ", -1)
}

func (a AwemeDetail) URL() (string, error) {
	if len(a.Video.Play_Addr.URL_List) == 0 {
		return "", errors.New("invalid slice")
	}
	first := a.Video.Play_Addr.URL_List[0]
	loc, err := url.Parse(first)
	if err != nil {
		return "", err
	}
	loc.RawQuery = ""
	loc.Scheme = "http"
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
	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("video not found")
	}
	newUrl, _ := url.Parse(resp.Request.URL.String())
	newUrl.RawQuery = ""
	newUrl.Scheme = "http"

	return utils.FileNameWithoutExtension(filepath.Base(newUrl.String())), nil
}

func (a AwemeDetail) DownloadVideo() (*os.File, error) {
	addr, err := a.URL()
	if err != nil {
		return nil, err
	}
	res, err := http.Get(addr)
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
