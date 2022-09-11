package helper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/gcli/v2/progress"
	"github.com/gookit/goutil/fmtutil"
)

// Downloader struct definition.
// refer: https://gist.github.com/albulescu/e61979cc852e4ee8f49c
type Downloader struct {
	saveAs string // build by SaveDir + Filename

	FileURL  string
	SaveDir  string
	Filename string // save file name.
	Progress bool   // display progress info
}

// Download begin
func (d *Downloader) Download() error {
	return nil
}

// show download progress info
func (d *Downloader) showProgress() {
}

// Download file from remote URL.
// from https://gist.github.com/albulescu/e61979cc852e4ee8f49c
func Download(url, saveDir string, rename ...string) error {
	filename := path.Base(url)
	fmt.Printf("Downloading file from %s\n", url)

	saveDir = strings.TrimRight(saveDir, "/")
	saveAs := saveDir + "/" + filename
	if len(rename) > 0 {
		saveAs = saveDir + "/" + rename[0]
	}

	exist := true
	if _, err := os.Stat(saveAs); err != nil {
		if os.IsNotExist(err) {
			exist = false
		}
	}

	if exist {
		fmt.Printf("Remove old file %s\n", saveAs)
		if err := os.Remove(saveAs); err != nil {
			return err
		}
	}

	start := time.Now()
	outFile, err := os.Create(saveAs)
	if err != nil {
		fmt.Println(saveAs)
		panic(err)
	}
	//noinspection GoUnhandledErrorResult
	defer outFile.Close()

	headResp, err := http.Head(url)
	if err != nil {
		panic(err)
	}
	//noinspection GoUnhandledErrorResult
	defer headResp.Body.Close()

	// get remote file size from header.
	length := headResp.Header.Get("Content-Length")
	size, err := strconv.Atoi(length)
	if err != nil { // no "Content-Length"
		size = 0
	}

	done := make(chan int64)
	go printDownloadPercent(done, saveAs, int64(size))

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	n, err := io.Copy(outFile, resp.Body)
	if err != nil {
		panic(err)
	}

	done <- n

	elapsed := time.Since(start)
	fmt.Printf("Download completed in %s\n", elapsed)
	return nil
}

func printDownloadPercent(done chan int64, path string, total int64) {
	fmtSize := "unknown"
	if total > 0 {
		fmtSize = fmtutil.DataSize(uint64(total))
	}

	fmt.Printf("Download total size: %s\n", fmtSize)

	for {
		select {
		case <-done: // end, output newline
			fmt.Println()
			return
		default:
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}

			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			// move to begin of the line and clear line text.
			fmt.Print("\x0D\x1B[2K")
			fmt.Printf("Downloaded %s", fmtutil.DataSize(uint64(size)))

			if total > 0 {
				percent := float64(size) / float64(total) * 100
				fmt.Printf(", Progress %.0f%%", percent)
			}
		}

		time.Sleep(time.Second)
	}
}

// SimpleDownload simple download
func SimpleDownload(url, saveAs string) (err error) {
	newFile, err := os.Create(saveAs)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer newFile.Close()

	s := progress.LoadingSpinner(
		progress.GetCharsTheme(18),
		time.Duration(100)*time.Millisecond,
	)

	s.Start("%s Downloading ... ...")

	client := http.Client{Timeout: 300 * time.Second}
	// Request the remote url.
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	_, err = io.Copy(newFile, resp.Body)
	if err != nil {
		return
	}

	s.Stop("Download completed")
	return
}
