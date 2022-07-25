package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	for hashName, makeHash := range map[string]func() Hash{
		"MD5":    NewMD5Hash,
		"SHA1":   NewSHA1Hash,
		"SHA256": NewSHA256Hash,
	} {
		t.Run(hashName, func(t *testing.T) {
			for tName, tCase := range map[string]func(t *testing.T, makeHash func() Hash){
				"AddingDataModifiesSum": func(t *testing.T, makeHash func() Hash) {
					h := makeHash()
					h.Add("foo")
					s0 := h.Sum()
					h.Add("bar")
					assert.NotEqual(t, s0, h.Sum())
				},
				"AddingSameValuesInDifferentOrderReturnsDifferentSum": func(t *testing.T, makeHash func() Hash) {
					h0 := makeHash()
					h0.Add("foo")
					h0.Add("bar")
					h1 := makeHash()
					h1.Add("bar")
					h1.Add("foo")
					assert.NotEqual(t, h0.Sum(), h1.Sum())
				},
				"SumSucceedsForNoData": func(t *testing.T, makeHash func() Hash) {
					assert.NotPanics(t, func() {
						makeHash().Sum()
					})
				},
				"SumReturnsConsistentValueForSameInput": func(t *testing.T, makeHash func() Hash) {
					h0 := makeHash()
					h0.Add("foo")
					h0.Add("bar")
					h1 := makeHash()
					h1.Add("foo")
					h1.Add("bar")
					assert.Equal(t, h0.Sum(), h1.Sum())
				},
			} {
				t.Run(tName, func(t *testing.T) {
					tCase(t, makeHash)
				})
			}
		})
	}
}
