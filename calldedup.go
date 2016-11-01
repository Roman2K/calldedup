package calldedup

type Dedup struct {
	do         func() interface{}
	lock       chan bool
	waiters    chan chan interface{}
	BeforeWait func()
}

func New(do func() interface{}) (d Dedup) {
	lock := make(chan bool, 1)
	lock <- true
	return Dedup{
		do:      do,
		lock:    lock,
		waiters: make(chan chan interface{}),
	}
}

func (d Dedup) Do() (result interface{}) {
	for {
		select {
		case <-d.lock:
			result = d.do()
			for {
				select {
				case ch := <-d.waiters:
					ch <- result
				default:
					d.lock <- true
					return
				}
			}
		default:
			ch := make(chan interface{}, 1)
			if d.BeforeWait != nil {
				d.BeforeWait()
			}
			select {
			case d.waiters <- ch:
				return <-ch
			case <-d.lock:
				d.lock <- true
			}
		}
	}
}
