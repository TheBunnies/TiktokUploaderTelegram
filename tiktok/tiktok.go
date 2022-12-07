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

const Origin = "https://api.tiktokv.com"

const Manifest = "221"
const AppVersion = "20.2.1"

func Parse(id string) (uint64, error) {
	return strconv.ParseUint(id, 10, 64)
}

func NewAwemeItem(id uint64) (*AwemeItem, error) {
	req, err := http.NewRequest("GET", Origin+"/aweme/v1/feed/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "com.ss.android.ugc.trill/494+Mozilla/5.0+(Linux;+Android+12;+2112123G+Build/SKQ1.211006.001;+wv)+AppleWebKit/537.36+(KHTML,+like+Gecko)+Version/4.0+Chrome/107.0.5304.105+Mobile+Safari/537.36")
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("aweme_id", strconv.FormatUint(id, 10))
	q.Add("version_name", AppVersion)
	q.Add("iid", "6165993682518218889")
	q.Add("build_number", AppVersion)
	q.Add("manifest_version_code", Manifest)
	q.Add("update_version_code", Manifest)
	q.Add("openudid", utils.RandomString(16))
	q.Add("uuid", utils.RandomDigits(16))
	q.Add("_rticket", strconv.FormatInt(time.Now().Unix()*1000, 10))
	q.Add("ts", strconv.FormatInt(time.Now().Unix(), 10))
	q.Add("device_brand", "Google")
	q.Add("device_type", "Pixel 4")
	q.Add("resolution", "1080*1920")
	q.Add("dpi", "420")
	q.Add("os_version", "10")
	q.Add("os_api", "29")
	q.Add("carrier_region", "US")
	q.Add("sys_region", "US")
	q.Add("region", "US")
	q.Add("app_name", "trill")
	q.Add("app_language", "en")
	q.Add("language", "en")
	q.Add("timezone_name", "America/New_York")
	q.Add("timezone_offset", "-14400")
	q.Add("channel", "googleplay")
	q.Add("ac", "wifi")
	q.Add("mcc_mnc", "310260")
	q.Add("is_my_cn", "0")
	q.Add("aid", "1180")
	q.Add("ssmix", "a")
	q.Add("as", "a1qwert123")
	q.Add("cp", "cbfhckdckkde1")
	q.Add("device_id", utils.RandomDigits(19))

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
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadGateway {
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
