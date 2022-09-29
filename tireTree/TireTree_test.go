package tireTree

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTireTree_Insert(t *testing.T) {
	tree := generateTireTree("abc", "hel")
	assert.Equal(t, tree.Search("abc"), true)
	assert.Equal(t, tree.Search("hel"), true)
	assert.Equal(t, tree.Search("abcd"), false)
	assert.Equal(t, tree.Search("he"), false)
}

func TestTireTree_SearchAndDec(t *testing.T) {
	dict := map[string]int{"abc": 2, "dch": 1}
	tree := generateTireTreeWithMap(dict)
	assert.Equal(t, tree.SearchAndDec("abc"), true)
	assert.Equal(t, tree.SearchAndDec("abc"), true)
	assert.Equal(t, tree.SearchAndDec("abc"), false)
}
