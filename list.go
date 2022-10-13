package regexp

type Element struct {
	prev, next *Element
	Value      interface{}
	list       *list
}

func (e *Element) Prev() *Element {
	if e.prev != nil && e.list != nil && e.prev != &e.list.root {
		return e.prev
	}
	return nil
}

func (e *Element) Next() *Element {
	if e.next != nil && e.list != nil && e.next != &e.list.root {
		return e.next
	}
	return nil
}

func (e *Element) RemoveSelf() {
	if e.prev != nil && e.next != nil {
		e.prev.next = e.next
		e.next.prev = e.prev
		e.prev = nil
		e.next = nil
	}
}

type list struct {
	root Element
}

func GenList() *list {
	l := list{}
	l.root.prev = &l.root
	l.root.next = &l.root
	return &l
}

func (l *list) init() {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.root.list = l
}

func (l *list) Front() *Element {
	return l.root.next
}

func (l *list) Back() *Element {
	return l.root.prev
}

func (l *list) PushFront(val interface{}) *Element {
	if l.root.next == nil {
		l.init()
	}
	return l.InsertAfter(val, &l.root)
}

func (l *list) InsertAfter(val interface{}, root *Element) *Element {
	e := &Element{Value: val}
	root.InsertElementAfter(e)
	return e
}

// 将e插入到root之后
func (root *Element) InsertElementAfter(e *Element) {
	e.list = root.list
	if e.next != nil {
		e.next.prev = e.prev
	}
	if e.prev != nil {
		e.prev.next = e.next
	}
	e.prev = root
	e.next = root.next
	root.next = e
	e.next.prev = e
}

func (l *list) PushBack(val interface{}) *Element {
	if l.root.next == nil {
		l.init()
	}
	return l.InsertBefore(val, &l.root)
}

func (l *list) InsertBefore(val interface{}, root *Element) *Element {
	e := &Element{Value: val}
	root.InsertElementBefore(e)
	return e
}

func (root *Element) InsertElementBefore(e *Element) {
	e.list = root.list
	if e.next != nil {
		e.next.prev = e.prev
	}
	if e.prev != nil {
		e.prev.next = e.next
	}
	e.prev = root.prev
	e.next = root
	root.prev = e
	e.prev.next = e
}

func (root *Element) MoveElementBefore(e *Element) {
	if e.list != root.list {
		panic("move a invalid Element")
	}
	if e.next != nil {
		e.next.prev = e.prev
	}
	if e.prev != nil {
		e.prev.next = e.next
	}
	e.prev = root.prev
	e.next = root
	root.prev = e
	e.prev.next = e
}

func (l *list) Collection() []interface{} {
	ans := make([]interface{}, 0, 10)
	for node := l.Front(); node != nil; node = node.Next() {
		ans = append(ans, node.Value)
	}
	return ans
}
