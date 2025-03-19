package lib

type Semaphore struct {
	tickets chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		tickets: make(chan struct{}, n),
	}
}

func (s *Semaphore) Acquire() {
	s.tickets <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.tickets
}
