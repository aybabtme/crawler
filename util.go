package crawler

import (
	"net/url"
)

// Set of string

type stringSet struct {
	list []string
	set  map[string]struct{}
}

func newStringSet() *stringSet {
	return &stringSet{
		list: make([]string, 0),
		set:  make(map[string]struct{}),
	}
}

func (s *stringSet) Contains(q string) bool { _, ok := s.set[q]; return ok }
func (s *stringSet) Slice() []string        { return s.list }
func (s *stringSet) Add(q string) {
	if !s.Contains(q) {
		s.list = append(s.list, q)
		s.set[q] = struct{}{}
	}
}

// Queue of URLs

type urlQueue struct {
	// GC of slices as they grow makes them a great base for (de)queues,
	// with a bit more memory usage than linked list, but simpler type
	// safe use, and also faster.
	//
	// see http://www.antoine.im/posts/someone_was_right_on_the_internet
	vec []*url.URL
}

func (q *urlQueue) IsEmpty() bool  { return len(q.vec) == 0 }
func (q *urlQueue) Add(s *url.URL) { q.vec = append(q.vec, s) }
func (q *urlQueue) Len() int       { return len(q.vec) }
func (q *urlQueue) Remove() *url.URL {
	d := q.vec[0]
	q.vec = q.vec[1:]
	return d
}

// Helper to query HTML elements

type resourceLocator struct {
	element string
	attr    string
}

func (r *resourceLocator) cssSelector() string {
	return r.element + "[" + r.attr + "]"
}

var htmlResources = []resourceLocator{
	{"a", "href"},
	{"area", "href"},
	{"base", "href"},
	{"link", "href"},

	{"audio", "src"},
	{"embed", "src"},
	{"iframe", "src"},
	{"img", "src"},
	{"input", "src"},
	{"script", "src"},
	{"source", "src"},
	{"track", "src"},
	{"video", "src"},

	{"blockquote", "cite"},
	{"del", "cite"},
	{"ins", "cite"},
	{"q", "cite"},

	{"code", "applet"},
	{"codebase", "applet"},

	{"object", "data"},

	{"html", "manifest"},

	{"video", "poster"},
}
