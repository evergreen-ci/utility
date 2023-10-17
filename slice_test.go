package utility

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSliceIntersection(t *testing.T) {
	a := []string{"A", "B", "C", "D"}
	b := []string{"C", "D", "E"}

	assert.Equal(t, 2, len(StringSliceIntersection(a, b)))
	assert.Contains(t, StringSliceIntersection(a, b), "C")
	assert.Contains(t, StringSliceIntersection(a, b), "D")
}

func TestUniqueStrings(t *testing.T) {
	assert.EqualValues(t, []string{"a", "b", "c", "d", "e"},
		UniqueStrings([]string{"a", "b", "c", "a", "a", "d", "d", "e"}))

	assert.EqualValues(t, []string{"a", "b", "c"},
		UniqueStrings([]string{"a", "b", "c"}))
}

func TestSplitCommas(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T){
		"ReturnsUnmodifiedStringsWithoutCommas": func(t *testing.T) {
			input := []string{"foo", "bar", "bat"}
			assert.Equal(t, input, SplitCommas(input))
		},
		"ReturnsSplitCommaStrings": func(t *testing.T) {
			input := []string{"foo,bar", "bat", "baz,qux,quux"}
			expected := []string{"foo", "bar", "bat", "baz", "qux", "quux"}
			assert.Equal(t, expected, SplitCommas(input))
		},
	} {
		t.Run(testName, func(t *testing.T) {
			testCase(t)
		})
	}
}

func TestStringSliceSymmetricDifference(t *testing.T) {
	a := []string{"a", "c", "f", "n", "q"}
	b := []string{"q", "q", "g", "y", "a"}

	onlyA, onlyB := StringSliceSymmetricDifference(a, b)
	assert.Contains(t, onlyA, "c")
	assert.Contains(t, onlyA, "f")
	assert.Contains(t, onlyA, "n")
	assert.Len(t, onlyA, 3)
	assert.Contains(t, onlyB, "g")
	assert.Contains(t, onlyB, "y")
	assert.Len(t, onlyB, 2)

	onlyB, onlyA = StringSliceSymmetricDifference(b, a)
	assert.Contains(t, onlyA, "c")
	assert.Contains(t, onlyA, "f")
	assert.Contains(t, onlyA, "n")
	assert.Len(t, onlyA, 3)
	assert.Contains(t, onlyB, "g")
	assert.Contains(t, onlyB, "y")
	assert.Len(t, onlyB, 2)

	onlyA, onlyB = StringSliceSymmetricDifference(a, a)
	assert.Zero(t, onlyA)
	assert.Zero(t, onlyB)

	empty1, empty2 := StringSliceSymmetricDifference([]string{}, []string{})
	assert.Zero(t, empty1)
	assert.Zero(t, empty2)
}

func TestGetSetDifference(t *testing.T) {
	assert := assert.New(t)
	setA := []string{"one", "four", "five", "three", "two"}
	setB := []string{"five", "two"}
	difference := GetSetDifference(setA, setB)
	sort.Strings(difference)

	// GetSetDifference returns the elements in A that are not in B
	assert.Equal(3, len(difference))
	assert.Equal("four", difference[0])
	assert.Equal("one", difference[1])
	assert.Equal("three", difference[2])
}

func TestIndexOf(t *testing.T) {
	assert.Equal(t, 3, IndexOf([]string{"a", "b", "c", "d", "e"}, "d"))
	assert.Equal(t, 0, IndexOf([]string{"a", "b", "c", "d", "e"}, "a"))
	assert.Equal(t, -1, IndexOf([]string{"a", "b", "c", "d", "e"}, "f"))
	assert.Equal(t, -1, IndexOf([]string{"a", "b", "c", "d", "e"}, "1"))
	assert.Equal(t, -1, IndexOf([]string{"a", "b", "c", "d", "e"}, "æ"))
}

func TestStringMatchesAnyRegex(t *testing.T) {
	domains := []string{".*.corp.mongodb.com", "https://something.mongodb.com"}
	assert.Equal(t, true, StringMatchesAnyRegex("https://patch-analysis-ui.server-tig.staging.corp.mongodb.com", domains))
	assert.Equal(t, true, StringMatchesAnyRegex("https://something.mongodb.com", domains))
	assert.Equal(t, false, StringMatchesAnyRegex("corp.mongodb.com", domains))
	assert.Equal(t, false, StringMatchesAnyRegex("https://something-else.mongodb.com", domains))
}

func TestFilterSlice(t *testing.T) {
	stringTest := []string{"a", "b", "c", "d", "e"}
	assert.Equal(t, []string{"a", "b", "c"}, FilterSlice(stringTest, func(s string) bool {
		return s < "d"
	}))

	intTest := []int{1, 2, 3, 4, 5}
	assert.Equal(t, []int{1, 2, 3}, FilterSlice(intTest, func(i int) bool {
		return i < 4
	}))

	type SomeCustomType struct {
		name string
		age  int
	}

	customTypeTest := []SomeCustomType{
		{"Alice", 23},
		{"Bob", 25},
		{"Charlie", 27},
	}

	assert.Equal(t, []SomeCustomType{
		{"Alice", 23},
		{"Bob", 25},
	}, FilterSlice(customTypeTest, func(c SomeCustomType) bool {
		return c.age < 27
	}))
}

func TestContainsOrderedSubsetWithComparator(t *testing.T) {
	supersetInt := []int{0, 1, 2, 1, 5}
	lessThanComparator := func(super, sub int) bool {
		return sub < super
	}
	assert.True(t, ContainsOrderedSubsetWithComparator(supersetInt, []int{0, 1, 2}, lessThanComparator))
	assert.True(t, ContainsOrderedSubsetWithComparator(supersetInt, []int{-1, 0, 1, 0, 4}, lessThanComparator))
	assert.True(t, ContainsOrderedSubsetWithComparator(supersetInt, []int{-1, 0, 2}, lessThanComparator))
	assert.False(t, ContainsOrderedSubsetWithComparator(supersetInt, []int{2, 0}, lessThanComparator))
	assert.False(t, ContainsOrderedSubsetWithComparator(supersetInt, []int{1, 0, 0, 4}, lessThanComparator))
	assert.False(t, ContainsOrderedSubsetWithComparator(supersetInt, []int{-1, 1, 5}, lessThanComparator))

	// Larger subset than superset
	assert.False(t, ContainsOrderedSubsetWithComparator([]int{0, 1, 2}, []int{-1, 0, 1, 1}, lessThanComparator))

	// Empty slices
	assert.True(t, ContainsOrderedSubsetWithComparator([]int{0, 1}, []int{}, lessThanComparator))
	assert.False(t, ContainsOrderedSubsetWithComparator([]int{}, []int{0, 1}, lessThanComparator))
	assert.True(t, ContainsOrderedSubsetWithComparator([]int{}, []int{}, lessThanComparator))
}

func TestContainsOrderedSubset(t *testing.T) {
	supersetStrings := []string{"a", "b", "c", "b", "z"}
	assert.True(t, ContainsOrderedSubset(supersetStrings, []string{"a", "b", "c"}))
	assert.True(t, ContainsOrderedSubset(supersetStrings, []string{"b", "z"}))
	assert.True(t, ContainsOrderedSubset(supersetStrings, []string{"a", "b", "b"}))
	assert.True(t, ContainsOrderedSubset(supersetStrings, []string{"a", "c", "b"}))
	assert.False(t, ContainsOrderedSubset(supersetStrings, []string{"b", "b", "c"}))
	assert.False(t, ContainsOrderedSubset(supersetStrings, []string{"a", "c", "b", "b"}))
	assert.False(t, ContainsOrderedSubset(supersetStrings, []string{"b", "z", "b"}))
	assert.False(t, ContainsOrderedSubset(supersetStrings, []string{"c", "b", "a"}))

	supersetInts := []int{0, 1, 2, 1, 5}
	assert.True(t, ContainsOrderedSubset(supersetInts, []int{0, 1, 2}))
	assert.True(t, ContainsOrderedSubset(supersetInts, []int{1, 1, 5}))
	assert.True(t, ContainsOrderedSubset(supersetInts, []int{0, 2, 1, 5}))
	assert.True(t, ContainsOrderedSubset(supersetInts, []int{1, 2, 5}))
	assert.False(t, ContainsOrderedSubset(supersetInts, []int{0, 1, 1, 2}))
	assert.False(t, ContainsOrderedSubset(supersetInts, []int{0, 2, 1, 1}))
	assert.False(t, ContainsOrderedSubset(supersetInts, []int{1, 5, 1}))
	assert.False(t, ContainsOrderedSubset(supersetInts, []int{2, 1, 0}))

	// Larger subset than superset
	assert.False(t, ContainsOrderedSubset([]int{0, 1, 2}, []int{0, 1, 2, 3}))

	// Empty slices
	assert.True(t, ContainsOrderedSubset([]int{0, 1}, []int{}))
	assert.False(t, ContainsOrderedSubset([]int{}, []int{0, 1}))
	assert.True(t, ContainsOrderedSubset([]int{}, []int{}))
}

func TestStringSliceContainsOrderedPrefixSubset(t *testing.T) {
	supersetStrings := []string{"apples", "bananas", "cabbages", "balloons", "applets"}
	assert.True(t, StringSliceContainsOrderedPrefixSubset(supersetStrings, []string{"a", "b", "c"}))
	assert.True(t, StringSliceContainsOrderedPrefixSubset(supersetStrings, []string{"app", "ballo", "appl"}))
	assert.True(t, StringSliceContainsOrderedPrefixSubset(supersetStrings, []string{"app", "banan", "ball", "appl"}))
	assert.True(t, StringSliceContainsOrderedPrefixSubset(supersetStrings, []string{"app", "banan", "ball", "appl"}))
	assert.False(t, StringSliceContainsOrderedPrefixSubset(supersetStrings, []string{"apple", "ballo", "cab"}))
	assert.False(t, StringSliceContainsOrderedPrefixSubset(supersetStrings, []string{"ba", "ba", "cab"}))
	assert.False(t, StringSliceContainsOrderedPrefixSubset(supersetStrings, []string{"a", "cab", "ba", "cab"}))
}
