package crawler

import (
	"code.google.com/p/go.net/html"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/PuerkitoBio/purell"
	"github.com/temoto/robotstxt-go"
	"log"
	"mime"
	"net/http"
	"net/url"
)

// Crawler produces a directed graphs of links within a single domain.
type Crawler interface {
	Crawl() (ResourceGraph, error)
}

// NewCrawler creates a single threaded Crawler that respects robots.txt,
// starting with domain and all robots.txt SiteMap links on the same domain
// for root values.
//
// Crawler will write to standard log as it progresses.
func NewCrawler(domain *url.URL, agent string) (Crawler, error) {
	base, err := cleanFromURLString(domain, "/")
	if err != nil {
		return nil, err
	}

	robot, err := makeRobot(domain)
	if err != nil {
		return nil, err
	}

	return &crawler{
		resources: htmlResources,
		base:      base,
		robot:     robot,
		group:     robot.FindGroup(agent),
		agent:     agent,
	}, err
}

func makeRobot(host *url.URL) (r *robotstxt.RobotsData, err error) {
	robotURL, err := host.Parse("/robots.txt")
	if err != nil {
		return nil, fmt.Errorf("parsing robots.txt URL, %v", err)
	}

	resp, err := http.Get(robotURL.String())
	if err != nil {
		return nil, fmt.Errorf("retrieving robots.txt, %v", err)
	}
	defer func() { err = resp.Body.Close() }()

	return robotstxt.FromResponse(resp)
}

type crawler struct {
	resources []resourceLocator
	base      *url.URL
	robot     *robotstxt.RobotsData
	group     *robotstxt.Group
	agent     string
}

func (c *crawler) Crawl() (ResourceGraph, error) {

	dig := newDigraph()

	var (
		fringe    urlQueue
		followers []*url.URL
		status    int
		err       error
	)

	for _, root := range c.findRoots() {
		fringe.Add(root)
	}

	log.Printf("[crawling] root has %d elements", fringe.Len())

	for !fringe.IsEmpty() {
		link := fringe.Remove()

		followers, status, err = c.generateFollowers(link)
		if err != nil {
			log.Printf("[crawling] error: %v", err)
			continue
		}
		dig.MarkStatus(link.String(), status)

		if status >= 400 {
			log.Printf("[crawling] invalid link: %q", link.String())
			continue
		}

		// log.Printf("[crawling] %d in fringe, %d followers to link %q", fringe.Len(), len(followers), link.String())

		for _, follow := range followers {
			if !c.isAcceptable(follow) {
				continue
			}
			if !dig.Contains(follow.String()) {
				fringe.Add(follow)
			}
			dig.AddEdge(link.String(), follow.String())
		}

	}

	log.Printf("[crawling] done crawling, %d resources, %d links", dig.ResourceCount(), dig.LinkCount())
	return dig, err
}

func (c *crawler) findRoots() []*url.URL {
	urls := []*url.URL{c.base}

	roots := newStringSet()
	roots.Add(c.base.String())

	for _, site := range c.robot.Sitemaps {
		dirtyU, err := url.Parse(site)
		if err != nil {
			continue
		}
		u, err := cleanFromURLString(dirtyU, "")
		if err != nil {
			log.Printf("Parsing sitemap URL %q: %v", site, err)
			continue
		}
		if u.Host != c.base.Host {
			log.Printf("Wrong sitemap domain: %q", site)
			continue
		}

		if roots.Contains(u.String()) {
			log.Printf("Duplicate in sitemap: %q", site)
			continue
		}
		roots.Add(u.String())
		urls = append(urls, u)
	}
	return urls
}

func (c *crawler) isAcceptable(u *url.URL) bool {

	if u.Host != c.base.Host {
		return false
	}

	allowed := c.group.Test(u.Path)
	if !allowed {
		return false
	}

	return true
}

func (c *crawler) generateFollowers(from *url.URL) (followers []*url.URL, status int, err error) {
	// use named return values to catch the resp.Body.Close() error

	req, err := http.NewRequest("GET", from.String(), nil)
	if err != nil {
		return nil, -1, err
	}
	req.Header.Add("User-Agent", c.agent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer func() { err = resp.Body.Close() }()

	status = resp.StatusCode

	switch status / 100 {
	case 4:
		log.Printf("[Client error] %s", resp.Status)
		return
	case 5:
		log.Printf("[Server error] %s", resp.Status)
		return
	}

	// the link exists/is usable (not 4xx/5xx)

	mediatype, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return
	}

	switch mediatype {
	// only try to find links in HTML, or perhaps XML documents
	case "text/html",
		"application/atom+xml",
		"text/xml",
		"text/plain",
		"image/svg+xml":
	default: // ignore everything else
		return
	}

	node, err := html.Parse(resp.Body)
	if err != nil {
		return
	}

	add := func(link string) {

		u, err := cleanFromURLString(from, link)
		if err == nil {
			followers = append(followers, u)
		}
	}

	doc := goquery.NewDocumentFromNode(node)

	for _, res := range c.resources {
		doc.Find(res.cssSelector()).Each(func(_ int, s *goquery.Selection) {
			val, ok := s.Attr(res.attr)
			if !ok {
				return
			}

			add(val)
		})
	}

	return
}

func cleanFromURLString(from *url.URL, link string) (*url.URL, error) {

	u, err := url.Parse(link)
	if u.Host == "" {
		u.Scheme = from.Scheme
		u.Host = from.Host
	}
	uStr := purell.NormalizeURL(u, purell.FlagsUsuallySafeGreedy)

	clean, err := from.Parse(uStr)

	return clean, err
}
