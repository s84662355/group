package group
 

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	
)

 

type (
	A[T any] func(ctx context.Context) (T, bool)
	B[T any] func(t T)
)

type FirstResultGroup[T any] struct {
	has    atomic.Bool
	c      atomic.Int32
	result T
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewFirstResultGroup[T any]() *FirstResultGroup[T]{
   f:=&FirstResultGroup[T]{}
   f.ctx,f.cancel = context. WithCancel(context.Background())  

   return f
}

func (f *FirstResultGroup[T]) Go(a A[T], b B[T]) {
	f.c.Add(1)
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		r, ok := a(f.ctx)
		if ok {
			if f.has.CompareAndSwap(false, true) {
				f.result = r
				f.cancel()
			} else {
				b(r)
			}
		}
		if f.c.Add(-1) == 0 {
			f.cancel()
		}
	}()
}

func (f *FirstResultGroup[T]) GetResult(done <-chan struct{}) (T, error) {
	defer f.wg.Wait()
	select {
	case <-f.ctx.Done():
		if !f.has.CompareAndSwap(false, true) {
			return f.result, nil
		} else {
			return f.result, fmt.Errorf("任务失败")
		}

	case <-done:
		f.cancel()
		if f.has.CompareAndSwap(false, true) {
			return f.result, fmt.Errorf("自定义上下结束")
		}

	}

	<-f.ctx.Done()
	return f.result, nil
}

 
 
 