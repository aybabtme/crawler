# Crawler

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

# Use the lib!

If you want to use the library.

```
go get github.com/aybabtme/crawler
```

The godocs are on [godoc](http://godoc.org/github.com/aybabtme/crawler) (lol).

# Test it!

```
go get -t github.com/aybabtme/crawler
go test github.com/aybabtme/crawler
```
