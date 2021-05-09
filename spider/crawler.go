// Defines the HTML crawler implementation for a single domain.
package spider

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Map of URLs found.
type urlsMap = map[string]*url.URL

// Set of URLs found.
type urlsSet = map[string]bool

// The outcome of a single URL crawling.
type pageUrls struct {
	Url  *url.URL
	Urls *urlsMap
	Err  error
}

/// The HTML crawler for a single domain.
type crawler struct {
	// The HTTP client.
	client *http.Client
	// The root URL.
	rootUrl *url.URL
	// The HTML tokens parser.
	tokenParser htmlParser
	// The URLs filter.
	urlFilter urlFilter
}

// Constructs a new HTML crawler for the same domain.
func newCrawler(rootUrl *url.URL, timeout time.Duration) *crawler {
	return &crawler{&http.Client{Timeout: timeout}, rootUrl, &hrefParser{}, &sameDomainFilter{}}
}

// Crawls a single URL and send all the URLs found via the given channel.
func (crawler *crawler) crawl(seedUrl *url.URL, pageCh chan pageUrls) {
	resp, err := crawler.client.Get(seedUrl.String())
	if err != nil {
		pageCh <- pageUrls{seedUrl, nil, err}
		return
	}
	if resp.StatusCode != http.StatusOK {
		pageCh <- pageUrls{seedUrl, nil, fmt.Errorf("status code: %d", resp.StatusCode)}
		return
	}

	respBody := resp.Body
	defer respBody.Close()

	urls := make(urlsMap)
	// parse the response body until EOF
	tokenizer := html.NewTokenizer(respBody)
	for {
		switch tokenType := tokenizer.Next(); tokenType {
		case html.ErrorToken: // EOF
			pageCh <- pageUrls{seedUrl, &urls, nil}
			return
		case html.StartTagToken:
			token := tokenizer.Token()
			// extract the link from the token if it's a href attribute
			if u := crawler.tokenParser.filterMap(&token); u != nil {
				if u := crawler.urlFilter.filterMap(crawler.rootUrl, seedUrl, *u); u != nil {
					key, _ := url.PathUnescape(u.String())
					urls[key] = u
				}
			}
		}
	}
}

// HTML crawler tokens parser.
type htmlParser interface {
	// Parse the given HTML token and returns nil only if the given token must be
	// discarded. Otherwise returns the string representation of the parsed token.
	filterMap(token *html.Token) *string
}

// Parse and filter all tokens that are href attributes of an anchor element.
type hrefParser struct{}

func (parser *hrefParser) filterMap(token *html.Token) *string {
	if token.Data != "a" {
		// not an anchor element
		return nil
	}
	for _, attribute := range token.Attr {
		if attribute.Key == "href" {
			return &attribute.Val
		}
	}
	// no href attribute found
	return nil
}

// crawler URL filter.
type urlFilter interface {
	// Parse the given URL and returns nil only if the give URL must be discarded.
	// Otherwise returns the new URL.
	filterMap(rootUrl, parentUrl *url.URL, url string) *url.URL
}

// Parse and filter all URLs that do not belong to the same domain.
type sameDomainFilter struct{}

func (filter *sameDomainFilter) filterMap(rootUrl, parentUrl *url.URL, rawUrl string) *url.URL {
	if len(rawUrl) == 0 {
		return nil
	}

	// sanitize the raw URL
	rawUrl = strings.TrimSpace(strings.TrimSuffix(rawUrl, "/"))
	rawUrl, err := url.PathUnescape(rawUrl)
	if err != nil {
		//fmt.Println("URL Unescape error:", rawUrl, err)
		return nil
	}

	var url *url.URL
	if strings.HasPrefix(rawUrl, "//") {
		// URL relative to the document root scheme
		url, err = url.Parse(fmt.Sprintf("%s:%s", rootUrl.Scheme, rawUrl))
	} else if strings.HasPrefix(rawUrl, "/") {
		// URL relative to the document root
		url, err = url.Parse(fmt.Sprintf("%s%s", rootUrl, rawUrl))
	} else if !strings.Contains(rawUrl, "://") {
		// URL relative to the current path
		url, err = url.Parse(fmt.Sprintf("%s%s", parentUrl, rawUrl))
	} else {
		// absolute URL
		url, err = url.Parse(rawUrl)
	}

	if err != nil {
		//fmt.Println("Invalid URL:", rawUrl, err)
		return nil
	}

	// remove query and fragment from the URL
	url.Fragment = ""
	url.RawQuery = ""

	if url.Hostname() != rootUrl.Hostname() {
		// exclude URLs with a different domain
		return nil
	}

	return url
}
