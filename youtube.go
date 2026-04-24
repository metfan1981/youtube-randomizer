package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

type cacheEntry struct {
	ids       []string
	expiresAt time.Time
}

// Client fetches a channel's upload video IDs and caches them per handle.
type Client struct {
	svc   *youtube.Service
	ttl   time.Duration
	mu    sync.RWMutex
	cache map[string]cacheEntry
}

func newClient(apiKey string, ttl time.Duration) (*Client, error) {
	svc, err := youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &Client{
		svc:   svc,
		ttl:   ttl,
		cache: make(map[string]cacheEntry),
	}, nil
}

func cacheKey(handle string) string {
	return strings.ToLower(strings.TrimPrefix(handle, "@"))
}

// Uploads returns all video IDs from the channel's uploads playlist for handle (with or without leading @).
func (c *Client) Uploads(ctx context.Context, handle string) ([]string, error) {
	key := cacheKey(handle)

	c.mu.RLock()
	e, ok := c.cache[key]
	if ok && time.Now().Before(e.expiresAt) {
		c.mu.RUnlock()
		out := make([]string, len(e.ids))
		copy(out, e.ids)
		return out, nil
	}
	c.mu.RUnlock()

	ids, err := c.fetchAll(ctx, handle)
	if err != nil {
		return nil, err
	}

	// Do not cache empty lists: unknown handles, deleted playlists, or transient
	// API gaps would stay stale for the full TTL (bad UX for typos and new channels).
	if len(ids) > 0 {
		c.mu.Lock()
		c.cache[key] = cacheEntry{ids: ids, expiresAt: time.Now().Add(c.ttl)}
		c.mu.Unlock()
	}

	out := make([]string, len(ids))
	copy(out, ids)
	return out, nil
}

func (c *Client) fetchAll(ctx context.Context, handle string) ([]string, error) {
	forHandle := "@" + strings.TrimPrefix(handle, "@")

	chResp, err := c.svc.Channels.List([]string{"contentDetails"}).ForHandle(forHandle).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	if len(chResp.Items) == 0 {
		return nil, nil
	}

	item := chResp.Items[0]
	if item == nil || item.ContentDetails == nil || item.ContentDetails.RelatedPlaylists == nil {
		return nil, nil
	}

	uploadsPlaylistID := item.ContentDetails.RelatedPlaylists.Uploads
	if uploadsPlaylistID == "" {
		return nil, nil
	}

	var ids []string
	pageToken := ""
	for {
		call := c.svc.PlaylistItems.List([]string{"contentDetails"}).
			PlaylistId(uploadsPlaylistID).
			MaxResults(50).
			Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		plResp, err := call.Do()
		if err != nil {
			return nil, err
		}
		for _, item := range plResp.Items {
			if item == nil || item.ContentDetails == nil {
				continue
			}
			vid := item.ContentDetails.VideoId
			if vid == "" {
				continue
			}
			ids = append(ids, vid)
		}
		pageToken = plResp.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return ids, nil
}
