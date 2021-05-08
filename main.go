package main

import (
	"fmt"
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
			if event.Page.Err == nil {
				// TODO
				// log all the URLs that belong to the same domain found in this page
				fmt.Printf("\n[%v] Visited '%s'", event.When().Unix(), event.Page.Url)
				//for url := range *event.Page.Urls {
				//	fmt.Printf("\t%s\n", url)
				//}
			} else {
				//fmt.Printf("\n[%v] Crawling '%s' error: %v\n", event.When().Unix(), event.Page().Url, event.Page().Err)
			}
		case *spider.CrawlingEnded:
			fmt.Printf("\n[%v] Crawling '%s' ended! Visited %d unique URLs\n", event.When().Unix(), event.Id(), event.VisitedCount)
		default:
			fmt.Printf("\nUnhandled Event %T!\n", event)
		}
	}
}
