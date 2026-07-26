package util

type SNode struct {
	value interface{}
	pre   *SNode
	next  *SNode
}

type Stack struct {
	tail *SNode
	len  int
}

func (s *Stack) Len() int {
	return s.len
}

func (s *Stack) Peek() interface{} {
	if s.tail == nil {
		return nil
	}
	return s.tail.value
}

func (s *Stack) Push(value interface{}) {
	newNode := &SNode{value, nil, nil}
	if s.tail == nil {
		s.tail = newNode
	} else {
		newNode.pre = s.tail
		s.tail.next = newNode
		s.tail = newNode
	}
	s.len++
	newNode = nil
}

func (s *Stack) Pop() interface{} {
	if s.tail == nil {
		return nil
	}
	last := s.tail
	value := last.value

	s.tail = last.pre
	last.pre = nil
	last.next = nil
	last.value = nil
	s.len--
	last = nil

	return value
}

func (s *Stack) ToSlice() []interface{} {
	items := make([]interface{}, 0, s.len)
	for node := s.tail; node != nil; node = node.pre {
		items = append(items, node.value)
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items
}

// DeepCopy creates a deep copy of the stack, preserving order and content.
// Each value is copied by value (interface{} reference is copied, not the
// underlying data). If stack items require custom deep copy semantics, the
// caller must handle that separately.
func (s *Stack) DeepCopy() *Stack {
	if s == nil {
		return nil
	}
	newStack := &Stack{
		len: s.len,
	}
	if s.tail == nil {
		return newStack
	}

	// Collect all nodes from tail to head (reverse order)
	items := s.ToSlice()

	// Rebuild stack in original order
	for _, item := range items {
		newStack.Push(item)
	}

	return newStack
}
