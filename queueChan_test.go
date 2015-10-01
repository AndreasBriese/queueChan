package queueChan

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	//	"time"
)

var minQueueLen = 1000
var wg sync.WaitGroup

func TestQueueSimple(t *testing.T) {
	fmt.Println("Testing Push(), Pop(), PopPush(), auto grow, dynamic shrink")

	q := (&QueueChan{}).New(minQueueLen)
	pow2 := 1
	for {
		if pow2 > minQueueLen {
			break
		}
		pow2 <<= 2
	}
	if q.Capacity() != pow2 {
		t.Error("predefined length:", q.Capacity(), "instead of ", pow2)
	}

	q = (&QueueChan{}).New()

	for i := 0; i < minQueueLen; i++ {
		q.Push(i)
	}

	if q.Capacity() != pow2 {
		t.Error("auto grow:", q.Capacity(), "instead of ", pow2)
	}

	q.Dynamic()

	for i := 0; i < minQueueLen; i++ {
		q.Pop()
	}

	if q.Capacity() != 4 {
		t.Error("auto shrink:", q.Capacity(), "instead of ", pow2)
	}

	q = (&QueueChan{}).New()

	for i := 0; i < minQueueLen; i++ {
		q.Push(i)
	}

	if q.Length() != minQueueLen {
		t.Error("Pop:", q.Length(), "instead of ", minQueueLen)
	}

	for i := 0; i < minQueueLen; i++ {
		if e := q.PopPush().(int); e != i {
			t.Error("PopPush", i, "had value", e)
		}
	}

	if q.Length() != minQueueLen {
		t.Error("PopPush:", q.Length(), "instead of ", minQueueLen)
	}

	for i := 0; i < minQueueLen; i++ {
		if e := q.Pop().(int); e != i {
			t.Error("Pop", i, "had value", q.Pop())
		}
	}
}

func TestQueueSimpleTS(t *testing.T) {
	fmt.Println("Testing threadsafe: PushTS(), PopTS(), PopPushTS(), auto grow, dynamic shrink")

	q := (&QueueChan{}).New(minQueueLen)
	pow2 := 1
	for {
		if pow2 > minQueueLen {
			break
		}
		pow2 <<= 2
	}
	if q.Capacity() != pow2 {
		t.Error("predefined length:", q.Capacity(), "instead of ", pow2)
	}

	q = (&QueueChan{}).New()

	wg.Add(minQueueLen)
	for i := 0; i < minQueueLen; i++ {
		go func(i int) {
			q.PushTS(i)
			wg.Done()
		}(i)
	}
	wg.Wait()

	if q.Capacity() != pow2 {
		t.Error("auto grow:", q.Capacity(), "instead of ", pow2)
	}

	q.Dynamic()

	wg.Add(minQueueLen)
	for i := 0; i < minQueueLen; i++ {
		go func(i int) {
			q.PopTS()
			wg.Done()
		}(i)
	}
	wg.Wait()

	if q.Capacity() != 4 {
		t.Error("auto shrink:", q.Capacity(), "instead of ", pow2)
	}

	q = (&QueueChan{}).New()

	wg.Add(minQueueLen)
	sum := int64(0)
	for i := 0; i < minQueueLen; i++ {
		go func(i int) {
			q.PushTS(i)
			atomic.AddInt64(&sum, int64(i))
			wg.Done()
		}(i)
	}
	wg.Wait()

	if q.Length() != minQueueLen {
		t.Error("Pop:", q.Length(), "instead of ", minQueueLen)
	}

	wg.Add(minQueueLen)
	sumPop := int64(0)
	for i := 0; i < minQueueLen; i++ {
		go func() {
			e := q.PopPushTS().(int)
			atomic.AddInt64(&sumPop, int64(e))
			wg.Done()
		}()
	}
	wg.Wait()

	if sum != sumPop {
		t.Error("sum PopPushTS", sumPop, "instead of ", sum)
	}

	if q.Length() != minQueueLen {
		t.Error("PopPushTS Length:", q.Length(), "instead of ", minQueueLen)
	}

	wg.Add(minQueueLen)
	sumPop = 0
	for i := 0; i < minQueueLen; i++ {
		go func() {
			e := q.PopTS().(int)
			atomic.AddInt64(&sumPop, int64(e))
			wg.Done()
		}()
	}
	wg.Wait()

	if sum != sumPop {
		t.Error("sum PopTS", sumPop, "instead of ", sum)
	}
}

func TestQueuePopChan(t *testing.T) {
	fmt.Println("Testing: PopChan()")

	q := (&QueueChan{}).New(minQueueLen)

	wg.Add(minQueueLen)
	sum := int64(0)
	for i := 0; i < minQueueLen; i++ {
		go func(i int) {
			q.PushTS(i)
			atomic.AddInt64(&sum, int64(i))
			wg.Done()
		}(i)
	}
	wg.Wait()

	sumPop := 0
loop:
	for {
		select {
		case e := <-q.PopChan():
			sumPop += e.(int)
		case <-q.Empty:
			//  ended garceful: queue is empty
			break loop
		}
	}

	if sum != int64(sumPop) {
		t.Error("sum PopTS", sumPop, "instead of ", sum)
	}
}

func TestQueuePopChanPush(t *testing.T) {
	fmt.Println("Testing: PopChanPush()")

	q := (&QueueChan{}).New(minQueueLen)

	sum := 0
	for i := 0; i < minQueueLen; i++ {
		q.Push(i)
		sum += i
	}
	for i := 0; i < minQueueLen>>1; i++ {
		sum += i
	}

	sumPop := 0
loop:
	for i := 0; i < (minQueueLen + minQueueLen>>1); i++ {
		select {
		case e := <-q.PopChanPush():
			sumPop += e.(int)
		case <-q.Empty:
			//  ended garceful: queue is empty
			break loop
		}
	}

	if sum != sumPop {
		t.Error("sum PopTS", sumPop, "instead of ", sum)
	}
}

func TestQueuePopChanTS(t *testing.T) {
	fmt.Println("Testing: PopChanPushTS()")

	q := (&QueueChan{}).New(minQueueLen)

	sum := 0
	for i := 0; i < minQueueLen; i++ {
		q.Push(i)
		sum += i
	}
	for i := 0; i < minQueueLen>>1; i++ {
		sum += i
	}

	sumPop := int64(0)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < (minQueueLen + minQueueLen>>1); i++ {
			select {
			case e := <-q.PopChanPushTS():
				atomic.AddInt64(&sumPop, int64(e.(int)))
			case <-q.Empty:
				//  ended garceful: queue is empty
				return
			}
		}
	}()
	wg.Wait()
	if int64(sum) != sumPop {
		t.Error("sum PopTS", sumPop, "instead of ", sum)
	}
}

// BENCHMARKING
// go test -bench "Bench*" -count 5
// on my PowerBookPro first run of concurrent
// triggers b.N testing and multicore use checks  -> first run is lame!

func BenchmarkQueue_serial_PushPop(b *testing.B) {
	q := (&QueueChan{}).New()
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
	for i := 0; i < b.N; i++ {
		q.Pop()
	}
}

func BenchmarkQueue_serial_PushPopPush(b *testing.B) {
	q := (&QueueChan{}).New()
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
	for i := 0; i < b.N; i++ {
		q.PopPush()
	}
}

func BenchmarkQueue_serial_PushPop_withLength(b *testing.B) {
	q := (&QueueChan{}).New(b.N)
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
	for i := 0; i < b.N; i++ {
		q.Pop()
	}
}

func BenchmarkQueue_serial_PushPopPush_withLength(b *testing.B) {
	q := (&QueueChan{}).New(b.N)
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
	for i := 0; i < b.N; i++ {
		q.PopPush()
	}
}

func BenchmarkQueue_serial_PopChan(b *testing.B) {
	q := (&QueueChan{}).New(b.N)

	for i := 0; i < b.N; i++ {
		q.Push(i)
	}

	wg.Add(1)
	go func() {
		n := 0
		defer func() {
			//			fmt.Printf("b.N == n ? -> %v == %v -> %v\n", b.N, n, b.N == n)
			wg.Done()
		}()
		for {
			select {
			case <-q.PopChan():
				n++

			case <-q.Empty:
				//  ended graceful: queue is empty
				return
			}

		}

	}()

	wg.Wait()

}

func BenchmarkQueue_serial_PopChanPush(b *testing.B) {
	q := (&QueueChan{}).New(b.N)

	for i := 0; i < b.N; i++ {
		q.Push(i)
	}

	wg.Add(1)
	go func() {
		n := 0
		defer func() {
			//			fmt.Printf("b.N == n ? -> %v == %v -> %v\n", b.N, n, b.N == n)
			wg.Done()
		}()
		for {
			select {
			case <-q.PopChanPush():
				q.Pop()
				n++

			case <-q.Empty:
				//  ended graceful: queue is empty
				return
			}

		}

	}()

	wg.Wait()

}

func BenchmarkQueue_concurr_PushPopTS(b *testing.B) {
	q := (&QueueChan{}).New()
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			q.PushTS(i)
			wg.Done()
		}(i)
	}
	wg.Wait()

	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			q.PopTS()
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkQueue_concurr_PushPopPushTS(b *testing.B) {
	q := (&QueueChan{}).New()
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			q.PushTS(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			q.PopPushTS()
			wg.Done()
		}()
	}
	wg.Wait()
}
func BenchmarkQueue_concurr_PushPopTS_withLength(b *testing.B) {
	q := (&QueueChan{}).New(b.N)
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			q.PushTS(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			q.PopTS()
			wg.Done()
		}()
	}
	wg.Wait()
}
func BenchmarkQueue_concurr_PushPopPushTS_withLength(b *testing.B) {
	q := (&QueueChan{}).New(b.N)
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			q.PushTS(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			q.PopPushTS()
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkQueue_concurr_PopChanTS(b *testing.B) {
	q := (&QueueChan{}).New(b.N)

	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			q.PushTS(i)
			wg.Done()
		}(i)
	}
	wg.Wait()

	wg.Add(1)
	go func() {
		n := 0
		defer func() {
			//			fmt.Printf("b.N == n ? -> %v == %v -> %v\n", b.N, n, b.N == n)
			wg.Done()
		}()
		for {
			select {
			case <-q.PopChanTS():
				n++

			case <-q.Empty:
				//  ended graceful: queue is empty
				return
			}

		}

	}()

	wg.Wait()

}

func BenchmarkQueue_concurr_PopChanPushTS(b *testing.B) {
	q := (&QueueChan{}).New(b.N)

	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			q.PushTS(i)
			wg.Done()
		}(i)
	}
	wg.Wait()

	wg.Add(1)
	go func() {
		n := 0
		defer func() {
			//			fmt.Printf("b.N == n ? -> %v == %v -> %v\n", b.N, n, b.N == n)
			wg.Done()
		}()
		for {
			select {
			case <-q.PopChanPushTS():
				q.PopTS()
				n++

			case <-q.Empty:
				//  ended graceful: queue is empty
				return
			}

		}

	}()

	wg.Wait()

}
