package webpage

import (
	"errors"
	"golang.org/x/net/html"
	"net/http"
)

func GetTitle(url string) (string, error) {

	rsp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	doc, err := html.Parse(rsp.Body)

	if err != nil {
		return "", err
	}

	defer rsp.Body.Close()

	var title string
	var f func(*html.Node)

	f = func(n *html.Node) {

		if n.Type == html.ElementNode && n.Data == "title" {
			title = n.FirstChild.Data
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	if title == "" {
		return "", errors.New("Failed to glean title from URL")
	}

	return title, nil
}
