package composition

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_FetchDefinition_DiscoveredBy(t *testing.T) {
	a := assert.New(t)

	testSubject := FetchDefinition{}

	testSubject.DiscoveredBy("127.0.0.1:53")

	a.NotNil(testSubject.ServiceDiscovery)
	a.True(testSubject.ServiceDiscoveryActive)
}

func Test_FetchDefinition_DiscoveredByError(t *testing.T) {
	a := assert.New(t)

	testSubject := FetchDefinition{}

	a.Panics(func() {
		testSubject.DiscoveredBy("a")
	})
}
