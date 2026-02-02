// Package mars contains functions for interacting with the mars api
package mars

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	marshttp "mars/internal/http"
	"mars/internal/spotify"

	"github.com/hashicorp/go-retryablehttp"
)

func Login(ctx context.Context, client marshttp.Client, email, password string) (
	accessToken string, refreshToken string, csrfToken string, err error,
) {
	const endpoint = "http://localhost:8080/api/login"
	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return "", "", "", fmt.Errorf("marshaling body: %w", err)
	}
	req, err := retryablehttp.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return "", "", "", fmt.Errorf("request failed with non-200 status: status=%d body=%s", res.StatusCode, string(body))
	}

	for _, cookie := range res.Cookies() {
		switch cookie.Name {
		case "access":
			accessToken = cookie.Value
		case "refresh":
			refreshToken = cookie.Value
		case "csrf":
			csrfToken = cookie.Value
		}
	}
	return accessToken, refreshToken, csrfToken, nil
}

func ListUsers(ctx context.Context, client marshttp.Client, accessToken string) ([]string, error) {
	const endpoint = "http://localhost:8080/api/users"
	req, err := retryablehttp.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Cookie", "access="+accessToken)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("request failed with non-200 status: status=%d body=%s", res.StatusCode, string(body))
	}

	var body struct {
		Ids []string `json:"ids"`
	}
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return body.Ids, nil
}

func RefreshSpotifyTokens(ctx context.Context, client marshttp.Client, email, password string) error {
	accessToken, _, csrfToken, err := Login(ctx, client, email, password)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	userids, err := ListUsers(ctx, client, accessToken)
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}

	var wg sync.WaitGroup
	var mtx sync.Mutex
	var errs []error
	for _, id := range userids {
		// would probably cause hella rate-limiting errors, but i'm the only user so who cares
		wg.Go(func() {
			err := spotify.RefreshToken(ctx, client, id, accessToken, csrfToken)
			if err != nil {
				mtx.Lock()
				errs = append(errs, fmt.Errorf("refreshing tokens for user (%s): %w", id, err))
				mtx.Unlock()
			}
		})
	}
	wg.Wait()

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func SyncSpotifyTracks(ctx context.Context, client marshttp.Client, email, password string) error {
	accessToken, _, csrfToken, err := Login(ctx, client, email, password)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	userids, err := ListUsers(ctx, client, accessToken)
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}

	var wg sync.WaitGroup
	var mtx sync.Mutex
	var errs []error
	for _, id := range userids {
		wg.Go(func() {
			err := spotify.SyncTracks(ctx, client, id, accessToken, csrfToken)
			if err != nil {
				mtx.Lock()
				errs = append(errs, fmt.Errorf("syncing tracks for user (%s): %w", id, err))
				mtx.Unlock()
			}
		})
	}
	wg.Wait()

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func CreatePlaylist(ctx context.Context, client marshttp.Client,
	email, password, playlisttype string,
	year int, month time.Month, day int,
) error {
	accessToken, _, csrfToken, err := Login(ctx, client, email, password)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	userids, err := ListUsers(ctx, client, accessToken)
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}

	var wg sync.WaitGroup
	var mtx sync.Mutex
	var errs []error
	for _, id := range userids {
		wg.Go(func() {
			// Create request
			const endpoint = "http://localhost:8080/api/playlists"
			body, err := json.Marshal(map[string]any{
				"user_id": id,
				"start_date": map[string]int{
					"year":  year,
					"month": int(month),
					"day":   day,
				},
				"type": playlisttype,
			})
			if err != nil {
				mtx.Lock()
				errs = append(errs, fmt.Errorf("marshaling body for user (%s): %w", id, err))
				mtx.Unlock()
				return
			}
			req, err := retryablehttp.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
			if err != nil {
				mtx.Lock()
				errs = append(errs, fmt.Errorf("creating request for user (%s): %w", id, err))
				mtx.Unlock()
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", csrfToken)
			req.Header.Set("Cookie", fmt.Sprintf("access=%s; csrf=%s", accessToken, csrfToken))

			// Send request
			res, err := client.Do(req)
			if err != nil {
				mtx.Lock()
				errs = append(errs, fmt.Errorf("sending request for user (%s): %w", id, err))
				mtx.Unlock()
				return
			}
			defer func() { _ = res.Body.Close() }()
			// No tracks listened to, this is fine.
			if res.StatusCode == http.StatusConflict {
				return
			}
			if res.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(res.Body)
				mtx.Lock()
				errs = append(errs,
					fmt.Errorf("request failed with non-201 status: status=%d body=\"%s\"", res.StatusCode, string(body)))
				mtx.Unlock()
				return
			}

			// Parse playlist ID from response
			var respBody struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
				mtx.Lock()
				errs = append(errs, fmt.Errorf("decoding response for user (%s): %w", id, err))
				mtx.Unlock()
				return
			}

			// Create playlist on Spotify
			if err := spotify.CreatePlaylist(ctx, client, id, respBody.ID, accessToken, csrfToken); err != nil {
				mtx.Lock()
				errs = append(errs, fmt.Errorf("creating spotify playlist for user (%s): %w", id, err))
				mtx.Unlock()
				return
			}
		})
	}
	wg.Wait()

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}
