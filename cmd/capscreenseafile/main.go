package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jgrossophoff/seafile"
)

var (
	baseURL,
	repoID,
	username,
	password string
)

func main() {
	if !checkDeps() {
		exitErr("import (imagemagick) and xclip need to be installed and available in $PATH")
	}

	flag.StringVar(&baseURL, "baseurl", "https://your.seafile.org", "Seafile API domain without path")
	flag.StringVar(&repoID, "repo", "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", "Seafile upload repository id")
	flag.StringVar(&username, "username", "your@username.org", "Seafile username")
	flag.StringVar(&password, "password", "password", "Seafile password")
	flag.Parse()

	cap := NewScreenCapture()
	if err := cap.Capture(); err != nil {
		exitErr("image capture aborted: %s", err)
	}
	defer cap.Cleanup()

	client := seafile.NewClient(baseURL, repoID)
	if err := client.FetchToken(username, password); err != nil {
		exitErr("unable to fetch seafile token: %s", err)
	}

	resp, err := client.UploadFile(cap.Filepath())
	if err != nil {
		exitErr("unable to upload file: %s", err)
	}

	numUploads := len(resp)
	if numUploads != 1 {
		exitErr("expected number of uploaded items to be 1, was %d", numUploads)
	}

	upload := resp[0]
	share, err := client.ShareFile(upload.Name)
	if err != nil {
		exitErr("error creating share link: %s", err)
	}

	copyLink(share.Link)
}

type ScreenCapture struct {
	filepath string
}

func (c *ScreenCapture) Capture() error {
	_, err := exec.Command("import", c.filepath).CombinedOutput()
	return err
}

func (c *ScreenCapture) Filepath() string {
	return c.filepath
}

func (c *ScreenCapture) Cleanup() error {
	return os.Remove(c.filepath)
}

func NewScreenCapture() *ScreenCapture {
	return &ScreenCapture{fmt.Sprintf("%s/%d.png", os.TempDir(), time.Now().Unix())}
}

func checkDeps() bool {
	if _, err := exec.LookPath("import"); err != nil {
		return false
	}
	if _, err := exec.LookPath("xclip"); err != nil {
		return false
	}
	return true
}

func copyLink(link string) {
	cmd := exec.Command("xclip")

	p, err := cmd.StdinPipe()
	if err != nil {
		exitErr("error creating pipe to xclip: %s", err)
	}
	defer p.Close()

	if err := cmd.Start(); err != nil {
		exitErr("error running xclip command: %s")
	}

	b := strings.NewReader(link)
	_, err = io.Copy(p, b)
	if err != nil {
		exitErr("error writing to xclip pipe: %s", err)
	}
	notify("uploaded screen capture and copied link to clipboard!")
}

func notify(msg string) {
	exec.Command("notify-send", msg).CombinedOutput()
}

func exitErr(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)
	notify(msg)
	fmt.Println(msg)
	os.Exit(1)
}
