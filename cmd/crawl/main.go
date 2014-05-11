package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aybabtme/crawler"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"
	"time"
)

var (
	agent           = "Crawler/abuse:github.com/aybabtme/crawler"
	defaultFilename = fmt.Sprintf("site_map_%s.json", time.Now().Format("2006-01-02-15-04"))
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [opts]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func perror(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	usage()
}

func main() {

	host := flag.String("h", "", "host to crawl")
	filename := flag.String("f", defaultFilename, "file where to write sitemap, will truncate if already exists")
	flag.Usage = usage
	flag.Parse()

	switch {
	case *host == "":
		perror("missing host name\n")
		flag.PrintDefaults()
	}

	// use all cores by default
	runtime.GOMAXPROCS(runtime.NumCPU())

	hostURL, err := url.Parse(*host)
	if err != nil {
		perror("invalid host: %v\n", err)
	}

	start := time.Now()
	log.Printf("starting crawl on %v", hostURL.String())
	defer func() { log.Printf("done in %v", time.Since(start)) }()

	c, err := crawler.NewCrawler(hostURL, agent)
	if err != nil {
		log.Fatalf("[error] creating crawler, %v", err)
	}

	dig, err := c.Crawl()
	if err != nil {
		log.Fatalf("[error] during crawl, %v", err)
	}

	log.Printf("preparing sitemap")

	data, err := json.MarshalIndent(dig, "", "    ")
	if err != nil {
		log.Fatalf("[error] marshaling sitemap to JSON, %v", err)
	}

	log.Printf("saving to %q", *filename)
	if err := ioutil.WriteFile(*filename, data, 0666); err != nil {
		log.Fatalf("[error] writing sitemap to file, %v", err)
	}

}
