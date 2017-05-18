package composition

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DefaultDeduplicationStrategy(t *testing.T) {
	a := assert.New(t)
	stylesheets := []string{"a", "b"}
	result := stylesheetDeduplicationStrategy.Deduplicate(stylesheets)
	a.EqualValues(stylesheets, result)
}

func Test_SimpleDeduplicationStrategy(t *testing.T) {
	a := assert.New(t)
	stylesheets := []string{"a", "b", "a", "b", "c", "a"}
	deduper := new(SimpleDeduplicationStrategy)
	result := deduper.Deduplicate(stylesheets)
	a.EqualValues([]string{"a", "b", "c"}, result)
}

type Strategy struct {
	collecttion []string
}

func (strategy *Strategy) Deduplicate(hrefs []string) []string {
	newhrefs := []string{}
	for _, href := range hrefs {
		newhrefs = append(newhrefs, href+"?abcdef")
		strategy.collecttion = append(strategy.collecttion, href)
	}
	return newhrefs
}

func Test_OwnDeduplicationStrategy(t *testing.T) {
	strategy := new(Strategy)
	SetStrategy(strategy)

	a := assert.New(t)
	stylesheets := []string{"a", "b"}
	expected := []string{"a?abcdef", "b?abcdef"}
	result := stylesheetDeduplicationStrategy.Deduplicate(stylesheets)
	a.EqualValues(expected, result)
	a.EqualValues(strategy.collecttion, stylesheets)
}
