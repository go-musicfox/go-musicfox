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
