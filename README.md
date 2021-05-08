# spider-go


## TODO
- [ ] Proper visibility to types and functions
- [ ] Optimize?
- [ ] Unit tests!
- [ ] Integration tests!
- [ ] Remove unused fmt.Print
- [ ] Documentation
    - [ ] What has been prioritized?
    - [ ] Tradeoff
    - [ ] Go vs Rust
    - [ ] What would have done differently with more time/different requirements


## Possible improvements
- Optimization (check URLs already in frontiers memory vs CPU) -> needs profiling
- Persistency (DB)
- Logger
- Error handling could be improved (requires signal back, retries?)
- URL parsing more adherent to specs
- Integration tests: can we guarantee a stable internet connection? Use custom website to crawl with known sitemap.