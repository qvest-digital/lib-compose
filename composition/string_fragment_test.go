package composition

import (
	"bytes"
	"errors"
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
			data:     map[string]interface{}{"foo": map[string]interface{}{"bar": "bazz"}},
			template: "§[foo.bar]§",
			expected: "bazz",
		},
		{
			data:     map[string]interface{}{"foo": map[string]interface{}{"bar": "bazz"}, "foo.bar": "overwrite"},
			template: "§[foo.bar]§",
			expected: "overwrite",
		},
		{
			data:     map[string]interface{}{"foo": map[string]interface{}{"bar": "bazz"}},
			template: "§[foo.bar.nothing]§",
			expected: "",
		},
		{
			data:     map[string]interface{}{"foo": "bar"},
			template: "§[ foo ]§",
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
		{
			data:     map[string]interface{}{},
			template: "xxx-]§-yyy",
			expected: "xxx-]§-yyy",
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
		fragments   map[string]string
		template    string
		expected    string
		expectedErr error
	}{
		{
			fragments: map[string]string{"foo": "bar"},
			template:  "§[> foo]§",
			expected:  "bar",
		},
		{
			fragments: map[string]string{"foo": "bar"},
			template:  "§[>   foo   ]§",
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
			fragments:   map[string]string{},
			template:    "xxx-§[> not_existent_fragment]§-yyy",
			expected:    "xxx-",
			expectedErr: errors.New("Fragment does not exist: not_existent_fragment"),
		},
	}

	for _, test := range tests {
		f := StringFragment(test.template)
		buf := bytes.NewBufferString("")
		executeNestedFragment := func(nestedFragmentName string) error {
			if val, exist := test.fragments[nestedFragmentName]; exist {
				io.WriteString(buf, val)
				return nil
			}
			return errors.New("Fragment does not exist: " + nestedFragmentName)
		}
		err := f.Execute(buf, nil, executeNestedFragment)

		a.Equal(test.expected, buf.String())
		if test.expectedErr == nil {
			a.NoError(err)
		} else {
			a.Equal(test.expectedErr, err)
		}

	}
}

func Test_StringFragment_ParsingErrors(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		template          string
		expectedErrString string
	}{
		{
			template:          "xxx-§[-yyy",
			expectedErrString: "Fragment Parsing error, missing ending separator:",
		},
		{
			template:          "xxx-]§§[-yyy",
			expectedErrString: "Fragment Parsing error, missing ending separator:",
		},
	}

	for _, test := range tests {
		f := StringFragment(test.template)
		buf := bytes.NewBufferString("")
		executeNestedFragment := func(nestedFragmentName string) error {
			return nil
		}
		err := f.Execute(buf, map[string]interface{}{}, executeNestedFragment)
		a.Error(err)
		a.Contains(err.Error(), test.expectedErrString)
	}
}
