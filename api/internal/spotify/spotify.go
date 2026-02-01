// Package spotify contains functions for maintaining spotify connections and interacting with spotify
package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	marshttp "mars/internal/http"

	"github.com/hashicorp/go-retryablehttp"
)

// RefreshToken sends a request to the refresh spotify token oauth endpoint.
func RefreshToken(ctx context.Context, client marshttp.Client, userid, accessToken, csrfToken string) error {
	const endpoint = "http://localhost:8080/api/oauth/spotify/token/refresh"
	body, err := json.Marshal(map[string]string{
		"user_id": userid,
	})
	if err != nil {
		return fmt.Errorf("marshaling body: %w", err)
	}
	req, err := retryablehttp.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Cookie", fmt.Sprintf("access=%s; csrf=%s", accessToken, csrfToken))
	req.Header.Add("X-CSRF-Token", csrfToken)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("request failed with non-200 status: status=%d body=%s", res.StatusCode, string(body))
	}
	return nil
}
