package composition

import (
	"fmt"
	"io"
	"strings"
)

const PlaceholderStart = "§["
const PlaceholderEnd = "]§"

// StringFragment is a simple template based representation of a fragment.
// On Execute(), the following replacements will be done:
// §[ aVariable ] inserts a variable from the data map
// §[> fragment ] executed a nexted fragment by executeNestedFragment()
type StringFragment string

func (f StringFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	t := string(f)
	for len(t) > 0 {
		start := strings.Index(t, PlaceholderStart)
		if start > -1 {
			end := strings.Index(t, PlaceholderEnd)
			if end < start {
				return fmt.Errorf("Fragment Parsing error, missing ending separator: %v", f)
			}
			io.WriteString(w, t[:start])
			placeholder := t[start+len(PlaceholderStart) : end]
			f.writePlaceholder(w, placeholder, data, executeNestedFragment)
			t = t[end+len(PlaceholderEnd):]
		} else {
			io.WriteString(w, t)
			t = ""
		}
	}
	return nil
}

func (f StringFragment) writePlaceholder(w io.Writer, placeholder string, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	if len(placeholder) > 1 && placeholder[0] == byte('>') {
		placeholder = strings.TrimSpace(placeholder[1:])
		executeNestedFragment(placeholder)
	} else {
		placeholder = strings.TrimSpace(placeholder)
		if d, exist := data[placeholder]; exist {
			io.WriteString(w, fmt.Sprintf("%v", d))
		}
	}
	return nil
}

func (f StringFragment) String() string {
	return string(f)
}
