# 一个并发执行多个任务，获取最快完成的任务结果

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
	f := group.NewFirstResultGroup[string]()

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

	fmt.Println(f.GetResult(context.Background().Done()))
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
