package utils

type QNode struct {
	value interface{}
	next  *QNode
}

type Queue struct {
	head *QNode
	tail *QNode
	len  int
}

func (q *Queue) Len() int {
	return q.len
}

func (q *Queue) Peek() interface{} {
	if q.head == nil {
		return nil
	}
	return q.head.value
}

func (q *Queue) Enqueue(value interface{}) {
	newNode := &QNode{value, nil}
	if q.tail == nil {
		q.head = newNode
		q.tail = newNode
	} else {
		q.tail.next = newNode
		q.tail = newNode
	}
	q.len++
}

func (q *Queue) Dequeue() interface{} {
	if q.head == nil {
		return nil
	}
	first := q.head
	value := first.value

	q.head = first.next
	first.next = nil
	first.value = nil
	q.len--
	first = nil

	return value
}
