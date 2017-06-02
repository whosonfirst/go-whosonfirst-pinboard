package internetarchive

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type WaybackMachine struct {
}

func NewWaybackMachine() (*WaybackMachine, error) {

	wb := WaybackMachine{}
	return &wb, nil
}

func (wb *WaybackMachine) SaveURL(u string) (string, error) {

	/*

	   curl -s -v 'https://web.archive.org/save/https://whosonfirst.mapzen.com/theory'

	   < HTTP/1.1 302 Found
	   < Server: Tengine/2.1.0
	   < Date: Tue, 30 May 2017 22:33:17 GMT
	   < Content-Type: text/html;charset=utf-8
	   < Content-Length: 3153
	   < Connection: keep-alive
	   < Content-Location: /web/20170530223315/https://whosonfirst.mapzen.com/theory
	*/

	url := fmt.Sprintf("https://web.archive.org/save/%s", u)

	rsp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer rsp.Body.Close()

	loc := rsp.Header.Get("Content-Location")

	if loc == "" {
		return "", errors.New("Could not determine content-location")
	}

	parts := strings.Split(loc, "/")
	dt := parts[2]

	return dt, nil
}
