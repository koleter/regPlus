package regexp

type node struct {
	Next map[rune]*node
	Cnt  int
}

func (t *node) Insert(strs ...string) {
	for _, str := range strs {
		t.InsertWithTimes(str, 1)
	}
}

// 插入str字符串x次
func (t *node) InsertWithTimes(str string, x int) {
	if t.Next == nil {
		t.Next = map[rune]*node{}
	}
	for _, r := range str {
		v, ok := t.Next[r]
		if !ok {
			n := &node{Next: map[rune]*node{}}
			t.Next[r] = n
			v = n
		}
		t = v
	}
	t.Cnt += x
}

func (t *node) Search(str string) bool {
	if t.Next == nil {
		return false
	}
	for _, r := range str {
		t = t.Next[r]
		if t == nil {
			return false
		}
	}
	return t.Cnt > 0
}

func (t *node) SearchAndDec(str string) bool {
	if t.Next == nil {
		return false
	}
	for _, r := range str {
		t = t.Next[r]
		if t == nil {
			return false
		}
	}
	if t.Cnt > 0 {
		t.Cnt--
		return true
	}
	return false
}
