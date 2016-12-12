package composition

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
)

const (
	PlaceholderStart  = "§["
	PlaceholderEnd    = "]§"
	StartInclude      = ">"
	StartIncludeBlock = "#>"
	EndIncludeBlock   = "/"
)

// Write a template to an output stream.
// The following replacements will be done:
// §[ aVariable ]§ inserts a variable from the data map
// §[> fragment ]§ executes a nested fragment by executeNestedFragment() and fails on error
// §[#> fragment ]§ alt text §[/fragment]§ executes a nested fragment by executeNestedFragment().
//                  On Error, the alternative Text within the block will be executed.
func executeTemplate(w io.Writer, template string, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	t := template
	for len(t) > 0 {
		start := strings.Index(t, PlaceholderStart)
		if start > -1 {
			end := strings.Index(t, PlaceholderEnd)
			if end < start {
				return fmt.Errorf("Fragment parsing error, missing ending separator: %v", template)
			}
			io.WriteString(w, t[:start])
			placeholder := t[start+len(PlaceholderStart) : end]

			if strings.HasPrefix(placeholder, StartIncludeBlock) {
				placeholder = strings.TrimSpace(strings.TrimPrefix(placeholder, StartIncludeBlock))
				blockEndText := PlaceholderStart + EndIncludeBlock + placeholder + PlaceholderEnd
				blockEndTextPosition := strings.Index(t, blockEndText)
				if blockEndTextPosition < end {
					return fmt.Errorf("Fragment parsing error, missing ending block: %v", blockEndText)
				}
				if err := executeNestedFragment(placeholder); err != nil {
					io.WriteString(w, t[end+len(PlaceholderEnd):blockEndTextPosition])
				}
				t = t[blockEndTextPosition+len(blockEndText):]
			} else {
				if err := writePlaceholder(w, placeholder, data, executeNestedFragment); err != nil {
					return err
				}
				t = t[end+len(PlaceholderEnd):]
			}
		} else {
			io.WriteString(w, t)
			t = ""
		}
	}
	return nil
}

func expandTemplateVars(template string, data map[string]interface{}) (string, error) {
	buff := bytes.NewBufferString("")
	err := executeTemplate(buff, template, data, nil)
	return buff.String(), err
}

func writePlaceholder(w io.Writer, placeholder string, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	placeholder = strings.TrimSpace(placeholder)
	if strings.HasPrefix(placeholder, StartInclude) {
		placeholder = strings.TrimSpace(strings.TrimPrefix(placeholder, StartInclude))
		if err := executeNestedFragment(placeholder); err != nil {
			return err
		}
	} else {
		if d, exist := getDataFromMap(data, placeholder); exist {
			io.WriteString(w, fmt.Sprintf("%v", d))
		}
	}
	return nil
}

// getDataFromMap returns the data defined by a key,
// where the key may contain multiple path elements separated by a '.'.
// If the map contains the full key on top-level, this value is preferred.
func getDataFromMap(data map[string]interface{}, key string) (result interface{}, exist bool) {
	if data == nil {
		return nil, false
	}

	if d, exist := data[key]; exist {
		return d, true
	}

	result = data
	parts := strings.Split(key, ".")
	for _, part := range parts {
		switch resultMap := result.(type) {
		case map[string]interface{}:
			if d, exist := resultMap[part]; exist {
				result = d
			} else {
				return nil, false
			}
		case url.Values: // Get parameters from an url
			result = resultMap.Get(part)
		default:
			return nil, false
		}
	}
	return result, true
}
