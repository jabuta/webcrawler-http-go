package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type linkCount struct {
	link  string
	count int
}

func sortLinks(linkMap map[string]int) []linkCount {
	linkCountList := make([]linkCount, len(linkMap))
	k := 0
	for link, count := range linkMap {
		linkCountList[k] = linkCount{link, count}
		k++
	}
	for i := 0; i < len(linkCountList)-1; i++ {
		for j := i + 1; j < len(linkCountList); j++ {
			if linkCountList[i].count < linkCountList[j].count {
				interim := linkCountList[i]
				linkCountList[i] = linkCountList[j]
				linkCountList[j] = interim
			}
		}
	}
	return linkCountList
}

// Collect all links from response body and return it as an array of strings
func getLinksOnPage(url string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		errBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, string(errBody)))
	}

	var links []string
	z := html.NewTokenizer(res.Body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return links, nil
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" &&
						len(attr.Val) > 0 &&
						attr.Val[0:1] != "#" &&
						!strings.Contains(attr.Val, "cdn-cgi") {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
}

func normaliseLinks(links []string, root url.URL) []string {
	var normLinks []string
	for _, link := range links {
		fmt.Printf("normalising %s\n", link)
		fullLink := link
		if link[0:1] == "/" {
			fullLink = root.Scheme + "://" + root.Host + link
		}
		u, err := url.Parse(fullLink)
		if err != nil || !strings.Contains(u.Scheme, "http") || u.Host == "" {
			fmt.Printf("%v is not a proper http/s URL\n", fullLink)
			continue
		} else if u.Host != root.Host {
			fmt.Printf("%v is not an internal link\n", fullLink)
			continue
		}
		normLinks = append(normLinks, "https://"+u.Host+u.Path)

	}
	return normLinks
}

func crawlSite(linksToCrawl []string, crawledLinks map[string]int, root url.URL) {
	for _, link := range linksToCrawl {
		if _, ok := crawledLinks[link]; ok {
			crawledLinks[link]++
			continue
		}
		if len(linksToCrawl) == 1 {
			crawledLinks[link] = 0
		} else {
			crawledLinks[link] = 1
		}
		linksToParse, err := getLinksOnPage(link)
		if err != nil {
			fmt.Println(err)
			continue
		}
		parsedLinks := normaliseLinks(linksToParse, root)
		crawlSite(parsedLinks, crawledLinks, root)
	}
}

func main() {
	crawledLinks := make(map[string]int)
	fmt.Print("Enter Root Domain:")
	rootURLString := os.Args[1]
	rootURL, err := url.Parse(rootURLString)
	if err != nil || rootURL.Scheme == "" || rootURL.Host == "" {
		fmt.Println(err)
		return
	}
	linksToCrawl := normaliseLinks([]string{rootURLString}, *rootURL)
	crawlSite(linksToCrawl, crawledLinks, *rootURL)

	sortedLinks := sortLinks(crawledLinks)
	for _, linkInfo := range sortedLinks {
		fmt.Printf("%s has %v instances\n", linkInfo.link, linkInfo.count)
	}

}
