package composition

import "strings"

// Initialization: Sets the default deduplication strategy to be used.
func init() {
	SetStrategy(new(IdentityDeduplicationStrategy))
}

// Set another deduplication strategy.
func SetStrategy(strategy StylesheetDeduplicationStrategy) {
	stylesheetDeduplicationStrategy = strategy
}

// NOOP strategy.
// This stragegy will insert all found stylesheets w/o any filtering.
type IdentityDeduplicationStrategy struct {
}

func (strategy *IdentityDeduplicationStrategy) Deduplicate(hrefs []string) []string {
	return hrefs
}

// Simple strategy
// Implements a very simple deduplication stragegy. That is, it filters out
// stylesheets with duplicate href value.
type SimpleDeduplicationStrategy struct {
}

// Remove duplicate entries from hrefs.
func (strategy *SimpleDeduplicationStrategy) Deduplicate(hrefs []string) []string {
	var knownHrefs string
	var result []string
	const delimiter = "-|-"
	for _, href := range hrefs {
		if !strings.Contains(knownHrefs, href) {
			result = append(result, href)
			knownHrefs += delimiter + href
		}
	}
	return result
}

// Variable to hold the active strategy
var stylesheetDeduplicationStrategy StylesheetDeduplicationStrategy
