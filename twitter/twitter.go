package twitter

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/TheBunnies/TiktokUploaderTelegram/config"
	"github.com/gocolly/colly"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	rgxBearer  = regexp.MustCompile(`"Bearer.*?"`)
	rgxNum     = regexp.MustCompile(`[0-9]+`)
	rgxAddress = regexp.MustCompile(`https.*m3u8`)
	rgxFormat  = regexp.MustCompile(`.*m3u8`)
)

func NewTwitterVideoDownloader(url string) *VideoDownloader {
	self := new(VideoDownloader)
	self.VideoUrl = url
	return self
}

func (s *VideoDownloader) GetBearerToken() string {
	c := colly.NewCollector()
	if config.ProxyUrl != "" {
		c.SetProxy(config.ProxyUrl)
	}
	c.SetProxy(config.ProxyUrl)
	c.OnResponse(func(r *colly.Response) {
		s.BearerToken = strings.Trim(rgxBearer.FindString(string(r.Body)), `"`)
	})

	c.Visit("https://abs.twimg.com/web-video-player/TwitterVideoPlayerIframe.cefd459559024bfb.js")

	return s.BearerToken
}

func (s *VideoDownloader) GetXGuestToken() string {
	c := colly.NewCollector()

	if config.ProxyUrl != "" {
		c.SetProxy(config.ProxyUrl)
	}
	c.SetProxy(config.ProxyUrl)
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Authorization", s.BearerToken)
	})

	c.OnResponse(func(r *colly.Response) {
		s.GuestToken = rgxNum.FindString(string(r.Body))
	})

	c.Post("https://api.twitter.com/1.1/guest/activate.json", nil)

	return s.GuestToken
}

func (s *VideoDownloader) GetM3U8Urls() string {
	var m3u8_urls string

	c := colly.NewCollector()

	if config.ProxyUrl != "" {
		c.SetProxy(config.ProxyUrl)
	}
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Authorization", s.BearerToken)
		r.Headers.Set("x-guest-token", s.GuestToken)
	})

	c.OnResponse(func(r *colly.Response) {
		m3u8_urls = strings.ReplaceAll(rgxAddress.FindString(string(r.Body)), "\\", "")
	})

	url := "https://api.twitter.com/1.1/videos/tweet/config/" +
		strings.TrimPrefix(s.VideoUrl, "https://twitter.com/i/status/") +
		".json"

	c.Visit(url)

	return m3u8_urls
}

func (s *VideoDownloader) GetM3U8Url(m3u8_urls string) string {
	var m3u8_url string

	c := colly.NewCollector()

	if config.ProxyUrl != "" {
		c.SetProxy(config.ProxyUrl)
	}

	c.OnResponse(func(r *colly.Response) {
		m3u8_urls := rgxFormat.FindAllString(string(r.Body), -1)
		m3u8_url = "https://video.twimg.com" + m3u8_urls[len(m3u8_urls)-1]
	})

	c.Visit(m3u8_urls)

	return m3u8_url
}

func (s *VideoDownloader) Download() (*os.File, error) {
	s.GetBearerToken()
	s.GetXGuestToken()
	m3u8_urls := s.GetM3U8Urls()
	m3u8_url := s.GetM3U8Url(m3u8_urls)

	sum := md5.Sum([]byte(m3u8_url))
	filename := hex.EncodeToString(sum[:]) + ".mp4"

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
	}

	response, err := client.Get(m3u8_url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	cmd := exec.Command("ffmpeg", "-y", "-http_proxy", config.ProxyUrl, "-i", m3u8_url, "-c", "copy", filename)
	cmd.Run()
	openedFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return openedFile, nil
}
