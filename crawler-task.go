package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type urlData struct {
	url  string
	body string
	urls []string
	err  error
}

type Crawler struct {
	crawledUrls map[string]bool
	crawlChan   chan urlData
	fetcher     Fetcher
}

func NewCrawler(fetcher Fetcher) *Crawler {
	var crawler *Crawler = new(Crawler)
	crawler.crawledUrls = make(map[string]bool)
	crawler.crawlChan = make(chan urlData)
	crawler.fetcher = fetcher
	return crawler
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func (cr *Crawler) Crawl(url string, depth int) {
	go cr.CrawlIt(url, depth)
	var runningCount int = 1
	for runningCount > 0 {
		//get crawled data
		data := <-cr.crawlChan
		runningCount--

		if data.err != nil {
			fmt.Println(data.err)
		} else {
			cr.crawledUrls[data.url] = true
			fmt.Printf("found: %s %q\n", data.url, data.body)
		}
		for _, u := range data.urls {
			go cr.CrawlIt(u, depth-1)
			runningCount++
		}
	}
}

func (cr *Crawler) CrawlIt(url string, depth int) {
	if depth <= 0 {
		return
	}
	if cr.crawledUrls[url] {
		err := fmt.Errorf("skip crawled: %s", url)
		cr.crawlChan <- urlData{url, "", nil, err}
		return
	}
	body, urls, err := cr.fetcher.Fetch(url)
	cr.crawlChan <- urlData{url, body, urls, err}
	return
}

func CrawlOrig(url string, depth int, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		CrawlOrig(u, depth-1, fetcher)
	}
	return
}

func main() {
	var crawler = NewCrawler(fetcher)
	crawler.Crawl("http://golang.org/", 4)
	fmt.Println("Run orig crawl")
	CrawlOrig("http://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
