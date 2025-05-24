package group

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type (
	A[T any] func(ctx context.Context) (T, bool) ///任务函数
	B[T any] func(t T)                           // 当任务执行成功，但是结果没有被采纳就会执行
)

type FirstResultGroup[T any] struct {
	has    atomic.Bool
	result T
	err    error
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	doOne  sync.Once
}

func NewFirstResultGroup[T any]() *FirstResultGroup[T] {
	f := &FirstResultGroup[T]{}
	f.ctx, f.cancel = context.WithCancel(context.Background())

	return f
}

func (f *FirstResultGroup[T]) Go(a A[T], b B[T]) {
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
	}()
}

func (f *FirstResultGroup[T]) GetResult(done <-chan struct{}) (T, error) {
	f.doOne.Do(func() {
		d := make(chan struct{})
		defer func() {
			for range d {
				/* code */
			}
		}()
		go func() {
			defer close(d)
			f.wg.Wait()
			f.cancel()
		}()
		select {
		case <-f.ctx.Done():
			if !f.has.CompareAndSwap(false, true) {
				return
			} else {
				f.err = fmt.Errorf("任务失败")
				return
			}

		case <-done:
			f.cancel()
			if f.has.CompareAndSwap(false, true) {
				f.err = fmt.Errorf("自定义上下结束")
				return
			}

		}
	})
	<-f.ctx.Done()
	return f.result, f.err
}
