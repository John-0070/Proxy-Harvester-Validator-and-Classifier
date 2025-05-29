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
