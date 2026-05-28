package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://stackoverflow.com",
	}

	var wg sync.WaitGroup
	client := http.Client{Timeout: 5 * time.Second}
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			resp, err := client.Get(u)
			if err != nil {
				fmt.Printf("%s: 失败（%v）\n", u, err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("%s: 成功 (状态码: %d)\n", u, resp.StatusCode)
		}(url)
	}
	wg.Wait()
}
