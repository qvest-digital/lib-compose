package composition

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_MemoryContent_MemorySize(t *testing.T) {
	a := assert.New(t)

	m := MemoryContent{
		meta: map[string]interface{}{"foo": "bar"}, // 20
		head: NewStringFragment("0123456789"),         // 10
		body: map[string]Fragment{
			"a": NewStringFragment("0123456789"), // 10
			"b": NewStringFragment("0123456789"), // 10
		},
		tail:       NewStringFragment("0123456789"), // 10
		httpHeader: http.Header{"foo": {"bar"}},  // 20
	}

	a.Equal(80, m.MemorySize())
}
