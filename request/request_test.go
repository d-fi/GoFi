package request

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestGetCacheKeyIncludesParams(t *testing.T) {
	previousClient := Client
	previousCache := cache
	t.Cleanup(func() {
		Client = previousClient
		cache = previousCache
	})

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/gateway.php", r.URL.Path)
		require.Equal(t, "test_method", r.URL.Query().Get("method"))
		requests++
		_, err := fmt.Fprintf(w, `{"error":[],"results":{"page":%q}}`, r.URL.Query().Get("page"))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	Client = resty.New().SetBaseURL(server.URL)
	cache = expirable.NewLRU[string, []byte](cacheSize, nil, cacheTTL)

	first, err := RequestGet("test_method", map[string]any{"page": 1})
	require.NoError(t, err)
	second, err := RequestGet("test_method", map[string]any{"page": 2})
	require.NoError(t, err)
	again, err := RequestGet("test_method", map[string]any{"page": 1})
	require.NoError(t, err)

	assert.JSONEq(t, `{"page":"1"}`, string(first))
	assert.JSONEq(t, `{"page":"2"}`, string(second))
	assert.Equal(t, string(first), string(again))
	assert.Equal(t, 2, requests)
}

func TestEncodeQueryParamsIsStable(t *testing.T) {
	params := map[string]string{"b": "2", "a": "1"}

	assert.Equal(t, "a=1&b=2", encodeQueryParams(params))
}
