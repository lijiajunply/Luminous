package util

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultTimeout = 30 * time.Second
const MaxRetries = 3
const maxRedirects = 3

var privateCIDRs = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"::1/128",
	"fe80::/10",
}

var privateNets []*net.IPNet

func init() {
	for _, cidr := range privateCIDRs {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic("invalid built-in CIDR: " + cidr)
		}
		privateNets = append(privateNets, block)
	}
}

func isPrivateIP(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, block := range privateNets {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:132.0) Gecko/20100101 Firefox/132.0",
}

func RandomUserAgent() string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents))))
	if err != nil {
		return userAgents[0]
	}
	return userAgents[n.Int64()]
}

type HTTPClient struct {
	client *http.Client
}

func safeCheckRedirect(allowedHosts map[string]bool) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		host, _, err := net.SplitHostPort(req.URL.Host)
		if err != nil {
			host = req.URL.Host
		}
		if isPrivateIP(host) {
			return fmt.Errorf("redirect to private IP forbidden: %s", host)
		}
		if len(allowedHosts) > 0 && !allowedHosts[host] {
			return fmt.Errorf("redirect to unallowed host: %s", host)
		}
		return nil
	}
}

func NewHTTPClient() *HTTPClient {
	return NewHTTPClientWithAllowedHosts(nil)
}

func NewHTTPClientWithAllowedHosts(allowedHosts map[string]bool) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout:        DefaultTimeout,
			CheckRedirect:  safeCheckRedirect(allowedHosts),
		},
	}
}

func isRetriable(err error) bool {
	if err == nil {
		return false
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}

var errStatus = fmt.Errorf("upstream returned status")

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i < MaxRetries; i++ {
		if i > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			req.Body = body
		}
		resp, err = c.client.Do(req)
		if err == nil {
			if resp.StatusCode >= 500 {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if i < MaxRetries-1 {
					time.Sleep(time.Duration(1<<uint(i)) * time.Second)
				}
				continue
			}
			return resp, nil
		}
		if !isRetriable(err) {
			return nil, err
		}
		if i < MaxRetries-1 {
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
		}
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("%w: %d after %d retries", errStatus, resp.StatusCode, MaxRetries)
}

func (c *HTTPClient) setDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", RandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
}

func (c *HTTPClient) Get(url string) (*http.Response, error) {
	return c.GetWithContext(context.Background(), url)
}

func (c *HTTPClient) GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setDefaultHeaders(req)
	return c.Do(req)
}

var DefaultClient = NewHTTPClient()

func (c *HTTPClient) FetchWithCookie(url, cookie string) ([]byte, error) {
	resp, err := c.GetWithCookie(url, cookie)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("nil response from server")
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *HTTPClient) GetWithCookie(url string, cookie string) (*http.Response, error) {
	return c.GetWithCookieContext(context.Background(), url, cookie)
}

func (c *HTTPClient) GetWithCookieContext(ctx context.Context, url string, cookie string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setDefaultHeaders(req)
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	return c.Do(req)
}

func (c *HTTPClient) GetWithHeaders(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setDefaultHeaders(req)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.Do(req)
}

func (c *HTTPClient) PostForm(url string, data url.Values, headers map[string]string) (*http.Response, error) {
	body := data.Encode()
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	}
	req.Header.Set("User-Agent", RandomUserAgent())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.Do(req)
}
