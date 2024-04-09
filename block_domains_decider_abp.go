package blocker

import (
	"bufio"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type BlockDomainsDeciderABP struct {
	blocklistResolver *BlocklistResolver
	lastUpdated       time.Time
	log               Logger

	blocklistLock sync.RWMutex
	blocklist     map[string]bool
}

// Name ...
func NewBlockDomainsDeciderABP(blocklistResolver *BlocklistResolver, logger Logger) BlockDomainsDecider {
	d := &BlockDomainsDeciderABP{
		blocklistResolver: blocklistResolver,
		log:               logger,

		blocklistLock: sync.RWMutex{},
		blocklist:     map[string]bool{},
	}

	ch := make(chan string, 1)
	blocklistResolver.Subscribe(ch)
	go func() {
		for contents := range ch {
			d.log.Infof("received updated blocklist contents. Length %d", len(contents))
			d.UpdateBlocklist(contents)
		}
	}()

	return d
}

// IsDomainBlocked ...
func (d *BlockDomainsDeciderABP) IsDomainBlocked(domain string) bool {
	d.blocklistLock.RLock()
	defer d.blocklistLock.RUnlock()

	// We will check every subdomain of the given domain against the blocklist. i.e. if example.com
	// is blocked, then every subdomain of that (subdomain.example.com, sub1.sub2.example.com) is
	// blocked. However, example.com.org is not blocked.
	comps := strings.Split(domain, ".")
	current := comps[len(comps)-1]
	for i := len(comps) - 2; i >= 0; i-- {
		newCurrent := strings.Join([]string{
			comps[i],
			current,
		}, ".")

		if d.blocklist[newCurrent] {
			return true
		}

		current = newCurrent
	}

	return false
}

// UpdateBlocklist ...
func (d *BlockDomainsDeciderABP) UpdateBlocklist(contents string) error {
	d.blocklistLock.Lock()
	defer d.blocklistLock.Unlock()

	// Update process
	numBlockedDomainsBefore := len(d.blocklist)
	lastUpdatedBefore := d.lastUpdated

	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		hostLine := scanner.Text()
		if !strings.HasPrefix(hostLine, "||") || !strings.HasSuffix(hostLine, "^") {
			d.log.Warningf("line \"%s\" does not match parseable ABP syntax subset", hostLine)
			continue
		}

		hostLine = strings.TrimPrefix(hostLine, "||")
		hostLine = strings.TrimSuffix(hostLine, "^")
		d.blocklist[dns.Fqdn(hostLine)] = true
	}

	d.lastUpdated = time.Now()

	d.log.Infof("updated blocklist; blocked domains: before: %d, after: %d; last updated: before: %v, after: %v",
		numBlockedDomainsBefore, len(d.blocklist), lastUpdatedBefore, d.lastUpdated)
	blocklistSize.WithLabelValues(d.blocklistResolver.uri).Set(float64(len(d.blocklist)))

	return nil
}
