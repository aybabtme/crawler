package crawler

import (
	"encoding/json"
	"sync"
)

// ResourceGraph represents a website as a directed graph where resources
// are connected by URLs.
type ResourceGraph interface {
	// if a URL is reachable on the website.
	Contains(string) bool
	// the number of resources in the graph.
	ResourceCount() int
	// the number of links interconnecting the resources.
	LinkCount() int
	// walks over the graph as long as the func returns true.
	Walk(func(link string, status int, refersTo, referedBy []string) bool)
	// can be marshalled to JSON.
	json.Marshaler
}

// digraph implements ResourceGraph + extra methods needed by the crawler.
// It is immutable by external users who only see the interface.
type digraph struct {
	// convention: only exported funcs take a lock. Exported funcs
	// never call other exported funcs. This avoid locking yourself
	// in (mutexes aren't reentrant)
	lock  sync.RWMutex
	e     int
	nodes map[string]*resource
}

func newDigraph() *digraph {
	return &digraph{
		e:     0,
		lock:  sync.RWMutex{},
		nodes: make(map[string]*resource),
	}
}

func (d *digraph) Contains(v string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.contains(v)
}

func (d *digraph) AddEdge(v, w string) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.addEdge(v, w)
}

func (d *digraph) MarkStatus(v string, code int) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	node, ok := d.nodes[v]
	if ok {
		node.status = code
	}
	return ok
}

func (d *digraph) ResourceCount() int {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return len(d.nodes)
}

func (d *digraph) LinkCount() int {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.e
}

func (d *digraph) Walk(walker func(string, int, []string, []string) bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	for _, res := range d.nodes {
		wantsMore := walker(
			res.link,
			res.status,
			res.refersTo.Slice(),
			res.referedBy.Slice(),
		)
		if !wantsMore {
			return
		}
	}
}

func (d *digraph) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ResourceCount int                  `json:"resource_count"`
		LinkCount     int                  `json:"link_count"`
		Resources     map[string]*resource `json:"resources"`
	}{d.ResourceCount(), d.LinkCount(), d.nodes})
}

func (d *digraph) contains(v string) bool {
	_, ok := d.nodes[v]
	return ok
}

func (d *digraph) addEdge(from, to string) bool {
	res, ok := d.nodes[from]
	if !ok {
		res = newResource(from)
		d.nodes[from] = res
	}

	res.refersTo.Add(to)

	toRes, ok := d.nodes[to]
	if !ok {
		toRes = newResource(to)
		d.nodes[to] = toRes
	}
	toRes.referedBy.Add(from)

	d.e++
	return true
}

type resource struct {
	referedBy *stringSet
	refersTo  *stringSet
	link      string
	status    int
}

func newResource(link string) *resource {
	return &resource{
		referedBy: newStringSet(),
		refersTo:  newStringSet(),
		link:      link,
		status:    -1,
	}
}

func (r *resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ReferedBy []string `json:"refered_by"`
		RefersTo  []string `json:"refers_to"`
		Status    int      `json:"status_code"`
	}{r.referedBy.Slice(), r.refersTo.Slice(), r.status})
}
