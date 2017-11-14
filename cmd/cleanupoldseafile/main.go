package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jgrossophoff/seafile"
)

var (
	baseURL,
	repoID,
	username,
	password string
	maxAge time.Duration
)

func main() {
	flag.StringVar(&baseURL, "baseurl", "https://your.seafile.org", "Seafile API domain without path")
	flag.StringVar(&repoID, "repo", "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", "Seafile upload repository id")
	flag.StringVar(&username, "username", "your@username.org", "Seafile username")
	flag.StringVar(&password, "password", "password", "Seafile password")
	flag.DurationVar(&maxAge, "maxage", time.Hour*24*14, "A file's max age until it's cleaned up")
	flag.Parse()

	client := seafile.NewClient(baseURL, repoID)
	if err := client.FetchToken(username, password); err != nil {
		exitErr("unable to fetch seafile token: %s", err)
	}

	files, err := client.ListDirectoryEntries()
	if err != nil {
		exitErr("unable to list directory entries: %s", err)
	}

	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go func(f *seafile.ListDirectoryResponseItem) {
			d, err := client.FileDetail(f.Name)
			if err != nil {
				fmt.Println("unable to get detail for file %s: %s", f.Name, err)
				return
			}
			if fileReadyForCleanup(maxAge, d.Mtime) {
				fmt.Printf("deleting file %s %q\n", f.ID, f.Name)
				client.DeleteFile(d.Name)
			}
			wg.Done()
		}(f)
	}
	wg.Wait()
}

func fileReadyForCleanup(maxAge time.Duration, mtime uint64) bool {
	return time.Now().Sub(time.Unix(int64(mtime), 0)) > maxAge
}

func exitErr(f string, args ...interface{}) {
	fmt.Printf(f+"\n", args...)
	os.Exit(1)
}
