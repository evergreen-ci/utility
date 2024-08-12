package utility

import (
	"regexp"
	"sort"
	"strings"
)

// StringSliceContains determines if a string is in a slice.
func StringSliceContains(slice []string, item string) bool {
	if len(slice) == 0 {
		return false
	}

	for idx := range slice {
		if slice[idx] == item {
			return true
		}
	}

	return false
}

// StringSliceIntersection returns the intersecting elements of slices a and b.
func StringSliceIntersection(a, b []string) []string {
	inA := map[string]bool{}
	var out []string
	for _, elem := range a {
		inA[elem] = true
	}
	for _, elem := range b {
		if inA[elem] {
			out = append(out, elem)
		}
	}
	return out
}

// StringSliceSymmetricDifference returns only elements not in common between 2 slices
// (ie. inverse of the intersection).
func StringSliceSymmetricDifference(a, b []string) ([]string, []string) {
	mapA := map[string]bool{}
	mapAcopy := map[string]bool{}
	for _, elem := range a {
		mapA[elem] = true
		mapAcopy[elem] = true
	}
	var inB []string
	for _, elem := range b {
		if mapAcopy[elem] { // need to delete from the copy in case B has duplicates of the same value in A
			delete(mapA, elem)
		} else {
			inB = append(inB, elem)
		}
	}
	var inA []string
	for elem := range mapA {
		inA = append(inA, elem)
	}
	return inA, inB
}

// UniqueStrings takes a slice of strings and returns a new slice with duplicates removed.
// Order is preserved.
func UniqueStrings(slice []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range slice {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// SplitCommas returns the slice of strings after splitting each string by
// commas.
func SplitCommas(originals []string) []string {
	splitted := []string{}
	for _, original := range originals {
		splitted = append(splitted, strings.Split(original, ",")...)
	}
	return splitted
}

// GetSetDifference returns the elements in A that are not in B.
func GetSetDifference(a, b []string) []string {
	setB := make(map[string]struct{})
	setDifference := make(map[string]struct{})

	for _, e := range b {
		setB[e] = struct{}{}
	}
	for _, e := range a {
		if _, ok := setB[e]; !ok {
			setDifference[e] = struct{}{}
		}
	}

	d := make([]string, 0, len(setDifference))
	for k := range setDifference {
		d = append(d, k)
	}

	return d
}

// IndexOf returns the first occurrence of a string in a sorted array.
func IndexOf(a []string, toFind string) int {
	i := sort.Search(len(a), func(index int) bool {
		return strings.Compare(a[index], toFind) >= 0
	})
	if i < 0 || i >= len(a) {
		return -1
	}
	if a[i] == toFind {
		return i
	}
	return -1
}

// StringMatchesAnyRegex determines if the string item matches any regex in
// the slice.
func StringMatchesAnyRegex(item string, regexps []string) bool {
	for _, re := range regexps {
		matched, err := regexp.MatchString(re, item)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// FilterSlice filters a slice of elements based on a filter function.
func FilterSlice[T any](slice []T, filterFunction func(T) bool) []T {
	var filteredSlice []T
	for _, item := range slice {
		if filterFunction(item) {
			filteredSlice = append(filteredSlice, item)
		}
	}
	return filteredSlice
}

// ContainsOrderedSubsetWithComparator returns whether a slice
// contains an ordered subset using the given compare function.
func ContainsOrderedSubsetWithComparator[T any](superset, subset []T, compare func(T, T) bool) bool {
	if len(superset) < len(subset) {
		return false
	}

	var j int
	for i := 0; i < len(superset) && j < len(subset); i++ {
		if compare(superset[i], subset[j]) {
			j++
		}
	}

	return len(subset) == j
}

// ContainsOrderedSubset returns whether a slice
// contains an ordered subset using the equality
// operator.
func ContainsOrderedSubset[T comparable](superset, subset []T) bool {
	return ContainsOrderedSubsetWithComparator(superset, subset, func(a, b T) bool {
		return a == b
	})
}

// StringSliceContainsOrderedPrefixSubset returns whether a slice
// contains an ordered subset of prefixes using strings.HasPrefix.
func StringSliceContainsOrderedPrefixSubset(superset, subset []string) bool {
	return ContainsOrderedSubsetWithComparator(superset, subset, func(a, b string) bool {
		return strings.HasPrefix(a, b)
	})
}

// SliceBatches partitions the elems slice into batches with at most
// maxElemsPerBatch in each batch. If maxElemsPerBatch is not a positive
// integer, it will make batches of size 1.
func MakeSliceBatches[T any](elems []T, maxElemsPerBatch int) [][]T {
	if len(elems) == 0 {
		return nil
	}
	if maxElemsPerBatch <= 0 {
		maxElemsPerBatch = 1
	}

	remainingElems := elems
	var batches [][]T
	for len(remainingElems) > maxElemsPerBatch {
		batches = append(batches, remainingElems[0:maxElemsPerBatch:maxElemsPerBatch])
		remainingElems = remainingElems[maxElemsPerBatch:]
	}
	batches = append(batches, remainingElems)

	return batches
}
