package composition

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func Test_StringFragment_Variables(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		data     map[string]interface{}
		template string
		expected string
	}{
		{
			data:     map[string]interface{}{},
			template: "xxx",
			expected: "xxx",
		},
		{
			data:     map[string]interface{}{},
			template: "",
			expected: "",
		},
		{
			data:     map[string]interface{}{"foo": "bar"},
			template: "§[foo]§",
			expected: "bar",
		},
		{
			data:     map[string]interface{}{"foo": "bar"},
			template: "xxx-§[foo]§-yyy",
			expected: "xxx-bar-yyy",
		},
		{
			data:     map[string]interface{}{"foo": "bar", "bli": "blub"},
			template: "xxx-§[foo]§-yyy-§[bli]§-zzz",
			expected: "xxx-bar-yyy-blub-zzz",
		},
		{
			data:     map[string]interface{}{},
			template: "xxx-§[not_existent_variable]§-yyy",
			expected: "xxx--yyy",
		},
	}

	for _, test := range tests {

		f := StringFragment(test.template)
		buf := bytes.NewBufferString("")
		err := f.Execute(buf, test.data, nil)
		a.NoError(err)

		a.Equal(test.expected, buf.String())
	}
}

func Test_StringFragment_Includes(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		fragments map[string]string
		template  string
		expected  string
	}{
		{
			fragments: map[string]string{"foo": "bar"},
			template:  "§[> foo]§",
			expected:  "bar",
		},
		{
			fragments: map[string]string{"foo": "bar"},
			template:  "xxx-§[> foo]§-yyy",
			expected:  "xxx-bar-yyy",
		},
		{
			fragments: map[string]string{"foo": "bar", "bli": "blub"},
			template:  "xxx-§[> foo]§-yyy-§[> bli]§-zzz",
			expected:  "xxx-bar-yyy-blub-zzz",
		},
		{
			fragments: map[string]string{},
			template:  "xxx-§[> not_existent_fragment]§-yyy",
			expected:  "xxx--yyy",
		},
	}

	for _, test := range tests {
		f := StringFragment(test.template)
		buf := bytes.NewBufferString("")
		executeNestedFragment := func(nestedFragmentName string) error {
			if val, exist := test.fragments[nestedFragmentName]; exist {
				io.WriteString(buf, val)
			}
			return nil
		}
		err := f.Execute(buf, nil, executeNestedFragment)
		a.NoError(err)

		a.Equal(test.expected, buf.String())
	}
}
