package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DefaultTTL is applied to any DNS record without an explicit TTL.
const DefaultTTL = uint32(300)

// DNSConfig describes the desired DNS configuration, usually loaded from dnsconfig.js.
type DNSConfig struct {
	Registrars         []*RegistrarConfig            `json:"registrars"`
	DNSProviders       []*DNSProviderConfig          `json:"dns_providers"`
	Domains            []*DomainConfig               `json:"domains"`
	RegistrarsByName   map[string]*RegistrarConfig   `json:"-"`
	DNSProvidersByName map[string]*DNSProviderConfig `json:"-"`
	SkipRecordAudit    bool                          `json:"skiprecordaudit,omitempty"`
}

// FindDomain returns the *DomainConfig for domain query in config.
func (config *DNSConfig) FindDomain(query string) *DomainConfig {
	for _, b := range config.Domains {
		if b.Name == query {
			return b
		}
	}
	return nil
}

// RegistrarConfig describes a registrar.
type RegistrarConfig struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Metadata json.RawMessage `json:"meta,omitempty"`
}

// DNSProviderConfig describes a DNS service provider.
type DNSProviderConfig struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Metadata json.RawMessage `json:"meta,omitempty"`
}

// Nameserver describes a nameserver.
type Nameserver struct {
	Name string `json:"name"` // Normalized to a FQDN with NO trailing "."
	// NB(tlim): DomainConfig.Nameservers are stored WITH a trailing "." (Sorry!)
}

// FIXME(tal): In hindsight the Nameserver struct is overkill. We
// could have just used string.  Currently there are very few places
// that refer to the .Name field directly.  All new code should use
// ToNameservers/ToNameserversStripTD and NameserversToStrings to make
// future refactoring easier.  See
// https://github.com/StackExchange/dnscontrol/issues/577

func (n *Nameserver) String() string {
	return n.Name
}

// ToNameservers turns a list of strings into a list of Nameservers.
// It is an error if any string has a trailing dot. Either remove the
// trailing dot before you call this or (much preferred) use ToNameserversStripTD.
func ToNameservers(nss []string) ([]*Nameserver, error) {
	nservers := []*Nameserver{}
	for _, ns := range nss {
		if strings.HasSuffix(ns, ".") {
			return nil, fmt.Errorf("provider code leaves trailing dot on nameserver")
			// If you see this error, maybe the provider should call
			// ToNameserversStripTD instead.
		}
		nservers = append(nservers, &Nameserver{Name: ns})
	}
	return nservers, nil
}

// ToNameserversStripTD is like ToNameservers but strips the trailing
// dot from each item. It is an error if there is no trailing dot.
func ToNameserversStripTD(nss []string) ([]*Nameserver, error) {
	nservers := []*Nameserver{}
	for _, ns := range nss {
		if !strings.HasSuffix(ns, ".") {
			return nil, fmt.Errorf("provider code already removed nameserver trailing dot (%v)", ns)
			// If you see this error, maybe the provider should call ToNameservers instead.
		}
		nservers = append(nservers, &Nameserver{Name: ns[0 : len(ns)-1]})
	}
	return nservers, nil
}

// NameserversToStrings constructs a list of strings from *Nameserver structs
func NameserversToStrings(nss []*Nameserver) (s []string) {
	for _, ns := range nss {
		s = append(s, ns.Name)
	}
	return s
}

// Correction is anything that can be run. Implementation is up to the specific provider.
type Correction struct {
	F   func() error `json:"-"`
	Msg string
}

// DomainContainingFQDN finds the best domain from the dns config for the given record fqdn.
// It will chose the domain whose name is the longest suffix match for the fqdn.
func (config *DNSConfig) DomainContainingFQDN(fqdn string) *DomainConfig {
	fqdn = strings.TrimSuffix(fqdn, ".")
	longestLength := 0
	var d *DomainConfig
	for _, dom := range config.Domains {
		if (dom.Name == fqdn || strings.HasSuffix(fqdn, "."+dom.Name)) && len(dom.Name) > longestLength {
			longestLength = len(dom.Name)
			d = dom
		}
	}
	return d
}

//// IgnoreName describes an IGNORE_NAME rule.
//type IgnoreName struct {
//	Pattern string `json:"pattern"` // Glob pattern.
//	Types   string `json:"types"`   // All caps rtype names, comma separated.
//}
//
//// IgnoreTarget describes an IGNORE_TARGET rule.
//type IgnoreTarget struct {
//	Pattern string `json:"pattern"` // Glob pattern.
//	Type    string `json:"type"`    // All caps rtype name.
//}
//
//func (i *IgnoreTarget) String() string {
//	return i.Pattern
//}
