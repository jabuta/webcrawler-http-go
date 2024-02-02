package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

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
						attr.Val[0:1] != "#" &&
						!strings.Contains(attr.Val, "cdn-cgi") {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
}

func parseLinks(links []string, root string) []string {
	var newLinks []string
	for _, link := range links {
		fullLink := link
		if link[0:1] == "/" {
			fullLink = root + link
		}
		u, err := url.Parse(fullLink)
		if err != nil || u.Scheme == "" || u.Host == "" {
			fmt.Printf("%v is not a proper URL\n", fullLink)
			continue
		} else if u.Host != root {
			fmt.Printf("%v is not an internal link\n", fullLink)
			continue
		}
		newLinks = append(newLinks, u.Scheme+"://"+u.Host+u.Path)

	}
	return newLinks
}

func main() {
	fmt.Print("Enter Root Domain:")
	rootURL := "https://www.obrasximpuestos.com"
	linkToParse, err := getLinksOnPage(rootURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, link := range linkToParse {
		fmt.Println(link)
	}
	parsedLinks := parseLinks(linkToParse, rootURL)
	for _, link := range parsedLinks {
		fmt.Println(link)
	}
}
