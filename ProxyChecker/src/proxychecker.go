package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// Struct to store detailed proxy information
type Proxy struct {
	IP         string
	Port       string
	Source     string // Source where the proxy was scraped from
	Protocol   string // Proxy type: HTTP, HTTPS, SOCKS4, SOCKS5
	Anonymity  string // Proxy anonymity level: Transparent, Anonymous, Elite
	Alive      bool   // Proxy status: true if alive, false otherwise
	ResponseMs int    // Response time in milliseconds
	LastCheck  string // Timestamp for when the proxy was last validated
}

var userAgents []string
var proxySources map[string]string

// Load user agents from a text file into memory
func loadUserAgents(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open user agents file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		userAgents = append(userAgents, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading user agents file: %v", err)
	}
	return nil
}

// Load proxy sources from a JSON file into a map
func loadProxySources(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open proxy sources file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&proxySources); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}
	return nil
}

// Get a random User-Agent for request headers
func getRandomUserAgent() string {
	if len(userAgents) == 0 {
		log.Println("No user agents loaded. Using default user agent.")
		return "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	}
	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// Validate and classify proxy type and anonymity level
func checkAndClassifyProxy(proxy Proxy) Proxy {
	testURL := "https://httpbin.org/ip" // Test endpoint to verify proxy functionality
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%s", proxy.IP, proxy.Port),
			}),
		},
		Timeout: 5 * time.Second, // Fast timeout for quicker dead proxy detection
	}

	start := time.Now()
	req, _ := http.NewRequest("GET", testURL, nil)
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil || resp.StatusCode != 200 {
		proxy.Alive = false
		proxy.ResponseMs = int(elapsed)
		return proxy
	}

	defer resp.Body.Close()
	proxy.Alive = true
	proxy.ResponseMs = int(elapsed)
	proxy.Protocol = "HTTP/HTTPS"

	// Classify anonymity based on response and headers
	if resp.Request.Header.Get("Via") != "" || resp.Request.Header.Get("X-Forwarded-For") != "" {
		proxy.Anonymity = "Transparent"
	} else if resp.Request.Header.Get("Forwarded") != "" {
		proxy.Anonymity = "Anonymous"
	} else {
		proxy.Anonymity = "Elite"
	}

	return proxy
}

// Fetch proxies from a specific source
func fetchProxiesFromSource(url, sourceName string) ([]Proxy, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var proxies []Proxy
	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return proxies, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			t := tokenizer.Token()
			if t.Data == "tr" {
				proxy := extractProxyFromHTML(t, sourceName)
				if proxy != nil {
					proxies = append(proxies, *proxy)
				}
			}
		}
	}
}

// Helper function to extract proxy details from HTML
func extractProxyFromHTML(t html.Token, sourceName string) *Proxy {
	for _, attr := range t.Attr {
		if attr.Key == "data-proxy" {
			parts := strings.Split(attr.Val, ":")
			if len(parts) == 2 {
				return &Proxy{IP: parts[0], Port: parts[1], Source: sourceName}
			}
		}
	}
	return nil
}

// Main function to coordinate scraping, validation, and classification
func main() {
	// Load user agents and proxy sources
	if err := loadUserAgents("useragents.txt"); err != nil {
		log.Fatalf("Error loading user agents: %v", err)
	}

	if err := loadProxySources("ProxyList.json"); err != nil {
		log.Fatalf("Error loading proxy sources: %v", err)
	}

	var wg sync.WaitGroup
	proxyChannel := make(chan Proxy)
	checkedProxyChannel := make(chan Proxy)

	// Fetch proxies concurrently from all sources
	for sourceName, url := range proxySources {
		wg.Add(1)
		go func(sourceName, url string) {
			defer wg.Done()
			proxies, err := fetchProxiesFromSource(url, sourceName)
			if err == nil {
				for _, proxy := range proxies {
					proxyChannel <- proxy
				}
			} else {
				log.Printf("Error fetching from %s: %v", sourceName, err)
			}
		}(sourceName, url)
	}

	go func() {
		wg.Wait()
		close(proxyChannel)
	}()

	// Validate and classify proxies concurrently
	var validationWG sync.WaitGroup
	for proxy := range proxyChannel {
		validationWG.Add(1)
		go func(proxy Proxy) {
			defer validationWG.Done()
			checkedProxy := checkAndClassifyProxy(proxy)
			checkedProxyChannel <- checkedProxy
		}(proxy)
	}

	go func() {
		validationWG.Wait()
		close(checkedProxyChannel)
	}()

	// Write classified and validated proxies to file
	file, err := os.Create("validated_proxies.txt")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for proxy := range checkedProxyChannel {
		status := "Alive"
		if !proxy.Alive {
			status = "Dead"
		}
		writer.WriteString(fmt.Sprintf("%s:%s | Source: %s | Protocol: %s | Anonymity: %s | Status: %s | Response Time: %dms\n",
			proxy.IP, proxy.Port, proxy.Source, proxy.Protocol, proxy.Anonymity, status, proxy.ResponseMs))
	}
	writer.Flush()

	log.Println("Proxy checking and classification completed successfully.")
}
