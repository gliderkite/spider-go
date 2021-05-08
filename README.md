# spider-go


## TODO
- [ ] Replace `fmt` with log levels
- [ ] Command line argument parser
- [ ] Limit URLs to visit per step
- [ ] Set HTTP client timeout
- [ ] Proper visibility to types and functions
- [ ] Optimize?
- [ ] Unit tests!
- [ ] Integration tests!
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
