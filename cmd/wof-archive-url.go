package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type WaybackAPI struct {
}

func NewWaybackAPI() (*WaybackAPI, error) {

	wb := WaybackAPI{}
	return &wb, nil
}

func (wb *WaybackAPI) SaveURL(u string) (string, error) {

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

type PinboardAPI struct {
	AuthToken string
}

func NewPinboardAPI(auth_token string) (*PinboardAPI, error) {

	p := PinboardAPI{
		AuthToken: auth_token,
	}

	return &p, nil
}

func (p *PinboardAPI) SaveBookmark(bm *Bookmark) error {

	req, err := http.NewRequest("GET", "https://api.pinboard.in/v1/posts/add", nil)

	if err != nil {
		return err
	}

	params := bm.ToParams()
	params.Set("auth_token", p.AuthToken)
	req.URL.RawQuery = (*params).Encode()

	client := &http.Client{}

	rsp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return errors.New(rsp.Status)
	}

	return nil
}

func (p *PinboardAPI) GetBookmark(uri string) (*Bookmark, error) {

	req, err := http.NewRequest("GET", "https://api.pinboard.in/v1/posts/get", nil)

	if err != nil {
		return nil, err
	}

	params := url.Values{}

	params.Set("url", uri)
	params.Set("auth_token", p.AuthToken)
	params.Set("format", "json")

	req.URL.RawQuery = params.Encode()

	client := &http.Client{}

	rsp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return nil, errors.New(rsp.Status)
	}

	body, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	post := gjson.GetBytes(body, "posts.0")

	if !post.Exists() {
		return nil, nil
	}

	var url string
	var title string
	var tags []string

	for k, v := range post.Map() {

		switch k {
		case "href":
			url = v.String()
		case "description":
			title = v.String()
		case "tags":
			tags = strings.Split(v.String(), " ")
		default:
			// pass
		}
	}

	bm := Bookmark{
		URL:   url,
		Title: title,
		Tags:  tags,
	}

	return &bm, nil

}

type Bookmark struct {
	URL   string
	Title string
	Tags  []string
}

func (bm Bookmark) ToParams() *url.Values {

	params := url.Values{}

	params.Set("url", bm.URL)
	params.Set("description", bm.Title)
	params.Set("tags", strings.Join(bm.Tags, " "))

	return &params
}

func main() {

	var wofid = flag.Int64("wofid", 0, "A valid Who's On First ID")
	var to_archive = flag.String("url", "", "The URL you want to bookmark.")
	var auth_token = flag.String("auth-token", "", "A valid Pinboard API auth token.")
	var data_root = flag.String("data-root", "https://whosonfirst.mapzen.com/data/", "...")

	flag.Parse()

	if *wofid == 0 {
		log.Fatal("Missing WOF ID")
	}

	if *to_archive == "" {
		log.Fatal("Missing URL")
	}

	if *auth_token == "" {
		*auth_token = os.Getenv("PINBOARD_AUTH_TOKEN")
	}

	if *auth_token == "" {

		usr, err := user.Current()

		if err != nil {
			log.Fatal(err)
		}

		home := usr.HomeDir
		dotpb := filepath.Join(home, ".pinboard")
		creds := filepath.Join(dotpb, "credentials")

		_, err = os.Stat(creds)

		if err == nil {

			fh, err := os.Open(creds)

			if err != nil {
				log.Fatal(err)
			}

			body, err := ioutil.ReadAll(fh)

			if err != nil {
				log.Fatal(err)
			}

			*auth_token = string(body)
		}
	}

	if *auth_token == "" {
		log.Fatal("Missing Pinboard API token")
	}

	pb, err := NewPinboardAPI(*auth_token)

	if err != nil {
		log.Fatal(err)
	}

	wb, err := NewWaybackAPI()

	if err != nil {
		log.Fatal(err)
	}

	// Who are we trying to bookmark

	parsed, err := url.Parse(*to_archive)

	if err != nil {
		log.Fatal(err)
	}

	host := parsed.Host

	if strings.HasPrefix(host, "www.") {
		host = strings.Replace(host, "www.", "", -1)
	}

	// Check to see if we've already bookmarked this

	old_bm, err := pb.GetBookmark(*to_archive)

	if err != nil {
		log.Fatal(err)
	}

	// fetch the WOF record and extract hierarchy and concordances
	// please move me in to a function or something...
	// (20170530/thisisaaronland)

	hierarchy := make(map[string]string)
	concordances := make([]string, 0)
	// placetypes

	abs_url, err := uri.Id2AbsPath(*data_root, *wofid)

	if err != nil {
		log.Fatal(err)
	}

	rsp, err := http.Get(abs_url)

	if err != nil {
		log.Fatal(err)
	}

	defer rsp.Body.Close()

	feature, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		log.Fatal(err)
	}

	h := gjson.GetBytes(feature, "properties.wof:hierarchy")

	h.ForEach(func(key, value gjson.Result) bool {

		for placetype, id := range value.Map() {
			hierarchy[id.String()] = placetype
		}

		return true // keep iterating
	})

	p := gjson.GetBytes(feature, "properties.wof:placetype")
	placetype := p.String()

	c := gjson.GetBytes(feature, "properties.wof:concordances")

	if c.Exists() {

		c.ForEach(func(key, value gjson.Result) bool {

			if key.String() != "sg:id" {
				tag := fmt.Sprintf("%s=%s", key, value)
				concordances = append(concordances, tag)
			}

			return true
		})
	}

	wof_tags := make([]string, 0)

	t := gjson.GetBytes(feature, "properties.wof:tags")

	if t.Exists() {

		t.ForEach(func(key, value gjson.Result) bool {
			wof_tags = append(wof_tags, value.String())
			return true
		})
	}

	// get URL title

	title, err := GetTitle(*to_archive)

	if err != nil {
		log.Fatal(err)
	}

	// archive URL and get datetime stamp

	dt, err := wb.SaveURL(*to_archive)

	if err != nil {
		log.Fatal(err)
	}

	// build bookmark

	tags := []string{
		host,
		fmt.Sprintf("wof:placetype=%s", placetype),
		fmt.Sprintf("wof:id=%d", *wofid),
		fmt.Sprintf("archive:dt=%s", dt),
	}

	for _, t := range concordances {
		tags = append(tags, t)
	}

	for _, t := range wof_tags {
		tags = append(tags, t)
	}

	for id, pt := range hierarchy {
		t := fmt.Sprintf("wof:%s=%s", pt, id)
		tags = append(tags, t)
	}

	if old_bm != nil {

		for _, t := range old_bm.Tags {
			tags = append(tags, t)
		}
	}

	new_bm := Bookmark{
		URL:   *to_archive,
		Title: title,
		Tags:  tags,
	}

	enc_bm, err := json.Marshal(new_bm)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", pretty.Pretty(enc_bm))

	// go!

	err = pb.SaveBookmark(&new_bm)

	if err != nil {
		log.Fatal(err)
	}
}

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
