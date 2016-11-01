package calldedup

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestNoDedup(t *testing.T) {
	const doCount = 3

	call := 0
	d := New(func() interface{} {
		call++
		return call
	})

	results := make([]int, doCount)
	for i := 0; i < doCount; i++ {
		results[i] = d.Do().(int)
	}

	for idx, res := range results {
		want := idx + 1
		assertIntEqual(t, want, res, fmt.Sprintf("results[%d]", idx))
	}
}

func TestDedup(t *testing.T) {
	const (
		doDuration = 200 * time.Millisecond
		doersCount = 2
		doCount    = 3
	)

	call := 0
	d := New(func() interface{} {
		time.Sleep(doDuration)
		call++
		return call
	})

	for i := 0; i < doersCount; i++ {
		results := doResults(d, doCount)
		want := i + 1
		for idx, res := range results {
			assertIntEqual(t, want, res, fmt.Sprintf("results[%d]", idx))
		}
	}
}

func TestWaiterSlowerThanDo(t *testing.T) {
	const (
		doDuration = 200 * time.Millisecond
		doCount    = 2
	)

	call := 0
	d := New(func() interface{} {
		time.Sleep(doDuration)
		call++
		return call
	})
	d.BeforeWait = func() {
		time.Sleep(doDuration + doDuration/2)
	}

	results := doResults(d, doCount)
	sort.Ints(results)

	for idx, res := range results {
		want := idx + 1
		assertIntEqual(t, want, res, fmt.Sprintf("results[%d]", idx))
	}
}

func doResults(d Dedup, doCount int) (results []int) {
	results = make([]int, doCount)
	wg := sync.WaitGroup{}
	wg.Add(doCount)
	for i := 0; i < doCount; i++ {
		idx := i
		go func() {
			defer wg.Done()
			results[idx] = d.Do().(int)
		}()
	}
	wg.Wait()
	return
}

func assertIntEqual(t *testing.T, want, got int, name string) {
	if got != want {
		t.Fatalf("%s: got %d, want %d", name, got, want)
	}
}

func BenchmarkOverhead(b *testing.B) {
	const (
		minDoCount = 1
		maxDoCount = 3
	)

	do := func() interface{} {
		return nil
	}
	d := New(do)

	parallel := func(do func() interface{}, doCount int) {
		wg := sync.WaitGroup{}
		wg.Add(doCount)
		for i := 0; i < doCount; i++ {
			go func() {
				defer wg.Done()
				do()
			}()
		}
		wg.Wait()
	}

	for doCount := minDoCount; doCount <= maxDoCount; doCount++ {
		b.Run(fmt.Sprintf("doCount=%d", doCount), func(b *testing.B) {
			b.Run("with", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					parallel(d.Do, doCount)
				}
			})
			b.Run("without", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					parallel(do, doCount)
				}
			})
		})
	}
}
