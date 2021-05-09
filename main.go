package main

import (
	"fmt"
	"net/url"
	"spider-go/spider"
)

func main() {
	// parse command line args to build the crawler configuration
	conf, url := spider.ParseFlags()

	eventCh := make(chan spider.Event)
	spider := spider.NewSpider(conf, eventCh)

	go spider.Crawl(url)

	runEventLoop(eventCh)
}

// Runs the event loop to handle all the crawlers events.
func runEventLoop(eventCh chan spider.Event) {
	// parse all crawlers events
	for event := range eventCh {
		switch event := event.(type) {
		case *spider.InvalidArg:
			fmt.Printf("\n[%v] Crawling '%s' invalid arg: %v\n", event.When().Unix(), event.Id(), event.Error())
		case *spider.CrawlingStarted:
			fmt.Printf("\n[%v] Crawling '%s' started\n", event.When().Unix(), event.Id())
		case *spider.PageVisited:
			u, _ := url.PathUnescape(event.Page.Url.String())
			fmt.Printf("\n[%v] Visited '%v'\n", event.When().Unix(), u)
			if event.Page.Err == nil {
				// log all the URLs that belong to the same domain found in this page
				for u := range *event.Page.Urls {
					u, _ = url.PathUnescape(u)
					fmt.Printf("\t%s\n", u)
				}
			}
		case *spider.CrawlingEnded:
			fmt.Printf("\n[%v] Crawling '%s' ended! Visited %d unique URLs\n", event.When().Unix(), event.Id(), event.VisitedCount)
		default:
			fmt.Printf("\nUnhandled Event %T!\n", event)
		}
	}
}
