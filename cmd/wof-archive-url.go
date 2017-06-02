package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tidwall/pretty"
	"github.com/whosonfirst/go-whosonfirst-pinboard/whosonfirst"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

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

	config, err := whosonfirst.NewDefaultConfig()

	if err != nil {
		log.Fatal(err)
	}

	config.DataRoot = *data_root
	config.AuthToken = *auth_token

	bm, err := whosonfirst.BookmarkURL(*to_archive, *wofid, config)

	if err != nil {
		log.Fatal(err)
	}

	enc_bm, err := json.Marshal(bm)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Sprintf("%s", pretty.Pretty(enc_bm))

	os.Exit(0)
}
