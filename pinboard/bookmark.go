package pinboard

import (
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
