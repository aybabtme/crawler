# Crawler

[![Build Status](https://drone.io/github.com/aybabtme/crawler/status.png)](https://drone.io/github.com/aybabtme/crawler/latest)
[![Coverage Status](https://img.shields.io/coveralls/aybabtme/crawler.svg)](https://coveralls.io/r/aybabtme/crawler)

A simple domain crawler.

* Respects `robots.txt`.
* Doesn't leave the domain it's given.
* Doesn't visit sub-domains.

# Crawl things!

Install the crawler:
```
go get github.com/aybabtme/crawler/cmd/crawl
```

Use it:

```
crawl -h http://antoine.im -f antoineim_map.json
```

Should print things like:

```
2014/05/11 02:28:04 starting crawl on http://antoine.im
2014/05/11 02:28:05 [crawler] root has 1 elements
2014/05/11 02:28:05 [crawler] fringe=10 found=12 (new=10, rejected=2) source="http://antoine.im"
2014/05/11 02:28:06 [crawler] fringe=10 found=27 (new=1, rejected=20) source="http://antoine.im/posts/someone_was_right_on_the_internet"
...
2014/05/11 02:28:07 [crawler] fringe=0  found=0 (new=0, rejected=0) source="http://antoine.im/assets/data/to_buffer_or_not_to_buffer/t1_micro_bench_1.0MB.svg"
2014/05/11 02:28:07 [crawler] done crawling, 15 resources, 45 links
2014/05/11 02:28:07 preparing sitemap
2014/05/11 02:28:07 saving to "antoineim_map.json"
2014/05/11 02:28:07 done in 3.006155429s
```

# Use the lib!

If you want to use the library.

```
go get github.com/aybabtme/crawler
```

The godocs are on [godoc](http://godoc.org/github.com/aybabtme/crawler) (lol).

# Test it!

```
go get -t github.com/aybabtme/crawler
make test
```

To view the coverage report:

```
make cover
```

# The output

The output of a crawl is a list of resources, along with:

* Where they refer to (points to something).
* Where are they are refered from (something points to that).
* What was the status code of reaching this resource.

The status code is interesting: it might show that you have dead links (404),
for instance.

Here's a snippet of crawling my [blog](http://antoine.im). The full map can be
found [here](sample_map.json) if you want to see it.

```json
{
    "resource_count": 15,
    "link_count": 45,
    "resources": [
        {
            "url": "http://antoine.im/posts/dynamic_programming_for_the_lazy",
            "refered_by": [
                "http://antoine.im"
            ],
            "refers_to": [
                "http://antoine.im",
                "http://antoine.im/assets/css/brog.css",
                "http://antoine.im/assets/css/font-awesome.min.css",
                "http://antoine.im/assets/css/styles/github.css",
                "http://antoine.im/assets/js/algo_convenience_hacks.js",
                "http://antoine.im/assets/js/brog.js"
            ],
            "status_code": 200
        },
        {
            "url": "http://antoine.im/assets/css/brog.css",
            "refered_by": [
                "http://antoine.im",
                "http://antoine.im/posts/someone_was_right_on_the_internet",
                "http://antoine.im/posts/someone_is_wrong_on_the_internet",
                "http://antoine.im/posts/the_story_of_select_and_the_goroutines",
                "http://antoine.im/posts/dynamic_programming_for_the_lazy",
                "http://antoine.im/posts/to_buffer_or_not_to_buffer",
                "http://antoine.im/posts/correction_hacks"
            ],
            "refers_to": [],
            "status_code": 200
        },
        // ...
    }
}
```
