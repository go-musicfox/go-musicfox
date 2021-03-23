package utils

type Node struct {
    value interface{}
    next  *Node
}

type Queue struct {
    head *Node
    tail *Node
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

func (q *Queue) enqueue(value interface{}) {
    newNode := &Node{value, nil}
    if q.tail == nil {
        q.head = newNode
        q.tail = newNode
    } else {
        q.tail.next = newNode
        q.tail = newNode
    }
    q.len++
    newNode = nil
}

func (q *Queue) dequeue() interface{} {
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