// Package entry point. The Spider component allows to asynchronously crawl a
// single domain, while communicating the outcome via predefined events using
// channels.
package spider

import (
	"flag"
	"fmt"
	"net/url"
	"time"
)

const (
	// Default URL to crawl.
	defaultUrl = "https://monzo.com/"
	// Default maximum number of URLs visited concurrently.
	defaultConcurrency = 64
	// Default HTTP request timeout in seconds.
	defaultTimeout = 5
)

/// The crawler general settings.
type SpiderConf struct {
	/// Number URLs visited concurrently from the frontier set.
	concurrency int
	// HTTP request timeout.
	timeout time.Duration
}

// Parses the CLI arguments to build the crawler configuration.
// Returns the configuration and the URL that needs to be crawled.
func ParseFlags() (*SpiderConf, string) {
	// URL to crawl
	url := flag.String("url", defaultUrl, "URL to crawl")
	// concurrency level
	concurrency := flag.Int("concurrency", defaultConcurrency, "Max number of URLs visited concurrently")
	// HTTP request timeout
	t := flag.Int("timeout", defaultTimeout, "HTTP request timeout in seconds")
	timeout := time.Duration(*t) * time.Second

	flag.Parse()

	return &SpiderConf{*concurrency, timeout}, *url
}

// The Spider in charge of crawling a single domain and communicating the crawling
// status asynchronously.
type Spider struct {
	// The crawler configuration.
	conf *SpiderConf
	// The channel used to communicate main crawling events.
	eventCh chan Event
}

// Constructs a new HTML crawler for a single domain.
func NewSpider(conf *SpiderConf, eventCh chan Event) *Spider {
	return &Spider{conf, eventCh}
}

// Starts crawling the original URL, and uses the channel to communicate via events.
func (spider *Spider) Crawl(rawUrl string) {
	rootUrl, err := url.Parse(rawUrl)
	if err != nil {
		err = fmt.Errorf("the given URL '%s' is not valid: %s", rawUrl, err)
		spider.eventCh <- newInvalidArg(rawUrl, err)
		return
	}

	spider.eventCh <- newCrawlingStarted(rootUrl.Hostname())
	visited := spider.run(rootUrl)
	spider.eventCh <- newCrawlingEnded(rootUrl.Hostname(), len(*visited))
}

// Run the crawling logic. Returns the visited URLs map.
func (spider *Spider) run(rootUrl *url.URL) *urlsMap {
	// the crawler implementation
	crawler := newCrawler(rootUrl, spider.conf.timeout)
	// the URLs left to visit (DFS)
	frontier := []*url.URL{rootUrl}
	// the set of URLs already visited
	visited := make(urlsMap)

	// channel used to asynchronously crawl the URLs in the frontiers
	pageCh := make(chan pageUrls, spider.conf.concurrency)
	eventId := rootUrl.Hostname()

	for len(frontier) > 0 {
		//fmt.Println("\nFrontier size:", len(frontier))
		//fmt.Println("Visited size:", len(visited))

		// limit URLs to visit per step
		toVisitCount := min(len(frontier), spider.conf.concurrency)
		//fmt.Printf("Visiting %d URLs\n", toVisitCount)

		// concurrent crawling a subset of the frontier
		for i := 0; i < toVisitCount; i++ {
			go crawler.crawl(frontier[i], pageCh)
		}
		frontier = frontier[toVisitCount:]

		toVisitNext := make(urlsMap)
		// current frontier set to avoid inserting duplicates
		// Note: this is a compromise that favors runtime over memory usage
		frontierSet := make(urlsSet, len(frontier))
		for _, u := range frontier {
			key, _ := url.PathUnescape(u.String())
			frontierSet[key] = true
		}

		// fetch crawling results
		for i := 0; i < toVisitCount; i++ {
			pageRes := <-pageCh

			// mark and signal that this page has been visited
			visited[pageRes.Url.String()] = pageRes.Url
			spider.eventCh <- newPageVisited(eventId, &pageRes)

			// store the set of URLs that will be part of the frontier at the next step
			if pageRes.Err == nil {
				for key, url := range *pageRes.Urls {
					// not visited yet
					if _, ok := visited[key]; !ok {
						// not in frontier already
						if _, ok := frontierSet[key]; !ok {
							toVisitNext[key] = url
						}
					}
				}
			}
		}

		// expand the frontiers with the new URLs to visit
		for _, url := range toVisitNext {
			frontier = append(frontier, url)

			// TODO: remove in the end
			checkDups(frontier)
		}
	}

	return &visited
}

// TODO delete
func checkDups(urls []*url.URL) {
	urlsset := make(urlsSet)
	for _, u := range urls {
		key, _ := url.PathUnescape(u.String())
		urlsset[key] = true
	}

	if len(urls) != len(urlsset) {
		fmt.Println(urls)
		panic("DUPLICATES!")
	}
}

// Returns the minimum value between the 2 given integers.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
