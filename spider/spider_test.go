package spider

import (
	"net/url"
	"testing"
	"time"
)

// NOTE: This integration test requires a connection to the internet.
func TestSpider(t *testing.T) {
	eventCh := make(chan Event)
	conf := SpiderConf{16, time.Duration(1) * time.Second}
	spider := NewSpider(&conf, eventCh)

	// should crawl a reachable URL and return the expected events
	url, _ := url.Parse("https://gliderkite.github.io")
	domain := url.Hostname()
	go spider.Crawl(url.String())

	event := <-eventCh
	switch event := event.(type) {
	case *CrawlingStarted:
		if event.Id() != domain {
			t.Errorf("Unexpected event ID '%s'", event.Id())
		}
	case *PageVisited:
		if event.Id() != domain {
			t.Errorf("Unexpected event ID '%s'", event.Id())
		}
		if event.Page.Err == nil {
			// TODO: if this is a known sitemap, assert that the page visited
			// belongs to the set of expected URL, and that the page has not
			// been visited before.
		} else {
			t.Errorf("Error crawling '%s'", event.Page.Url)
		}
	case *CrawlingEnded:
		if event.Id() != domain {
			t.Errorf("Unexpected event ID '%s'", event.Id())
		}
	default:
		t.Errorf("Unhandled Event %T!\n", event)
	}
}
