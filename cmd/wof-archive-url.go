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
	"strings"
)

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
	var token = flag.String("token", "", "A valid Pinboard API auth token.")
	var data_root = flag.String("data-root", "https://whosonfirst.mapzen.com/data/", "...")

	flag.Parse()

	if *wofid == 0 {
		log.Fatal("Missing WOF ID")
	}

	if *to_archive == "" {
		log.Fatal("Missing URL")
	}

	if *token == "" {
		log.Fatal("Missing Pinboard API token")
	}

	//

	parsed, err := url.Parse(*to_archive)

	if err != nil {
		log.Fatal(err)
	}

	host := parsed.Host

	if strings.HasPrefix(host, "www.") {
		host = strings.Replace(host, "www.", "", -1)
	}

	// fetch the WOF record and extract hierarchy and concordances
	// please move me in to a function or something...
	// (20170530/thisisaaronland)

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

	hierarchy := make(map[string]string)

	h := gjson.GetBytes(feature, "properties.wof:hierarchy")

	h.ForEach(func(key, value gjson.Result) bool {

		for placetype, id := range value.Map() {
			hierarchy[id.String()] = placetype
		}

		return true // keep iterating
	})

	p := gjson.GetBytes(feature, "properties.wof:placetype")
	placetype := p.String()

	concordances := make([]string, 0)

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

	title, err := ParseURL(*to_archive)

	if err != nil {
		log.Fatal(err)
	}

	// archive URL and get datetime stamp

	dt, err := ArchiveURL(*to_archive)

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

	bm := Bookmark{
		URL:   *to_archive,
		Title: title,
		Tags:  tags,
	}

	enc_bm, err := json.Marshal(bm)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", pretty.Pretty(enc_bm))

	// go!

	err = SaveBookmark(bm, *token)

	if err != nil {
		log.Fatal(err)
	}
}

func GetBookmark(uri string, token string) error {

	req, err := http.NewRequest("GET", "https://api.pinboard.in/v1/posts/get", nil)

	if err != nil {
		return err
	}

	params := url.Values{}

	params.Set("url", uri)
	params.Set("auth_token", token)
	params.Set("format", "json")

	req.URL.RawQuery = params.Encode()

	client := &http.Client{}

	rsp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return errors.New(rsp.Status)
	}

	body, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return err
	}

	/*

	2017/05/30 20:07:11 {"date":"2017-05-31T01:10:23Z","user":"mapzen","posts":[{"href":"https:\/\/missionlocal.org\/2016\/10\/trump-supporters-kicked-out-of-zeitgeist-bar-for-lewd-comments\/","description":"Trump Supporters Kicked Out of Zeitgeist Bar for Lewd Comments \u00bb MissionLocal","extended":"","meta":"725e08849afb1c22b5db857a711eda19","hash":"ab1c141c11e9d3522ff03c53ae3cf996","time":"2017-05-31T01:10:23Z","shared":"yes","toread":"no","tags":"missionlocal.org wof:placetype=venue wof:id=588371677 archive:dt=20170531011022 wof:region_id=85688637 wof:venue_id=588371677 wof:continent_id=102191575 wof:country_id=85633793 wof:county_id=102087579 wof:locality_id=85922583 wof:neighbourhood_id=85887415"}]}

	*/
	
	log.Println(string(body))

	return nil

}

func SaveBookmark(bm Bookmark, token string) error {

	req, err := http.NewRequest("GET", "https://api.pinboard.in/v1/posts/add", nil)

	if err != nil {
		return err
	}

	params := bm.ToParams()
	params.Set("auth_token", token)
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

func ArchiveURL(u string) (string, error) {

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

func ParseURL(url string) (string, error) {

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
