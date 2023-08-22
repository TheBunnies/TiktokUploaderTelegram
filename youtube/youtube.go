package youtube

import (
	"github.com/google/uuid"
	"os"
	"os/exec"
)

func DownloadVideo(url string) (*os.File, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	filename := u.String() + ".mp4"
	cmd := exec.Command("yt-dlp", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/mp4", url, "-o", u.String())
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	openedFile, err := os.Open(filename)
	if err != nil {
		openedFile.Close()
		os.Remove(openedFile.Name())
		return nil, err
	}
	return openedFile, nil
}
