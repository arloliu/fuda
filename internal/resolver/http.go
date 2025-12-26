package resolver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HTTPResolver resolves references using the http:// and https:// schemes.
type HTTPResolver struct {
	Client  *http.Client
	MaxSize int64 // Max size in bytes to read (default: 16MB)
}

// NewHTTPResolver creates a new HTTPResolver.
func NewHTTPResolver() *HTTPResolver {
	return &HTTPResolver{
		Client:  http.DefaultClient,
		MaxSize: 16 * 1024 * 1024, // 16MB default
	}
}

// Resolve fetches content from the given URI using an HTTP GET request.
func (r *HTTPResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI %q: %w", uri, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme for http resolver: %s", u.Scheme)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request failed with status: %d", resp.StatusCode)
	}

	limit := r.MaxSize
	if limit == 0 {
		limit = 16 * 1024 * 1024 // Fallback default
	}

	// Read with limit + 1 to detect overflow
	reader := io.LimitReader(resp.Body, limit+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > limit {
		return nil, fmt.Errorf("reference content exceeds maximum size of %d bytes", limit)
	}

	return data, nil
}
