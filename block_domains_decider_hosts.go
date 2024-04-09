package blocker

import (
	"bufio"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type BlockDomainsDeciderHosts struct {
	blocklistResolver *BlocklistResolver
	lastUpdated       time.Time
	log               Logger

	blocklistLock sync.RWMutex
	blocklist     map[string]bool
}

// Name ...
func NewBlockDomainsDeciderHosts(resolver *BlocklistResolver, logger Logger) BlockDomainsDecider {
	d := &BlockDomainsDeciderHosts{
		blocklistResolver: resolver,
		log:               logger,

		blocklistLock: sync.RWMutex{},
		blocklist:     map[string]bool{},
	}

	ch := make(chan string, 1)
	resolver.Subscribe(ch)

	go func() {
		for contents := range ch {
			d.log.Infof("received updated blocklist contents. Length %d", len(contents))
			d.UpdateBlocklist(contents)
		}
	}()

	return d
}

// IsDomainBlocked ...
func (d *BlockDomainsDeciderHosts) IsDomainBlocked(domain string) bool {
	d.blocklistLock.RLock()
	defer d.blocklistLock.RUnlock()

	return d.blocklist[domain]
}

// UpdateBlocklist ...
func (d *BlockDomainsDeciderHosts) UpdateBlocklist(contents string) error {
	d.blocklistLock.Lock()
	defer d.blocklistLock.Unlock()

	// Update process
	blocklistContent := strings.NewReader(contents)

	numBlockedDomainsBefore := len(d.blocklist)
	lastUpdatedBefore := d.lastUpdated

	scanner := bufio.NewScanner(blocklistContent)
	for scanner.Scan() {
		hostLine := scanner.Text()

		hostLine = strings.TrimSpace(hostLine)

		if hostLine == "" || strings.HasPrefix(hostLine, "#") {
			// Comment line
			continue
		}

		comps := strings.Split(hostLine, " ")
		if len(comps) < 2 {
			// Bad line in the input file
			d.log.Warningf("unformatted line present in the input file: %s", hostLine)
			continue
		}

		domain := comps[1]
		d.blocklist[dns.Fqdn(domain)] = true
	}

	d.lastUpdated = time.Now()

	d.log.Infof("updated blocklist; blocked domains: before: %d, after: %d; last updated: before: %v, after: %v",
		numBlockedDomainsBefore, len(d.blocklist), lastUpdatedBefore, d.lastUpdated)

	blocklistSize.WithLabelValues(d.blocklistResolver.uri).Set(float64(len(d.blocklist)))

	return nil
}
