package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/whosonfirst/go-sanitize"
	"github.com/whosonfirst/go-whosonfirst-pinboard/whosonfirst"
	"log"
	"net/http"
	"os"
	_ "os/user"
	_ "path/filepath"
	"strconv"
)

func main() {

	var auth_token = flag.String("auth-token", "", "A valid Pinboard API auth token.")
	var data_root = flag.String("data-root", "https://whosonfirst.mapzen.com/data/", "...")

	var host = flag.String("host", "localhost", "The hostname to listen for requests on")
	var port = flag.Int("port", 8080, "The port number to listen for requests on")

	flag.Parse()

	if *auth_token == "" {
		*auth_token = os.Getenv("PINBOARD_AUTH_TOKEN")
	}

	/*
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
	*/

	if *auth_token == "" {
		log.Fatal("Missing Pinboard API token")
	}

	config, err := whosonfirst.NewDefaultConfig()

	if err != nil {
		log.Fatal(err)
	}

	config.DataRoot = *data_root
	config.AuthToken = *auth_token

	handler := func(rsp http.ResponseWriter, req *http.Request) {

		query := req.URL.Query()

		to_archive := query.Get("url")
		str_id := query.Get("wof_id")

		opts := sanitize.DefaultOptions()

		var err error

		to_archive, err = sanitize.SanitizeString(to_archive, opts)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		str_id, err = sanitize.SanitizeString(str_id, opts)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		wof_id, err := strconv.ParseInt(str_id, 10, 64)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		bm, err := whosonfirst.BookmarkURL(to_archive, wof_id, config)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		enc_bm, err := json.Marshal(bm)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-Type", "text/json")
		rsp.Write(enc_bm)
		return
	}

	pong := func(rsp http.ResponseWriter, req *http.Request) {

		rsp.Header().Set("Content-Type", "text/plain")
		rsp.Write([]byte("PONG"))
	}

	endpoint := fmt.Sprintf("%s:%d", *host, *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/ping", pong)

	err = gracehttp.Serve(&http.Server{Addr: endpoint, Handler: mux})

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
