package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var intervalFlag = flag.Float64("interval", 1, "Interval between each download (sec)")
var httpClient = &http.Client{}

func main() {
	flag.Parse()
	url := flag.Arg(0)

	fmt.Printf("Scraping %s...", url)
	list, err := getImageList(url)
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("done")
	fmt.Printf("Found %d images.\n", len(list))

	for len(list) != 0 {
		imgUrl := list[0]
		fmt.Printf("Downloading %s...", imgUrl)
		err := download(imgUrl)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Retry...")
		} else {
			fmt.Println("done")
			list = list[1:]
		}
		if *intervalFlag > 0 {
			fmt.Printf("Waiting for %f seconds...", *intervalFlag)
			time.Sleep(time.Duration(*intervalFlag) * time.Second)
			fmt.Println("OK.")
		}
	}
}

func getImageList(url string) ([]string, error) {
	res, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}

	list := []string{}
	doc.Find(`div.ently_text a[target="_blank"]`).Each(func(_ int, el *goquery.Selection) {
		href, ok := el.Attr("href")
		if !ok { // href does not exists, so ignore it
			return
		}
		list = append(list, href)
	})

	return list, nil
}

func download(rawurl string) error {
	filename, err := fileNameOf(rawurl)
	if err != nil {
		return err
	}
	resp, err := http.Get(rawurl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

var reInPath = regexp.MustCompile("[^/]+$")

func fileNameOf(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	file := reInPath.FindString(url.Path)
	if file == "" {
		return "", fmt.Errorf("Filename not found: %s", rawurl)
	}
	return file, nil
}
