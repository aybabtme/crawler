package crawler

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

var (
	testAgent = "CrawlerUnderTest/abuse:github.com/aybabtme/crawler"
	testPath  = "testdata/antoine.im/"

	e              = struct{}{}
	filesNotToFind = map[string]struct{}{
		"/robots.txt":                                    e,
		"/sitemap.xml":                                   e,
		"/not_allowed_by_robots.txt":                     e,
		"/this_file_should_not_be_found.html":            e,
		"/posts/this_file_should_not_be_found.html":      e,
		"/assets/css/this_file_should_not_be_found.html": e,
	}
	errorsCases = map[string]int{
		"/400.html": 400,
		"/401.html": 401,
		"/402.html": 402,
		"/403.html": 403,
		"/404.html": 404,
		"/500.html": 500,
		"/501.html": 501,
		"/502.html": 502,
		"/503.html": 503,
	}
)

func TestCanCreateCrawler(t *testing.T) {
	withDomain(t, func(domain *url.URL) {
		_, err := NewCrawler(domain, testAgent)
		check(t, err == nil, "should be able to create crawler, got error: %v", err)
	})
}

func TestCanCrawl(t *testing.T) {
	withCrawler(t, func(_ *url.URL, c Crawler) {
		_, err := c.Crawl()
		check(t, err == nil, "should be able to crawler, got error: %v", err)
	})
}

func TestGraphContainsAllResources(t *testing.T) {
	withResourceGraph(t, func(g ResourceGraph, truth map[string]struct{}) {

		for want := range truth {
			check(t, g.Contains(want), "crawler should know of %q", want)
		}
	})
}

func TestGraphHasNoExtraResources(t *testing.T) {
	withResourceGraph(t, func(g ResourceGraph, truth map[string]struct{}) {

		g.Walk(func(link string, status int, _, _ []string) bool {

			_, foundIt := filesNotToFind[link]
			check(t, !foundIt, "graph should not have known of %q", link)

			_, shouldKnown := truth[link]
			if !shouldKnown {
				check(t, !shouldKnown, "graph should not know about %q", link)
				t.Logf("ok: graph doesnt know of %q", link)
			} else {
				check(t, shouldKnown, "graph should know about %q", link)
				t.Logf("ok: graph does know of %q", link)
			}

			return true
		})
	})
}

// context providers

// basic test harness starts a fake server to crawl, provides
// f with the domain URL to that fake server
func withDomain(t *testing.T, f func(*url.URL)) {
	route := mux.NewRouter()
	for path, code := range errorsCases {
		errorPath := path
		errorCode := code
		route.HandleFunc(errorPath, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(errorCode)
		})
	}

	route.NotFoundHandler = http.FileServer(http.Dir(testPath))
	server := httptest.NewServer(route)

	defer server.Close()

	domain, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("bad server URL, %v", err)
	}
	f(domain)
}

// builds on withDomain, returns a crawler ready to start
func withCrawler(t *testing.T, f func(*url.URL, Crawler)) {
	withDomain(t, func(domain *url.URL) {
		c, err := NewCrawler(domain, testAgent)
		if err != nil {
			t.Fatalf("couldn't create crawler, %v", err)
		}
		f(domain, c)
	})
}

// builds on withCrawler, returns a ResourceGraph resulting
// from a crawler.Crawl, and a set that represents the ground
// truth of files served by the fake server.
func withResourceGraph(t *testing.T, f func(ResourceGraph, map[string]struct{})) {
	withCrawler(t, func(domain *url.URL, c Crawler) {
		g, err := c.Crawl()
		if err != nil {
			t.Fatalf("coudn't prepare test crawl, %v", err)
		}

		groundTruth := make(map[string]struct{})

		err = filepath.Walk(testPath, func(path string, fi os.FileInfo, err error) error {
			if fi.IsDir() {
				return err
			}

			path = path[len(testPath):]

			u, err := cleanFromURLString(domain, path)
			groundTruth[u.String()] = struct{}{}
			return err
		})
		if err != nil {
			t.Fatalf("coudn't walk test path, %v", err)
		}

		for toIgnore := range filesNotToFind {
			u, err := cleanFromURLString(domain, toIgnore)
			if err != nil {
				t.Fatalf("coudn't parse URLs not to fine, %v", err)
			}
			if _, ok := groundTruth[u.String()]; !ok {
				t.Fatalf("mismatch, cant remove %q from groundTruth", u.String())
			}
			delete(groundTruth, u.String())
		}

		u, err := cleanFromURLString(domain, "/")
		if err != nil {
			t.Fatalf("coudn't parse root URL, %v", err)
		}

		groundTruth[u.String()] = struct{}{}

		f(g, groundTruth)
	})
}
