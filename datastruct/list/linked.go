package list

import "github.com/jujunwang/Mudis/lib/utils"

type LinkedList struct {
	first *node
	last  *node
	size  int
}

type node struct {
	val  interface{}
	prev *node
	next *node
}

func (list *LinkedList) Add(val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	n := &node{
		val: val,
	}
	if list.last == nil {

		list.first = n
		list.last = n
	} else {
		n.prev = list.last
		list.last.next = n
		list.last = n
	}
	list.size++
}

func (list *LinkedList) find(index int) (n *node) {
	if index < list.size/2 {
		n = list.first
		for i := 0; i < index; i++ {
			n = n.next
		}
	} else {
		n = list.last
		for i := list.size - 1; i > index; i-- {
			n = n.prev
		}
	}
	return n
}

func (list *LinkedList) Get(index int) (val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index >= list.size {
		panic("index out of bound")
	}
	return list.find(index).val
}

func (list *LinkedList) Set(index int, val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index > list.size {
		panic("index out of bound")
	}
	n := list.find(index)
	n.val = val
}

func (list *LinkedList) Insert(index int, val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index > list.size {
		panic("index out of bound")
	}

	if index == list.size {
		list.Add(val)
		return
	}

	pivot := list.find(index)
	n := &node{
		val:  val,
		prev: pivot.prev,
		next: pivot,
	}
	if pivot.prev == nil {
		list.first = n
	} else {
		pivot.prev.next = n
	}
	pivot.prev = n
	list.size++
}

func (list *LinkedList) removeNode(n *node) {
	if n.prev == nil {
		list.first = n.next
	} else {
		n.prev.next = n.next
	}
	if n.next == nil {
		list.last = n.prev
	} else {
		n.next.prev = n.prev
	}
	//gc处理
	n.prev = nil
	n.next = nil

	list.size--
}

func (list *LinkedList) Remove(index int) (val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index >= list.size {
		panic("index out of bound")
	}

	n := list.find(index)
	list.removeNode(n)
	return n.val
}

func (list *LinkedList) RemoveLast() (val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if list.last == nil {
		// empty list
		return nil
	}
	n := list.last
	list.removeNode(n)
	return n.val
}

func (list *LinkedList) RemoveAllByVal(val interface{}) int {
	if list == nil {
		panic("list is nil")
	}
	n := list.first
	removed := 0
	var nextNode *node
	for n != nil {
		nextNode = n.next
		if utils.Equals(n.val, val) {
			list.removeNode(n)
			removed++
		}
		n = nextNode
	}
	return removed
}

func (list *LinkedList) RemoveByVal(val interface{}, count int) int {
	if list == nil {
		panic("list is nil")
	}
	n := list.first
	removed := 0
	var nextNode *node
	for n != nil {
		nextNode = n.next
		if utils.Equals(n.val, val) {
			list.removeNode(n)
			removed++
		}
		if removed == count {
			break
		}
		n = nextNode
	}
	return removed
}

func (list *LinkedList) ReverseRemoveByVal(val interface{}, count int) int {
	if list == nil {
		panic("list is nil")
	}
	n := list.last
	removed := 0
	var prevNode *node
	for n != nil {
		prevNode = n.prev
		if utils.Equals(n.val, val) {
			list.removeNode(n)
			removed++
		}
		if removed == count {
			break
		}
		n = prevNode
	}
	return removed
}

func (list *LinkedList) Len() int {
	if list == nil {
		panic("list is nil")
	}
	return list.size
}

func (list *LinkedList) ForEach(consumer func(int, interface{}) bool) {
	if list == nil {
		panic("list is nil")
	}
	n := list.first
	i := 0
	for n != nil {
		goNext := consumer(i, n.val)
		if !goNext {
			break
		}
		i++
		n = n.next
	}
}

func (list *LinkedList) Contains(val interface{}) bool {
	contains := false
	list.ForEach(func(i int, actual interface{}) bool {
		if actual == val {
			contains = true
			return false
		}
		return true
	})
	return contains
}

func (list *LinkedList) Range(start int, stop int) []interface{} {
	if list == nil {
		panic("list is nil")
	}
	if start < 0 || start >= list.size {
		panic("`start` out of range")
	}
	if stop < start || stop > list.size {
		panic("`stop` out of range")
	}

	sliceSize := stop - start
	slice := make([]interface{}, sliceSize)
	n := list.first
	i := 0
	sliceIndex := 0
	for n != nil {
		if i >= start && i < stop {
			slice[sliceIndex] = n.val
			sliceIndex++
		} else if i >= stop {
			break
		}
		i++
		n = n.next
	}
	return slice
}

func Make(vals ...interface{}) *LinkedList {
	list := LinkedList{}
	for _, v := range vals {
		list.Add(v)
	}
	return &list
}
