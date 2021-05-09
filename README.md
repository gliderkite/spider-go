# spider-go

This is a basic web crawler implementation for HTML pages belonging to the same
domain, written in `Go`.

Given a starting URL, the crawler will visit each URL it finds on the same domain.
It will print each URL visited and the list of links found on that page (belonging
to the same domain).


## Project Structure

The project structure is relatively simple. You will find the following components:

```
spider-go
│   README.md
│   go.mod
|   go.sum
|   main.go -- The entry point of the application.
│
└───spider -- The package that defines all the crawler components.
│   │   spider.go   -- The main component of the crawler, exported to the user,
|   |                   with the logic used to crawl the whole sitemap.
│   │   crawler.go  -- The component that defines the logic for a single URL
|   |                   crawling (HTML parsing, filtering logic).
|   |   event.go    -- Defines all the asynchronous events that can be generated
|   |                   by the crawler (Spider).
|   |   *_test.go   -- Set of unit and integration tests for the various components.
```


## How to Use

### Requirements

In order to run this application you need to install
[the Go programming language](https://golang.org), and get the dependency
`golang.org/x/net/html` with:

```console
go get golang.org/x/net/html
```

### Running the crawler

The entry point of this application allows you to run the crawler for a single
domain, while also specifying via command line arguments some configuration
values. In particular, these are:

- `--url <URL to crawl>` (default: `https://monzo.com`)
- `--timeout <HTTP request timeout (s)>` (default: `5s`)
- `--concurrency <Max number of URLs to visit concurrently>` (default: `64`)

You can run the crawler with:

```console
go run main.go --url https://gliderkite.github.io
```

Note: once the crawling is over, you will see a message stating so; to terminate
the application you will need to abort the even loop and terminate the application
with `Ctrl+C` (more on this choice in the design section).


## Design

*Disclaimer:*
This is the second time ever I use `Go`, please apologize if the code and/or structure
is not as idiomatic as it should be.

While developing this project, the aspects that were prioritized the most were the following:
- Use of abstraction where deemed useful, to break components in smaller parts.
- Concurrency, with an highly based asynchronous message passing design.
- Ease of testability, with dependency injection to allow easier unit testing.
- Configuration, via few but significant options that allow to control the behavior
    of the crawler.

The main application behavior is structured around the crawler implementation,
which uses asynchronous goroutines to perform the crawling (according to the
level of concurrency desired) and communicates back the outcome of each *event*
so that they can be gathered in the same (possibly never ending) event loop.
This effectively represents a Multiple Producer Single Consumer (MPSC) communication
over channels.

This enables a high freedom of usage, and allows the user to implement a higher
number of features on top of the current implementation, according to the event
kind received.

To make an example:

```go
// create the Spider instance
eventCh := make(chan spider.Event)
spider := spider.NewSpider(conf, eventCh)

// the same instance can be used to crawl multiple domains asynchronously
go spider.Crawl("https://monzo.com")
go spider.Crawl("https://google.com")

// the event loop can be used to filter all the events identified by an unique ID
for event := range eventCh {
    switch event := event.(type) {
    case *spider.CrawlingStarted:
        fmt.Printf("Crawling '%s' started\n", event.Id())
    case *spider.PageVisited:
        // ...
    case *spider.CrawlingEnded:
        fmt.Printf("Crawling '%s' ended!\n", event.Id())
    }
}
```

Regarding the actual crawling, each crawler maintains a set of unique URLs
already visited and a FIFO frontier of URLs yet to visit. The exploration is 
based on a breadth-first search algorithm variant. This is because [if the
crawler wants to download pages with high Pagerank early during the
crawling process, then the partial Pagerank strategy is the better one, followed by
breadth-first and backlink-count][1] and also because [breadth-first crawl captures
pages with high Pagerank early in the crawl][2].

For each HTML page, all the URLs are scraped, sanitized, and filtered according
to the implementation of each of these components, which can be injected into the
crawler instance when being constructed and allow to unit test each single
component. The crawler internally runs a set of goroutines allowing each URL
found to be parsed concurrently.


## Follow up as possible improvements

There are indeed several aspects of this implementation that could be improved,
as well as many more features that could be enabled according to different
requirements and tradeoffs that have been made.

- Testing
    - Unit testing could be improved to cover more edge cases. This could
        be done easily thanks to the current design.
    - Integration testing could be significantly enhanced. This would require
        a stricter control over the resources and the environment where
        the tests run (for example, by defining a known sitemap and setting up a
        local DNS).
- Optimization
    - The current implementation tries to favor CPU over memory usage (especially
        when determining if a URL needs to be marked as to visit). However, no
        profiling has been done and, before reaching any conclusion, it should be
        highly considered.
    - It would be possible to not store the *root* URL in each URL because, 
        being a single domain crawler, it is expected to be the same for all of
        them.
- Logging
    - A production-ready solution would exploit logging by using different log
        levels and different sinks (stdout/stderr, files, centralized logging
        system running as part of a different service).
- Error handling
    - The current error handling is minimal. While the message passing allows the
        consumer to know exactly if any of the pages crawled returned an error,
        this could be enhanced by implementing, for example, a retry
        mechanism, but it would require a way of communicating back or a
        shared state via a persistent storage.
- Persistency
    - All the URLs are stored in memory. The current design could allow to expand
        the set of events to be able to report which URLs have been visited
        already and which are yet to be visited. The consumer could then use
        another new component to store these results in a database for example,
        which can then be used by the crawler to synchronize its state.
- Adherence to specs, including:
    - URL parsing, sanitization.
    - [robots.txt](https://en.wikipedia.org/wiki/Robots_exclusion_standard)
- Configuration
    - The crawler configuration could be enhanced by introducing new parameters,
        such as the maximum *depth* of exploration for example, depending on
        requirements.


[1]: https://en.wikipedia.org/wiki/Web_crawler#cite_note-11
[2]: https://en.wikipedia.org/wiki/Web_crawler#cite_note-12
