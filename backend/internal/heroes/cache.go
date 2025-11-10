package heroes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const heroStatsURL = "https://api.opendota.com/api/heroStats"
const refreshInterval = 24 * time.Hour

var (
	cacheMu sync.RWMutex
	cache   []Hero
	client  = &http.Client{Timeout: 10 * time.Second}
)

// Init loads the hero catalog and starts a background refresher that keeps the data
// in sync with OpenDota. It must be called before accessing the hero cache.
func Init() error {
	heroes, err := fetch()
	if err != nil {
		return err
	}

	cacheMu.Lock()
	cache = heroes
	cacheMu.Unlock()

	go refresher()

	return nil
}

// All returns a copy of the cached hero slice.
func All() []Hero {
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	result := make([]Hero, len(cache))
	copy(result, cache)
	return result
}

func refresher() {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		heroes, err := fetch()
		if err != nil {
			log.Printf("heroes: refresher failed to pull hero stats: %v", err)
			continue
		}

		cacheMu.Lock()
		cache = heroes
		cacheMu.Unlock()
	}
}

func fetch() ([]Hero, error) {
	req, err := http.NewRequest(http.MethodGet, heroStatsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from hero stats API: %d", resp.StatusCode)
	}

	var heroes []Hero
	if err := json.NewDecoder(resp.Body).Decode(&heroes); err != nil {
		return nil, err
	}

	if len(heroes) == 0 {
		return nil, fmt.Errorf("hero stats API returned no heroes")
	}

	return heroes, nil
}
