package hfclient

import (
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/Ericson246/npu-optimize/internal/cache"
	"github.com/Ericson246/npu-optimize/internal/constants"
)

const userAgent = "npu-optimize/0.1.0"

type AuthError struct {
	msg string
}

func (e *AuthError) Error() string { return e.msg }

type requestFunc func() (*http.Response, error)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	token      string
	Cache      *cache.Cache
}

func NewClient(token string) *Client {
	return &Client{
		BaseURL: constants.HFAPIBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

func (c *Client) SetCache(cache *cache.Cache) {
	c.Cache = cache
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func (c *Client) doRequest(url string) ([]byte, error) {
	return c.doWithRetry(func() (*http.Response, error) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		c.setHeaders(req)
		return c.HTTPClient.Do(req)
	})
}

func (c *Client) doWithRetry(fn requestFunc) ([]byte, error) {
	var lastErr error
	var retryAfter time.Duration

	for attempt := range maxRetries {
		if attempt > 0 {
			delay := backoffDuration(attempt, retryAfter)
			slog.Debug("retrying request",
				"attempt", attempt+1,
				"max", maxRetries,
				"delay_ms", delay.Milliseconds(),
			)
			time.Sleep(delay)
		}

		resp, err := fn()
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading response: %w", readErr)
		}

		limitInfo := parseRateLimit(resp.Header)

		switch {
		case resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent:
			if limitInfo.Remaining >= 0 && limitInfo.Remaining < 20 {
				slog.Warn("rate limit low",
					"remaining", limitInfo.Remaining,
				)
			}
			return body, nil

		case resp.StatusCode == http.StatusTooManyRequests:
			retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
			lastErr = &RateLimitError{
				msg:        "rate limited by HuggingFace API",
				RetryAfter: retryAfter,
				Limit:      limitInfo,
			}
			continue

		case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
			return nil, &AuthError{
				msg: "HuggingFace authentication required. Use --token or set HF_TOKEN env var",
			}

		case resp.StatusCode >= 500:
			lastErr = fmt.Errorf("server error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
			continue

		default:
			return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
	}

	return nil, lastErr
}

func parseRateLimit(h http.Header) RateLimitInfo {
	var info RateLimitInfo
	if rem := h.Get("X-RateLimit-Remaining"); rem != "" {
		if n, err := strconv.Atoi(rem); err == nil {
			info.Remaining = n
		} else {
			info.Remaining = -1
		}
	} else {
		info.Remaining = -1
	}
	if reset := h.Get("X-RateLimit-Reset"); reset != "" {
		if n, err := strconv.ParseInt(reset, 10, 64); err == nil {
			info.ResetAt = time.Unix(n, 0)
		}
	}
	return info
}

func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(val); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if t, err := time.Parse(http.TimeFormat, val); err == nil {
		return time.Until(t)
	}
	return 0
}

func backoffDuration(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}
	base := float64(time.Second)
	maxDelay := float64(30 * time.Second)
	exp := math.Pow(2, float64(attempt)) * base
	capDelay := math.Min(exp, maxDelay)
	jitter := float64(cryptoRandInt64N(int64(capDelay)))
	return time.Duration(jitter)
}

func cryptoRandInt64N(n int64) int64 {
	if n <= 0 {
		return 0
	}
	val, err := rand.Int(rand.Reader, big.NewInt(n))
	if err != nil {
		return n / 2
	}
	return val.Int64()
}

func (c *Client) cacheKey(kind, url string) string {
	return cache.Fingerprint(kind + "|" + url)
}

func (c *Client) getFromCache(key string) ([]byte, bool) {
	if c.Cache == nil {
		return nil, false
	}
	return c.Cache.Get(key)
}

func (c *Client) storeInCache(key string, data []byte, ttl time.Duration) {
	if c.Cache == nil {
		return
	}
	if err := c.Cache.Set(key, data, ttl); err != nil {
		slog.Warn("failed to write cache", "key", key, "err", err)
	}
}
