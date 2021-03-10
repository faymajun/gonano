package core

type CircularList struct {
	infos []interface{}
	index int
	size  int
}

func NewCircularQueue(size int) *CircularList {
	return &CircularList{
		infos: make([]interface{}, size),
		index: 0,
		size:  size,
	}
}

func (q *CircularList) Clear() {
	for i := 0; i < q.size; i++ {
		q.infos[i] = nil
	}
	q.index = 0
}

func (q *CircularList) EnQueue(info interface{}) {
	if info == nil {
		return
	}
	q.infos[q.index] = info
	q.index = (q.index + 1) % q.size
}

func (q *CircularList) Len() int {
	return len(q.List())
}

func (q *CircularList) List() (res []interface{}) {
	for i := 0; i < q.size; i++ {
		if q.infos[i] != nil {
			res = append(res, q.infos[i])
		}
	}
	return
}

func (q *CircularList) GetInfos() (res []interface{}) {
	return q.infos
}
