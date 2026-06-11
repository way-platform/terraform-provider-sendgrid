package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"time"
)

type retryTransport struct {
	maxRetries int
	next       http.RoundTripper
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var br *bytes.Reader
	if req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, req.Body); err != nil {
			_ = req.Body.Close()
			return nil, fmt.Errorf("buffering body before retry: %w", err)
		}
		_ = req.Body.Close()
		br = bytes.NewReader(buf.Bytes())
		req.Body = io.NopCloser(br)
	}
	var attemptCount int
	for {
		res, err := t.next.RoundTrip(req)
		attemptCount++
		if attemptCount-1 >= t.maxRetries {
			return res, err
		}
		if !shouldRetry(err, req, res) {
			return res, err
		}
		delay := retryDelay(attemptCount, res)
		if br != nil {
			if _, serr := br.Seek(0, 0); serr != nil {
				return res, fmt.Errorf("seeking body buffer after attempt: %w", serr)
			}
			req.Body = io.NopCloser(br)
		}
		if res != nil {
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()
		}
		if err := sleepWithContext(req.Context(), delay); err != nil {
			return nil, err
		}
	}
}

func shouldRetry(err error, req *http.Request, res *http.Response) bool {
	if err != nil {
		var dnse *net.DNSError
		if errors.As(err, &dnse) {
			return true
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return isIdempotent(req)
		}
		return false
	}
	if res.Header.Get("Retry-After") != "" {
		return true
	}
	switch res.StatusCode {
	case http.StatusTooManyRequests:
		return true
	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return isIdempotent(req)
	default:
		return false
	}
}

func retryDelay(attempt int, res *http.Response) time.Duration {
	if res != nil {
		if ra := res.Header.Get("Retry-After"); ra != "" {
			if i, err := strconv.Atoi(ra); err == nil {
				return addJitter(time.Duration(i) * time.Second)
			}
			if t, err := time.Parse(http.TimeFormat, ra); err == nil {
				return addJitter(time.Until(t))
			}
		}
	}
	return expBackoff(attempt)
}

func expBackoff(attempt int) time.Duration {
	const base = 250 * time.Millisecond
	const cap = 10 * time.Second
	exp := math.Pow(2, float64(attempt-1))
	v := int64(math.Min(float64(cap), float64(base)*exp))
	if v <= 0 {
		return 0
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(v))
	return time.Duration(n.Int64())
}

func addJitter(d time.Duration) time.Duration {
	const magnitude = 0.333
	mj := int64(float64(d) * magnitude)
	if mj <= 0 {
		return d
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(mj))
	coin, _ := rand.Int(rand.Reader, big.NewInt(2))
	if coin.Int64() == 0 {
		return d + time.Duration(n.Int64())
	}
	return d - time.Duration(n.Int64())
}

func isIdempotent(req *http.Request) bool {
	switch req.Method {
	case http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPut,
		http.MethodDelete:
		return true
	}
	return false
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
