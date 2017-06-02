package pinboard

import (
	"errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type API struct {
	AuthToken string
}

func NewAPI(auth_token string) (*API, error) {

	p := API{
		AuthToken: auth_token,
	}

	return &p, nil
}

func (p *API) SaveBookmark(bm *Bookmark) error {

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

func (p *API) GetBookmark(uri string) (*Bookmark, error) {

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
