package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	readability "github.com/go-shiori/go-readability"
	"github.com/tmahlburg/obelisk"
	archiveorg "github.com/wabarc/archive.org/pkg"
	"golang.org/x/net/html"
)

func main() {
	// archival type
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "article":
			if len(os.Args) > 2 {
				archiveArticle(os.Args[2], os.Args[3:])
			} else {
				printHelp()
				os.Exit(1)
			}
		case "help":
			printHelp()
		default:
			printHelp()
			os.Exit(1)
		}
	} else {
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("HELP!")
}

type article struct {
	url         string
	archivedUrl string
	date        time.Time
	tags        []string
	author      string
	title       string
}

func archiveArticle(url string, tags []string) {
	snapShotCh := make(chan string)
	go createSnapshot(url, snapShotCh)

	fullPageCh := make(chan []byte)
	go archiveFullPage(url, fullPageCh)

	headerPath := "header.html"
	headerCh := make(chan string)
	go readFile(headerPath, headerCh)

	footerPath := "footer.html"
	footerCh := make(chan string)
	go readFile(footerPath, footerCh)

	page := downloadPage(url)

	page_reader := bytes.NewReader(page)
	metaData := getMetaData(page_reader)
	metaData.url = url
	metaData.tags = tags

	page_reader.Seek(0, 0)
	plain, stripped := makeReadable(url, page_reader)

	title := "<h1>" + metaData.title + "</h1><hr>"
	stripped = fmt.Sprintf(<-headerCh, metaData.title) + title +  stripped + <-footerCh
	metaData.archivedUrl = <-snapShotCh

	fileContent := make(map[string][]byte)
	fileContent["plain.txt"] = []byte(plain)
	fileContent["stripped.html"] = []byte(stripped)
	fileContent["full_page.html"] = <-fullPageCh
	fileContent[".meta"] = []byte(createMetaFile(&metaData))
	dirName := metaData.date.Format("2006-02-01") + "-" + metaData.title
	saveToDisk(&fileContent, dirName)
}

func downloadPage(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return page
}

func getMetaData(page io.Reader) article {
	metaData := article{date: time.Now()}
	/* Parse HTML */
	doc, err := html.Parse(page)
	if err != nil {
		panic(err)
	}

	var findMeta func(n *html.Node, metaData *article)
	findMeta = func(n *html.Node, metaData *article) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "meta":
				switch n.Attr[0].Val {
				case "author":
					metaData.author = n.Attr[1].Val
				case "og:title":
					metaData.author = n.Attr[1].Val
					/* Implementation unsure
					case "og:url":
					metaData.url = n.Attr[1].Val
					*/
				}
			case "title":
				if metaData.title == "" {
					metaData.title = n.FirstChild.Data
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findMeta(c, metaData)
		}
	}
	findMeta(doc, &metaData)

	return metaData
}

func makeReadable(url string, page io.Reader) (string, string) {
	art, err := readability.FromReader(page, url)
	if err != nil {
		panic(err)
	}

	return art.TextContent, art.Content
}

func createSnapshot(url string, c chan string) {
	var wbrc archiveorg.Archiver
	result, err := wbrc.Wayback([]string{url})
	if err != nil {
		panic(err)
	}

	c <- result[url]
}

func archiveFullPage(url string, c chan []byte) {
	// Create archive
	req := obelisk.Request{URL: url}

	arc := obelisk.Archiver{EnableLog: true}
	arc.Validate()

	result, _, err := arc.Archive(context.Background(), req)
	if err != nil {
		panic(err)
	}

	c <- result
}

func saveToDisk(fileContent *map[string][]byte, dirName string) {
	os.MkdirAll(dirName, os.ModePerm)
	// write given files
	var wg sync.WaitGroup
	wg.Add(len(*fileContent))
	for fileName, content := range *fileContent {
		go func(fileName string, content []byte, dirName string) {
			defer wg.Done()
			err := ioutil.WriteFile(dirName+"/"+fileName, content, 0644)
			if err != nil {
				panic(err)
			}
		}(fileName, content, dirName)
	}
	wg.Wait()
}

func createMetaFile(metaData *article) string {
	content := "title=" + metaData.title + "\n" +
		"tags=[" + strings.Join(metaData.tags, " | ") + "]\n" +
		"date=" + metaData.date.Format("2006-02-01") + "\n" +
		"author=" + metaData.author + "\n" +
		"url=" + metaData.url + "\n" +
		"archived=" + metaData.archivedUrl + "\n"
	return content
}

func readFile(fileName string, returnCh chan string) {
	cont, err := ioutil.ReadFile(fileName)
	if err != nil {
    	panic(err)
	}
	returnCh <- string(cont)
}
