package regPlus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestList_Insert(t *testing.T) {
	list := &list{}
	back := list.PushBack(3)
	front := list.PushFront(1)
	list.PushBack(5)
	assert.Equal(t, list.Collection(), []interface{}{1, 3, 5})

	fmt.Println("======================================================")
	front.MoveElementBefore(back)
	assert.Equal(t, list.Collection(), []interface{}{3, 1, 5})

	back.RemoveSelf()
	assert.Equal(t, list.Collection(), []interface{}{1, 5})

	front.InsertElementAfter(back)
	assert.Equal(t, list.Collection(), []interface{}{1, 3, 5})
}

func TestList_CheckCycle(t *testing.T) {
	list := &list{}
	list.PushFront(2)
	assert.Equal(t, list.Front(), list.Back())
}
