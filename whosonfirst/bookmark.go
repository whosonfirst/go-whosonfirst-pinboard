package whosonfirst

import (
       "fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-pinboard/internetarchive"
	"github.com/whosonfirst/go-whosonfirst-pinboard/pinboard"
	"github.com/whosonfirst/go-whosonfirst-pinboard/webpage"		
	"github.com/whosonfirst/go-whosonfirst-uri"	
       "io/ioutil"	
       "net/http"	
       "net/url"
	"strings"
)

func BookmarkURL(to_archive string, wofid int64, config *Config) (*pinboard.Bookmark, error) {

	pb, err := pinboard.NewAPI(config.AuthToken)

	if err != nil {
	   	return nil, err
	}

	wb, err := internetarchive.NewWaybackMachine()

	if err != nil {
	   	return nil, err	
	}
	
	parsed, err := url.Parse(to_archive)

	if err != nil {
		return nil, err
	}

	host := parsed.Host

	if strings.HasPrefix(host, "www.") {
		host = strings.Replace(host, "www.", "", -1)
	}

	// Check to see if we've already bookmarked this

	old_bm, err := pb.GetBookmark(to_archive)

	if err != nil {
	   	return nil, err
	}

	// fetch the WOF record and extract hierarchy and concordances
	// please move me in to a function or something...
	// (20170530/thisisaaronland)

	hierarchy := make(map[string]string)
	concordances := make([]string, 0)
	// placetypes

	abs_url, err := uri.Id2AbsPath(config.DataRoot, wofid)

	if err != nil {
		return nil, err
	}

	rsp, err := http.Get(abs_url)

	if err != nil {
	   	return nil, err
	}

	defer rsp.Body.Close()

	feature, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
	   	return nil, err
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

	title, err := webpage.GetTitle(to_archive)

	if err != nil {
		return nil, err
	}

	// archive URL and get datetime stamp

	dt, err := wb.SaveURL(to_archive)

	if err != nil {
	   	return nil, err
	}

	// build bookmark

	tags := []string{
		host,
		fmt.Sprintf("wof:placetype=%s", placetype),
		fmt.Sprintf("wof:id=%d", wofid),
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

	new_bm := pinboard.Bookmark{
		URL:   to_archive,
		Title: title,
		Tags:  tags,
	}

	err = pb.SaveBookmark(&new_bm)

	if err != nil {
	   	return nil, err
	}

	return &new_bm, nil
}
