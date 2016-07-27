package composition

import (
	"github.com/tarent/lib-servicediscovery/servicediscovery"
)

// Fluent-interface decorator for the FetchDefinition that activates the ServiceDiscovery
func (d *FetchDefinition) DiscoveredBy(dnsServer string) *FetchDefinition {

	serviceDiscovery, err := servicediscovery.NewConsulServiceDiscovery(dnsServer)
	if err != nil {
		panic(err)
	}

	d.ServiceDiscovery = serviceDiscovery
	d.ServiceDiscoveryActive = true

	return d
}
