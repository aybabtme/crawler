package crawler

import (
	"encoding/json"
	"testing"
)

type edge struct{ from, to string }

var nilCode = -1

var digraphTT = []struct {
	name          string
	resourceCount int
	input         []edge
	invalids      map[string]int
}{
	{
		name:          "small tree",
		resourceCount: 3,
		input: []edge{
			{from: "a", to: "b"},
			{from: "b", to: "c"},
		},
		invalids: map[string]int{"c": nilCode},
	},
	{
		name:          "small digraph",
		resourceCount: 3,
		input: []edge{
			{from: "a", to: "b"},
			{from: "b", to: "c"},
			{from: "a", to: "c"},
		},
		invalids: map[string]int{"c": nilCode},
	},
	{
		name:          "small digraph with cycle",
		resourceCount: 3,
		input: []edge{
			{from: "a", to: "b"},
			{from: "b", to: "c"},
			{from: "c", to: "a"},
		},
		invalids: map[string]int{},
	},
	{
		name:          "small digraph with duplicate edges",
		resourceCount: 3,
		input: []edge{
			{from: "a", to: "b"},
			{from: "b", to: "c"},
			{from: "c", to: "a"},
		},
		invalids: map[string]int{},
	},
	{
		name:          "small tree with one invalid",
		resourceCount: 4,
		input: []edge{
			{from: "a", to: "b"},
			{from: "b", to: "c"},
			{from: "c", to: "d"},
		},
		invalids: map[string]int{"d": nilCode},
	},
	{
		name:          "small digraph and one invalid",
		resourceCount: 4,
		input: []edge{
			{from: "a", to: "b"},
			{from: "a", to: "c"},
			{from: "b", to: "c"},
			{from: "c", to: "d"},
		},
		invalids: map[string]int{"d": nilCode},
	},
	{
		name:          "small digraph with cycle and one invalid",
		resourceCount: 4,
		input: []edge{
			{from: "a", to: "b"},
			{from: "b", to: "c"},
			{from: "c", to: "a"},
			{from: "c", to: "d"},
		},
		invalids: map[string]int{"d": nilCode},
	},
	{
		name:          "small complete graph",
		resourceCount: 4,
		input: []edge{
			{from: "a", to: "b"},
			{from: "a", to: "c"},
			{from: "a", to: "d"},
			{from: "b", to: "a"},
			{from: "b", to: "c"},
			{from: "b", to: "d"},
			{from: "c", to: "a"},
			{from: "c", to: "b"},
			{from: "c", to: "d"},
			{from: "d", to: "a"},
			{from: "d", to: "b"},
			{from: "d", to: "c"},
		},
		invalids: map[string]int{},
	},
	{
		name:          "short list",
		resourceCount: 11,
		input: []edge{
			{from: "0", to: "1"},
			{from: "1", to: "2"},
			{from: "2", to: "3"},
			{from: "3", to: "4"},
			{from: "4", to: "5"},
			{from: "5", to: "6"},
			{from: "6", to: "7"},
			{from: "7", to: "8"},
			{from: "8", to: "9"},
			{from: "9", to: "10"},
		},
		invalids: map[string]int{"10": nilCode},
	},
}

func TestEmptyDigraphIsEmpty(t *testing.T) {
	for _, tt := range digraphTT {
		t.Logf("==== Digraph: %s ====", tt.name)

		dig := newDigraph()

		check(t, dig.ResourceCount() == 0, "new digraph should have no resources, got %d", dig.ResourceCount())
		check(t, dig.LinkCount() == 0, "new digraph should have no edges, got %d", dig.LinkCount())

		for _, notThere := range tt.input {
			check(t, !dig.Contains(notThere.from), "should not contain resources, got (from) %q", notThere.from)
			check(t, !dig.Contains(notThere.to), "should not contain resources, got (to) %q", notThere.to)
			check(t, !dig.MarkStatus(notThere.from, 200), "should not be able to invalidate unknown resource (from) %q", notThere.from)
			check(t, !dig.MarkStatus(notThere.to, 200), "should not be able to invalidate unknown resource (to) %q", notThere.to)
		}

		t.Log("ok!")
	}
}

func TestDigraphCanAddEdges(t *testing.T) {
	for _, tt := range digraphTT {
		t.Logf("==== Digraph: %s ====", tt.name)

		dig := newDigraph()
		for _, edge := range tt.input {
			wantLinkCount := dig.LinkCount() + 1
			ok := dig.AddEdge(edge.from, edge.to)
			check(t, ok, "should be able to add edges for first time")
			check(t, dig.Contains(edge.from), "should contain resource (from) %q", edge.from)
			check(t, dig.Contains(edge.to), "should contain resource (to) %q", edge.to)

			gotLinkCount := dig.LinkCount()
			check(t, wantLinkCount == gotLinkCount, "want edge size %d, got %d", wantLinkCount, gotLinkCount)
		}

		wantResCount := tt.resourceCount
		gotResCount := dig.ResourceCount()
		check(t, wantResCount == gotResCount, "want resource size %d, got %d", wantResCount, gotResCount)

		wantLinkCount := len(tt.input)
		gotLinkCount := dig.LinkCount()
		check(t, wantLinkCount == gotLinkCount, "want edge all %d edges accounted for, got %d", wantLinkCount, gotLinkCount)

		for invalid, code := range tt.invalids {
			dig.MarkStatus(invalid, code)
		}

		gotResCount = dig.ResourceCount()
		check(t, wantResCount == gotResCount, "after invalidating: want resource size %d, got %d", wantResCount, gotResCount)

		gotLinkCount = dig.LinkCount()
		check(t, wantLinkCount == gotLinkCount, "after invalidating: want edge all %d edges accounted for, got %d", wantLinkCount, gotLinkCount)

		t.Log("ok!")
	}
}

func TestDigraphCanInvalidateLinks(t *testing.T) {
	for _, tt := range digraphTT {
		t.Logf("==== Digraph: %s ====", tt.name)

		dig := newDigraph()
		for _, edge := range tt.input {
			dig.AddEdge(edge.from, edge.to)
			dig.MarkStatus(edge.from, 200)
		}

		for invalid, code := range tt.invalids {
			check(t, dig.MarkStatus(invalid, code), "should be able to invalidate resource %q", invalid)
			check(t, dig.Contains(invalid), "should contain resource %q even if invalidated", invalid)
		}

		t.Log("ok!")
	}
}

func TestDigraphCanStopWalkEarly(t *testing.T) {
	for _, tt := range digraphTT {
		t.Logf("==== Digraph: %s ====", tt.name)

		dig := newDigraph()
		for _, edge := range tt.input {
			dig.AddEdge(edge.from, edge.to)
			dig.MarkStatus(edge.from, 200)
		}

		for invalid, code := range tt.invalids {
			dig.MarkStatus(invalid, code)
		}

		wantWalkLen := tt.resourceCount / 2
		gotWalkLen := 0
		dig.Walk(func(_ string, _ int, _, _ []string) bool {
			gotWalkLen++
			return gotWalkLen != wantWalkLen
		})
		check(t, wantWalkLen == gotWalkLen, "should not walk more than requested, want walk of %d, got %d", wantWalkLen, gotWalkLen)
		t.Log("ok!")
	}
}

func TestDigraphWalkDoesntRepeatLinks(t *testing.T) {
	for _, tt := range digraphTT {
		t.Logf("==== Digraph: %s ====", tt.name)

		dig := newDigraph()
		for _, edge := range tt.input {
			dig.AddEdge(edge.from, edge.to)
			dig.MarkStatus(edge.from, 200)
		}

		for invalid, code := range tt.invalids {
			dig.MarkStatus(invalid, code)
		}

		alreadySeen := make(map[string]struct{})
		dig.Walk(func(link string, _ int, _, _ []string) bool {
			_, seen := alreadySeen[link]
			check(t, !seen, "should not walk over %q multiple times", link)
			alreadySeen[link] = struct{}{}
			return true
		})
		t.Log("ok!")
	}
}

func TestDigraphWalkReportsLinkValidity(t *testing.T) {
	for _, tt := range digraphTT {
		t.Logf("==== Digraph: %s ====", tt.name)

		dig := newDigraph()
		for _, edge := range tt.input {
			dig.AddEdge(edge.from, edge.to)
			dig.MarkStatus(edge.from, 200)
		}

		for invalid, code := range tt.invalids {
			dig.MarkStatus(invalid, code)
		}

		dig.Walk(func(link string, code int, refersTo, referedBy []string) bool {

			wantCode, wantInvalid := tt.invalids[link]
			if wantInvalid {
				check(t, code == wantCode, "want link %q to be invalid, but was not", link)
			} else {
				check(t, code != nilCode, "want link %q to be valid, but was not", link)
			}

			return true
		})

		t.Log("ok!")
	}
}

func TestDigraphWalkReportsReferences(t *testing.T) {
	for _, tt := range digraphTT {
		t.Logf("==== Digraph: %s ====", tt.name)

		dig := newDigraph()
		for _, edge := range tt.input {
			dig.AddEdge(edge.from, edge.to)
			dig.MarkStatus(edge.from, 200)
		}

		for invalid, code := range tt.invalids {
			dig.MarkStatus(invalid, code)
		}

		dig.Walk(func(link string, _ int, refersTo, referedBy []string) bool {

			verifyOutRef(t, tt.input, link, refersTo)
			verifyInRef(t, tt.input, link, referedBy)

			return true
		})

		t.Log("ok!")
	}
}

func TestDigraphCanMarshalToJSON(t *testing.T) {
	for _, tt := range digraphTT {
		t.Logf("==== Digraph: %s ====", tt.name)

		dig := newDigraph()
		for _, edge := range tt.input {
			dig.AddEdge(edge.from, edge.to)
			dig.MarkStatus(edge.from, 200)
		}

		for invalid, code := range tt.invalids {
			dig.MarkStatus(invalid, code)
		}

		data, err := json.Marshal(dig)
		check(t, err == nil, "should be able to marshal to JSON, but got error: %v", err)

		// don't unpack to a struct, as it will silently fail. we want panic!
		var unmarsh interface{}
		err = json.Unmarshal(data, &unmarsh)
		check(t, err == nil, "encoding/json error: %v", err)

		gotDig := unmarsh.(map[string]interface{})
		gotlinkCount := int(gotDig["link_count"].(float64))
		gotResCount := int(gotDig["resource_count"].(float64))
		gotResRawCount := len(gotDig["resources"].(map[string]interface{}))

		check(t, gotlinkCount == len(tt.input), "want link count %d, got %d", len(tt.input), gotlinkCount)
		check(t, gotResCount == tt.resourceCount, "want resource count %d, got %d", tt.resourceCount, gotlinkCount)
		check(t, gotResRawCount == tt.resourceCount, "want %d resource description, got %d", tt.resourceCount, gotResRawCount)

		t.Log("ok!")
	}
}

// verifies that all claimed references reachable by `link` are expected, and that
// all those expected are reported
func verifyOutRef(t *testing.T, knownEdges []edge, link string, outRef []string) {

	// it should be true that all pretented out-references are actual references
	for _, out := range outRef {
		found := false
		for _, want := range knownEdges {
			if want.from == link && want.to == out {
				found = true
			}
		}
		check(t, found, "%q should have out reference, but none was reported", link)
	}

	// it should be true that all actual out-references are returned as such
	for _, e := range knownEdges {
		if e.from != link {
			continue
		}
		found := false
		for _, out := range outRef {
			if e.to == out {
				found = true
			}
		}
		check(t, found, "%q should have out reference, but none was reported", link)
	}

}

// verifies that all claimed references reachable from `link` are expected, and that
// all those expected are reported
func verifyInRef(t *testing.T, knownEdges []edge, link string, inRef []string) {

	// it should be true that all pretented in-references are actual references
	for _, in := range inRef {
		found := false
		for _, want := range knownEdges {
			if want.from == in && want.to == link {
				found = true
			}
		}
		check(t, found, "%q should have in-reference, but none was reported", link)
	}

	// it should be true that all actual in-references are returned as such
	for _, e := range knownEdges {
		if e.to != link {
			continue
		}
		found := false
		for _, in := range inRef {
			if e.from == in {
				found = true
			}
		}
		check(t, found, "%q should have in-reference, but none was reported", link)
	}
}
