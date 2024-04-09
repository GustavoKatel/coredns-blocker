package blocker

import (
	"context"
	"fmt"
	"time"
)

// BlockDomainsDecider is the interface which must be implemented by any type which intends to
// become a blocker. The purpose of each of the functions is described below.
type BlockDomainsDecider interface {
	IsDomainBlocked(domain string) bool
	UpdateBlocklist(contents string) error
}

type BlocklistType string

const BlocklistType_Hosts BlocklistType = "hosts"
const BlocklistType_ABP BlocklistType = "abp"

// PrepareBlocklist ...
func PrepareBlocklist(uri string, blocklistUpdateFrequency string, blocklistType string, logger Logger) (BlockDomainsDecider, []func() error, error) {
	frequency, err := time.ParseDuration(blocklistUpdateFrequency)
	if err != nil {
		return nil, nil, err
	}

	resolver, err := NewBlocklistResolver(context.Background(), uri, frequency, logger)
	if err != nil {
		return nil, nil, err
	}

	var decider BlockDomainsDecider
	switch BlocklistType(blocklistType) {
	case BlocklistType_Hosts:
		decider = NewBlockDomainsDeciderHosts(resolver, logger)
	case BlocklistType_ABP:
		decider = NewBlockDomainsDeciderABP(resolver, logger)
	}

	// Always update the blocklist when the server starts up
	resolver.ScheduleUpdate()

	// Setup periodic updation of the blocklist
	resolver.Start()

	stopResolver := func() error {
		fmt.Println("[INFO] Ticker was stopped.")
		resolver.Stop()
		return nil
	}

	shutdownHooks := []func() error{
		stopResolver,
	}

	return decider, shutdownHooks, nil
}
