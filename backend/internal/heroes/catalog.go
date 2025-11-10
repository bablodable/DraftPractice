package heroes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://api.opendota.com"

// Catalog fetches hero data from the OpenDota public API and caches it for reuse.
type Catalog struct {
	client  *http.Client
	baseURL string
	ttl     time.Duration

	mu        sync.RWMutex
	cache     []Hero
	fetchedAt time.Time
}

// NewCatalog constructs a Catalog using the provided http.Client, base URL and cache TTL.
// If client is nil a default client with a reasonable timeout will be used.
func NewCatalog(client *http.Client, baseURL string, ttl time.Duration) *Catalog {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}

	return &Catalog{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		ttl:     ttl,
	}
}

// List returns the cached hero list or fetches it from the remote API when needed.
func (c *Catalog) List(ctx context.Context) ([]Hero, error) {
	c.mu.RLock()
	if len(c.cache) > 0 && time.Since(c.fetchedAt) < c.ttl {
		heroes := cloneHeroes(c.cache)
		c.mu.RUnlock()
		return heroes, nil
	}
	c.mu.RUnlock()

	return c.refresh(ctx)
}

func (c *Catalog) refresh(ctx context.Context) ([]Hero, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) > 0 && time.Since(c.fetchedAt) < c.ttl {
		return cloneHeroes(c.cache), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/heroes", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if len(c.cache) > 0 {
			return cloneHeroes(c.cache), nil
		}
		return nil, fmt.Errorf("fetch heroes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if len(c.cache) > 0 {
			return cloneHeroes(c.cache), nil
		}
		return nil, fmt.Errorf("unexpected status %d from hero API", resp.StatusCode)
	}

	var payload []openDotaHero
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		if len(c.cache) > 0 {
			return cloneHeroes(c.cache), nil
		}
		return nil, fmt.Errorf("decode heroes: %w", err)
	}

	heroes := make([]Hero, 0, len(payload))
	for _, h := range payload {
		if h.Name == "" || h.LocalizedName == "" {
			continue
		}
		heroes = append(heroes, Hero{
			ID:          canonicalID(h.Name),
			Name:        h.LocalizedName,
			PrimaryRole: primaryRole(h.Roles),
			Roles:       append([]string(nil), h.Roles...),
			AttackType:  h.AttackType,
		})
	}

	if len(heroes) == 0 {
		if len(c.cache) > 0 {
			return cloneHeroes(c.cache), nil
		}
		return nil, errors.New("hero catalog returned no entries")
	}

	c.cache = heroes
	c.fetchedAt = time.Now()

	return cloneHeroes(c.cache), nil
}

type openDotaHero struct {
	Name          string   `json:"name"`
	LocalizedName string   `json:"localized_name"`
	Roles         []string `json:"roles"`
	AttackType    string   `json:"attack_type"`
}

func canonicalID(name string) string {
	name = strings.TrimPrefix(name, "npc_dota_hero_")
	return strings.ReplaceAll(name, " ", "_")
}

func primaryRole(roles []string) string {
	if len(roles) == 0 {
		return ""
	}
	return roles[0]
}

func cloneHeroes(list []Hero) []Hero {
	out := make([]Hero, len(list))
	copy(out, list)
	return out
}
