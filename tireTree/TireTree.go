package tireTree

type node struct {
	next map[rune]*node
	cnt  int
}

type tireTree struct {
	root *node
}

func generateTireTree(strs ...string) *tireTree {
	t := &tireTree{root: &node{next: map[rune]*node{}}}
	for _, str := range strs {
		t.Insert(str)
	}
	return t
}

func generateTireTreeWithMap(dict map[string]int) *tireTree {
	t := &tireTree{root: &node{next: map[rune]*node{}}}
	for str, count := range dict {
		t.InsertWithTimes(str, count)
	}
	return t
}

func (t *tireTree) Insert(str string) {
	t.InsertWithTimes(str, 1)
}

// 插入str字符串x次
func (t *tireTree) InsertWithTimes(str string, x int) {
	cur := t.root
	for _, r := range str {
		v, ok := cur.next[r]
		if !ok {
			n := &node{next: map[rune]*node{}}
			cur.next[r] = n
			v = n
		}
		cur = v
	}
	cur.cnt += x
}

func (t *tireTree) Search(str string) bool {
	cur := t.root
	for _, r := range str {
		cur = cur.next[r]
		if cur == nil {
			return false
		}
	}
	return cur.cnt > 0
}

func (t *tireTree) SearchAndDec(str string) bool {
	cur := t.root
	for _, r := range str {
		cur = cur.next[r]
		if cur == nil {
			return false
		}
	}
	if cur.cnt > 0 {
		cur.cnt--
		return true
	}
	return false
}
