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

func NewFirstResultGroup[T any](ctx context.Context) *FirstResultGroup[T] {
	f := &FirstResultGroup[T]{}
	f.ctx, f.cancel = context.WithCancel(ctx)
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
				if b != nil {
					b(r)
				}
			}
		}
	}()
}

func (f *FirstResultGroup[T]) GetResult() (T, error) {
	f.doOne.Do(func() {
		f.wg.Wait()
		if f.has.CompareAndSwap(false, true) {
			f.err = fmt.Errorf("任务失败")
		}
	})
	return f.result, f.err
}
