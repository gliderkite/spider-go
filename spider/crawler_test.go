package spider

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestHrefParserShouldFindLinks(t *testing.T) {
	const (
		href  = "https://www.monzo.com"
		href2 = "https://www.w3schools.com"
	)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
		<html>
		<body>
		<h2>HTML Links</h2>
		<p>HTML links are defined with the a tag:</p>
		
		<a href="%s">This is a link</a>
		<a href="%s">This is a another link</a>
		
		</body>
		</html>`, href, href2)

	links := extractAttributes(htmlBody, &hrefParser{})
	if _, ok := links[href]; !ok {
		t.Errorf("href '%s' not found", href)
	}

	if _, ok := links[href2]; !ok {
		t.Errorf("href '%s' not found", href2)
	}
}

func TestHrefParserShouldNotFindLink(t *testing.T) {
	htmlBody := `<!DOCTYPE html>
		<html>
		<body>
				
		</body>
		</html>`

	links := extractAttributes(htmlBody, &hrefParser{})
	if len(links) > 0 {
		t.Errorf("href was not specified")
	}
}

/// Extracts all the attributes according to the parser implementation.
func extractAttributes(body string, parser htmlParser) map[string]bool {
	links := make(map[string]bool)

	tokenizer := html.NewTokenizer(strings.NewReader(body))
	for {
		switch tokenType := tokenizer.Next(); tokenType {
		case html.ErrorToken: // EOF
			return links
		case html.StartTagToken:
			token := tokenizer.Token()
			if href := parser.filterMap(&token); href != nil {
				links[*href] = true
			}
		}
	}
}

func TestSameDomainFilter(t *testing.T) {
	rootUrl, _ := url.Parse("https://monzo.com")
	parentUrl, _ := url.Parse("https://monzo.com/help")

	filter := &sameDomainFilter{}

	// same domain
	rawUrl := "https://monzo.com"
	if url := filter.filterMap(rootUrl, parentUrl, rawUrl); url == nil {
		t.Errorf("'%s' should have not been filtered", rawUrl)
	} else if url.String() != rawUrl {
		t.Errorf("invalid URL path: %s", url)
	}

	rawUrl = "https://monzo.com/help/savings"
	if url := filter.filterMap(rootUrl, parentUrl, rawUrl); url == nil {
		t.Errorf("'%s' should have not been filtered", rawUrl)
	} else if url.String() != rawUrl {
		t.Errorf("invalid URL path: %s", url)
	}

	// different domain
	rawUrl = "https://google.com"
	if filter.filterMap(rootUrl, parentUrl, rawUrl) != nil {
		t.Errorf("'%s' should have been filtered", rawUrl)
	}

	// subdomain
	rawUrl = "https://community.monzo.com"
	if filter.filterMap(rootUrl, parentUrl, rawUrl) != nil {
		t.Errorf("'%s' should have been filtered", rawUrl)
	}

	// invalid URL
	rawUrl = "%"
	if filter.filterMap(rootUrl, parentUrl, rawUrl) != nil {
		t.Errorf("'%s' should have been filtered", rawUrl)
	}

	rawUrl = ""
	if filter.filterMap(rootUrl, parentUrl, rawUrl) != nil {
		t.Errorf("'%s' should have been filtered", rawUrl)
	}
}

// NOTE: This integration test requires a connection to the internet.
func TestCrawler(t *testing.T) {
	const timeout = time.Duration(1) * time.Second
	rootUrl, _ := url.Parse("https://monzo.com")
	pageCh := make(chan pageUrls)

	crawler := newCrawler(rootUrl, timeout)

	// should crawl a reachable URL
	go crawler.crawl(rootUrl, pageCh)
	if pageRes := <-pageCh; pageRes.Err != nil {
		t.Errorf("Error crawling '%s': %v", rootUrl, pageRes.Err)
	}

	// should crawl a reachable URL page
	url, _ := url.Parse("https://monzo.com/help")
	go crawler.crawl(url, pageCh)
	if pageRes := <-pageCh; pageRes.Err != nil {
		t.Errorf("Error crawling '%s': %v", url, pageRes.Err)
	}

	// should not crawl a URL that returns 404
	url, _ = url.Parse("https://monzo.com/not-help")
	go crawler.crawl(url, pageCh)
	if pageRes := <-pageCh; pageRes.Err == nil {
		t.Errorf("Should not crawl '%s'", url)
	}

	// should not crawl a not existing domain
	url, _ = url.Parse("https://monzo.pt")
	go crawler.crawl(url, pageCh)
	if pageRes := <-pageCh; pageRes.Err == nil {
		t.Errorf("Should not crawl '%s'", url)
	}
}

func TestCrawlerSiteMap(t *testing.T) {
	// TODO: This integration test requires knowledge of the precise sitemap
	// of the URL to crawl, so that it would be possible to compare the set
	// of URLs crawled to those that were expected.
}
