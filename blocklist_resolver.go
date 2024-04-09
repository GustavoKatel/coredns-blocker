package blocker

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type BlocklistResolver struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	uri             string
	refreshInterval time.Duration

	subscribers     map[chan string]struct{}
	subscribersLock sync.RWMutex

	scheduleUpdate chan struct{}

	logger Logger
}

func NewBlocklistResolver(ctx context.Context, uri string, refreshInterval time.Duration, logger Logger) (*BlocklistResolver, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &BlocklistResolver{
		ctx:       ctx,
		ctxCancel: cancel,

		uri: uri,

		refreshInterval: refreshInterval,

		subscribers:     map[chan string]struct{}{},
		subscribersLock: sync.RWMutex{},

		scheduleUpdate: make(chan struct{}, 1),

		logger: logger,
	}, nil
}

func (resolver *BlocklistResolver) Start() {
	ticker := time.NewTicker(resolver.refreshInterval)

	go func() {
		for {
			select {
			case <-resolver.ctx.Done():
				ticker.Stop()
				return
			case <-resolver.scheduleUpdate:
				// Update the blocklist
				resolver.update()
			case <-ticker.C:
				// Update the blocklist
				resolver.update()
			}
		}
	}()
}

func (resolver *BlocklistResolver) Stop() {
	resolver.ctxCancel()
}

func (resolver *BlocklistResolver) ScheduleUpdate() {
	resolver.scheduleUpdate <- struct{}{}
}

func (resolver *BlocklistResolver) Subscribe(ch chan string) {
	resolver.subscribersLock.Lock()
	defer resolver.subscribersLock.Unlock()

	resolver.subscribers[ch] = struct{}{}
}
func (resolver *BlocklistResolver) update() {
	var contents string
	var err error

	if strings.HasPrefix(resolver.uri, "http") {
		contents, err = resolver.readFromRemote()
	} else {
		contents, err = resolver.readFromFile()
	}

	if err != nil {
		resolver.logger.Errorf("error reading blocklist: %v", err)
		return
	}

	resolver.notifySubscribers(contents)
}

func (resolver *BlocklistResolver) readFromFile() (string, error) {
	resolver.logger.Infof("reading blocklist from %s", resolver.uri)

	file, err := os.ReadFile(resolver.uri)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func (resolver *BlocklistResolver) readFromRemote() (string, error) {
	resolver.logger.Infof("fetching blocklist from %s", resolver.uri)

	ctx, cancel := context.WithTimeout(resolver.ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", resolver.uri, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "CoreDNS/Blocker/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func (resolver *BlocklistResolver) notifySubscribers(contents string) {
	resolver.subscribersLock.RLock()
	defer resolver.subscribersLock.RUnlock()

	for ch := range resolver.subscribers {
		select {
		case <-resolver.ctx.Done():
			return
		case ch <- contents:
		}
	}
}
