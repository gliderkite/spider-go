package spider

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

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

	// same do"main
	rawUrl := "https://monzo.com"
	if url := filter.filterMap(rootUrl, parentUrl, rawUrl); url == nil {
		t.Errorf("URL should have not been filtered")
	} else if url.String() != rawUrl {
		t.Errorf("invalid URL path")
	}

	rawUrl = "https://monzo.com/help/savings"
	if url := filter.filterMap(rootUrl, parentUrl, rawUrl); url == nil {
		t.Errorf("URL should have not been filtered")
	} else if url.String() != rawUrl {
		t.Errorf("invalid URL path")
	}

	// different domain
	if filter.filterMap(rootUrl, parentUrl, "https://google.com") != nil {
		t.Errorf("URL should have been filtered")
	}

	// invalid URL
	if filter.filterMap(rootUrl, parentUrl, "%%%") != nil {
		t.Errorf("URL should have been filtered")
	}
	if filter.filterMap(rootUrl, parentUrl, "") != nil {
		t.Errorf("URL should have been filtered")
	}
}
