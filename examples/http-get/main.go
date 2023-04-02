package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/borud/viassh"
)

var (
	viaHosts   stringArrayFlag
	targetURLs stringArrayFlag
)

func main() {
	flag.Var(&viaHosts, "via", "via host, format: user@host:port")
	flag.Var(&targetURLs, "target", "target URL")
	flag.Parse()

	if len(viaHosts) == 0 {
		log.Fatal("please specify at least one -via hosts")
	}

	if len(targetURLs) == 0 {
		log.Fatal("please specify at least one -target URL")
	}

	// create the viassh Dialer
	dialer, err := viassh.Create(viassh.Config{
		Hosts:  viaHosts,
		Logger: log.Default(),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := dialer.Close()
		if err != nil {
			log.Printf("error closing dialer: %v", err)
		}
	}()

	// Loop through target URLs to demonstrate that we don't need to set up a new
	// dialer for each connection.  We can tunnel multiple connections.
	for _, targetURL := range targetURLs {
		// http.Client is nice enough to provide a Transport config which allows
		// us to plug in our dialer.  If you implement a protocol some time please
		// remember this feature and consider offering it.
		httpClient := http.Client{
			Transport: &http.Transport{
				Dial: dialer.Dial,
			},
		}

		res, err := httpClient.Get(targetURL)
		if err != nil {
			log.Fatal(err)
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		res.Body.Close()

		fmt.Printf("---- [%s]\n%s\n----\n", targetURL, string(data))
	}
}
