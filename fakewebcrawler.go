package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, end chan bool) {
	// when ending function report that fact
	defer (func () { end <- true })()
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)

	waitForAll := make(chan bool)
	for _, u := range urls {
		go Crawl(u, depth-1, fetcher, waitForAll)
	}

	for i := 0; i < len(urls); i++ {
		<- waitForAll
	}
	return
}

func main() {
	end := make(chan bool)
	go Crawl("http://golang.org/", 4, fetcher, end)
	<- end
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher struct {
	results map[string]*fakeResult
	mutex   *sync.Mutex
}

type fakeResult struct {
	visited bool
	body    string
	urls    []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f.results[url]; ok {
		f.mutex.Lock()
		defer f.mutex.Unlock()
		if res.visited {
			return "", nil, fmt.Errorf("already visited: %s", url)
		}
		res.visited = true
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	map[string]*fakeResult{
		"http://golang.org/": &fakeResult{
			false,
			"The Go Programming Language",
			[]string{
				"http://golang.org/pkg/",
				"http://golang.org/cmd/",
			},
		},
		"http://golang.org/pkg/": &fakeResult{
			false,
			"Packages",
			[]string{
				"http://golang.org/",
				"http://golang.org/cmd/",
				"http://golang.org/pkg/fmt/",
				"http://golang.org/pkg/os/",
			},
		},
		"http://golang.org/pkg/fmt/": &fakeResult{
			false,
			"Package fmt",
			[]string{
				"http://golang.org/",
				"http://golang.org/pkg/",
			},
		},
		"http://golang.org/pkg/os/": &fakeResult{
			false,
			"Package os",
			[]string{
				"http://golang.org/",
				"http://golang.org/pkg/",
			},
		},
	},
	&sync.Mutex{},
}
