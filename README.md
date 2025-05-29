# Proxy Harvester, Validator, and Classifier (Go)

A simple multithreaded Go application for scraping, verifying, and classifying proxy servers from various public sources. Designed to help automate the discovery of alive proxies, identify anonymity levels, and write results to a structured output for future use in scraping, routing, or privacy applications.

## Features

- Concurrently scrapes proxies from multiple public sources (defined in a JSON config)
- Uses a rotating set of User-Agent headers for stealth
- Performs live connectivity tests via `httpbin.org` (Compatible with multiple sources)
- Classifies anonymity as:
  - **Transparent** (leaks client IP)
  - **Anonymous** (does not reveal IP but indicates use of proxy)
  - **Elite** (no detectable proxy headers)
- Measures response time (latency) for each proxy
- Outputs structured text log: IP:Port | Source | Protocol | Anonymity | Status | RTT

# Limitations
- Scraper assumes a specific format (<tr data-proxy="ip:port">). Other formats are ignored.

- Only HTTP/HTTPS proxy validation is supported (no SOCKS4/5 testing).

- Proxy protocol is assumed to be HTTP; no negotiation or SSL probing is done.

- Latency measurement is basic; no retries or jitter control.

- Anonymity classification is based solely on httpbin.org headers.

- Dead proxies may return false negatives due to timeout-based failure mode.

- The scraper is not resilient against anti-bot measures (CAPTCHAs, JavaScript rendering).

- No rate limiting or throttle control for scraping endpoints.

- Proxy sources are static (from JSON) and not dynamically discovered.

- Logging is console-based; no JSON or structured logging for ingestion pipelines.

# To Be Added
- SOCKS4/5 protocol validation support

- Retry logic and exponential backoff for failed validations

- CLI options for concurrency control, timeouts, and output verbosity

- Structured output formats: CSV, JSON, SQLite

- Support for additional HTML formats (table scraping, DOM parsing)

- Proxy deduplication across sources

- Geolocation and ASN enrichment of proxy data

- UI/TUI mode for interactive exploration
  
- Docker containerization for deployment
