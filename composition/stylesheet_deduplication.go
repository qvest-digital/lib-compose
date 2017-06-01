package composition

import (
	"golang.org/x/net/html"
)

// NOOP strategy.
// This strategy will insert all found stylesheets w/o any filtering.
type IdentityDeduplicationStrategy struct {
}

func (strategy *IdentityDeduplicationStrategy) Deduplicate(stylesheets [][]html.Attribute) [][]html.Attribute {
	return stylesheets
}

// Simple strategy
// Implements a very simple deduplication strategy. That is, it filters out
// stylesheets with duplicate href value.
type SimpleDeduplicationStrategy struct {
}

// Remove duplicate entries from hrefs.
func (strategy *SimpleDeduplicationStrategy) Deduplicate(stylesheets [][]html.Attribute) (result [][]html.Attribute) {
	knownHrefs := map[string]bool{}
	var meaninglessValue bool
	for _, stylesheetAttrs := range stylesheets {
		hrefAttr, attrExists := getAttr(stylesheetAttrs, "href")
		if !attrExists {
			continue
		}
		href := hrefAttr.Val
		_, known := knownHrefs[href]
		if !known {
			result = append(result, stylesheetAttrs)
			knownHrefs[href] = meaninglessValue
		}
	}
	return result
}
