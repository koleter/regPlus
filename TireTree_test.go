package regPlus

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTireTree_Insert(t *testing.T) {
	tree := node{}
	tree.Insert("abc", "hel")
	assert.Equal(t, tree.Search("abc"), true)
	assert.Equal(t, tree.Search("hel"), true)
	assert.Equal(t, tree.Search("abcd"), false)
	assert.Equal(t, tree.Search("he"), false)
}
