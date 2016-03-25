package util

type empty struct{}
type Semaphore chan empty

func NewSemaphore(size int) Semaphore {
	return make(Semaphore, size)
}

func (s Semaphore) Lock() {
	s <- empty{}
}

func (s Semaphore) Unlock() {
	<-s
}

func (s Semaphore) Available() int {
	return len(s)
}

func (s Semaphore) Size() int {
	return cap(s)
}
