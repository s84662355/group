# 一个并发执行多个任务的工具，可以获取最快完成的任务结果




## 源码
```shell

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
				b(r)
			}
		}
	}()
}

func (f *FirstResultGroup[T]) GetResult() (T, error) {
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
		}
	})
	return f.result, f.err
}





```
 











### 使用例子，获取本机的公网ip，会返回最快执行成功的那个任务的结果
```shell
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/s84662355/group"
)

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	f := NewFirstResultGroup[string](ctx)

	f.Go(func(ctx context.Context) (string, bool) {
		data, err := GetIPFromIPify(ctx)
		return data, err == nil
	}, func(data string) {
	})

	f.Go(func(ctx context.Context) (string, bool) {
		data, err := GetIPFromIPInfo(ctx)
		return data, err == nil
	}, func(data string) {
	})

	f.Go(func(ctx context.Context) (string, bool) {
		data, err := GetIPFromIcanhazip(ctx)
		return data, err == nil
	}, func(data string) {
	})

	fmt.Println(f.GetResult())
}

// GetIPFromIPify 通过ipify.org获取公网IP
func GetIPFromIPify(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.ipify.org", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body) + "api.ipify.org", nil
}

// GetIPFromIPInfo 通过ipinfo.io获取公网IP
func GetIPFromIPInfo(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://ipinfo.io/ip", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body) + "ipinfo.io/ip", nil
}

// GetIPFromicanhazip 通过icanhazip.com获取公网IP
func GetIPFromIcanhazip(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://icanhazip.com", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// icanhazip返回的内容可能包含换行符，需要清理
	return string(body) + "icanhazip.com", nil
}
```
