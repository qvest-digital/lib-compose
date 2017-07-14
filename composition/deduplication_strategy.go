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

func (strategy *IdentityDeduplicationStrategy) DeduplicateElements(scriptElements []ScriptElement) []ScriptElement {
	return scriptElements
}

// Simple strategy
// Implements a very simple deduplication strategy. That is, it filters out
// stylesheets with duplicate href value.
type SimpleDeduplicationStrategy struct {
}

func (strategy *SimpleDeduplicationStrategy) DeduplicateElements(scriptElements []ScriptElement) (result []ScriptElement) {
	knownSrc := map[string]struct{}{}
	for _, scriptElement := range scriptElements {

		// inline script, that is: it has no src attribute
		srcAttr, attrExists := getAttr(scriptElement.Attrs, "src")
		if !attrExists {
			if scriptElement.Text != nil {
				result = append(result, scriptElement)
			}
			continue
		}

		src := srcAttr.Val
		_, known := knownSrc[src]
		if !known {
			result = append(result, scriptElement)
			knownSrc[src] = struct{}{}
		}
	}
	return result
}

// Remove duplicate entries from hrefs.
func (strategy *SimpleDeduplicationStrategy) Deduplicate(stylesheets [][]html.Attribute) (result [][]html.Attribute) {
	knownHrefs := map[string]struct{}{}
	for _, stylesheetAttrs := range stylesheets {
		hrefAttr, attrExists := getAttr(stylesheetAttrs, "href")
		if !attrExists {
			continue
		}
		href := hrefAttr.Val
		_, known := knownHrefs[href]
		if !known {
			result = append(result, stylesheetAttrs)
			knownHrefs[href] = struct{}{}
		}
	}
	return result
}
