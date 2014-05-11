package crawler

import (
	"log"
	"net/url"
	"testing"
)

var stringSetTT = []struct {
	name  string
	input []string
	want  []string
}{
	{
		name:  "slice without duplicates should return all values",
		input: []string{"this", "is", "a", "beautiful", "world"},
		want:  []string{"this", "is", "a", "beautiful", "world"},
	},
	{
		name:  "slice with only duplicates should return single value",
		input: []string{"banana", "banana", "banana", "banana", "banana", "banana", "banana"},
		want:  []string{"banana"},
	},
	{
		name:  "slice with some duplicates should return unique values",
		input: []string{"this", "is", "a", "beautiful", "world", "this", "is", "awesome"},
		want:  []string{"this", "is", "a", "beautiful", "world", "awesome"},
	},
}

func TestStringSet(t *testing.T) {
	for _, tt := range stringSetTT {
		t.Logf("==== String set: %s ====", tt.name)
		set := newStringSet()

		beforeLen := len(set.Slice())
		check(t, beforeLen == 0, "set should be empty, but len=%d", beforeLen)

		for _, in := range tt.input {
			set.Add(in)
			check(t, set.Contains(in), "%q should be in set after adding it", in)
		}

		gotLen := len(set.Slice())
		wantLen := len(tt.want)
		check(t, gotLen == wantLen, "want slice of len %d, got len %d", wantLen, gotLen)

		for _, want := range tt.want {
			found := false
			for _, got := range set.Slice() {
				if got == want {
					found = true
				}
			}
			check(t, found, "want %q in set, but was not found", want)
		}

		t.Log("ok!")
	}
}

var urlQueueTT = []struct {
	name  string
	input []*url.URL
	want  []*url.URL
}{
	{
		name:  "empty queue returns nothing",
		input: URL( /* nothing */),
		want:  URL( /* nothing */),
	},
	{
		name:  "queue with single URL returns only that URL",
		input: URL("http://hello.com"),
		want:  URL("http://hello.com"),
	},
	{
		name:  "queue with many URL returns them all in order",
		input: URL("http://1.com", "http://2.org", "http://3.org", "https://4.im"),
		want:  URL("http://1.com", "http://2.org", "http://3.org", "https://4.im"),
	},
	{
		name:  "queue with duplicate URL returns them all in order, with duplicates still there",
		input: URL("http://hello.com", "http://bye.org", "http://bye.org", "https://antoine.im"),
		want:  URL("http://hello.com", "http://bye.org", "http://bye.org", "https://antoine.im"),
	},
}

func TestURLQueueAddOneRemoveOn(t *testing.T) {
	for _, tt := range urlQueueTT {
		t.Logf("==== URL queue: %s ====", tt.name)
		var q urlQueue
		check(t, q.IsEmpty(), "new queue should be empty")
		beforeLen := q.Len()
		check(t, beforeLen == 0, "new queue should have len=0, but len=%d", beforeLen)

		for _, in := range tt.input {
			q.Add(in)

			check(t, !q.IsEmpty(), "queue should not be empty")
			check(t, 1 == q.Len(), "queue should have grown, want %d but was %d", 1, q.Len())

			got := q.Remove()

			check(t, q.Len() == 0, "queue should have len=0 after removing everything, but len=%d", q.Len())
			check(t, in.String() == got.String(), "removal wants %q, got %q", in.String(), got.String())
			check(t, q.IsEmpty(), "queue should be empty after removing everything")
		}

		t.Log("ok!")
	}
}

func TestURLQueueAllAtOnce(t *testing.T) {
	for _, tt := range urlQueueTT {
		t.Logf("==== URL queue: %s ====", tt.name)
		var q urlQueue
		check(t, q.IsEmpty(), "new queue should be empty")
		beforeLen := q.Len()
		check(t, beforeLen == 0, "new queue should have len=0, but len=%d", beforeLen)

		for _, in := range tt.input {
			beforeLen := q.Len()

			q.Add(in)
			check(t, !q.IsEmpty(), "queue should not be empty")

			wantLen := beforeLen + 1
			gotLen := q.Len()
			check(t, wantLen == gotLen, "queue should have grown, want %d but was %d", wantLen, gotLen)
		}

		for _, want := range tt.want {
			check(t, !q.IsEmpty(), "queue should not be empty yet")

			beforeLen := q.Len()
			got := q.Remove()

			check(t, want.String() == got.String(), "removal wants %q, got %q", want.String(), got.String())

			wantLen := beforeLen - 1
			gotLen := q.Len()
			check(t, wantLen == gotLen, "queue should have shrunk, want %d but was %d", wantLen, gotLen)
		}

		check(t, q.IsEmpty(), "queue should be empty after removing everything")
		afterLen := q.Len()
		check(t, afterLen == 0, "queue should have len=0 after removing everything, but len=%d", afterLen)

		t.Log("ok!")
	}
}

func URL(str ...string) (urls []*url.URL) {
	for _, s := range str {
		u, err := url.Parse(s)
		if err != nil {
			log.Fatalf("%q is not a valid url, %v", str, err)
		}
		urls = append(urls, u)
	}
	return
}

func check(t *testing.T, cond bool, format string, args ...interface{}) {
	if !cond {
		t.Fatalf(""+format, args...)
	}
}
