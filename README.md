limits-go
==========
一个非常简单易用的分布式限流器，支持redis和memory两种模式

由teambition/ratelimiter-go完善而来

## 特性

- 分布式
- 原子性
- 高性能
- 支持Redis集群，已测试：阿里云、腾讯云（已测试）
- 支持内存存储
- 已支持Redis v8
- 支持令牌桶 （待更新）

## 使用

```sh
go get https://github.com/ilam01/limits-go
```
## 使用方法（Redis）
### 1、简单使用
```go
//1、建立一个客户端
limiter := ratelimiter.New(ratelimiter.Options{
    Client:   &redisClient{client}, //使用Redis时此项必须，否则会使用内存方式
})
//2、使用
var t := 1000 //毫秒（ms），时间片间隔
var limit := 10 //单个时间片允许的数量
res, err := limiter.Get(r.URL.Path,limit,t)
if res.Remaining >= 0 {
    //执行业务流程
} else {
    //已耗尽，处理失败逻辑
}
```

### 2、简单使用二
```go
//1、建立一个桶
limiter := ratelimiter.New(ratelimiter.Options{
    Max:      10,//单个时间片允许的数量
    Duration: time.Minute, // 一分钟只允许10个事件
    Client:   &redisClient{client}, //使用Redis时此项必须，否则会使用内存方式
})
//2、使用
res, err := limiter.Get(r.URL.Path)
if res.Remaining >= 0 {
    //执行业务流程
} else {
    //已耗尽，处理失败逻辑
}
```

### 3、内存使用一
```go
//1、建立一个客户端
limiter := ratelimiter.newMemoryLimiter(Options{})
//2、使用
var t := 1000 //毫秒（ms），时间片间隔
var limit := 10 //单个时间片允许的数量
res, err := limiter.Get(r.URL.Path,limit,t)
if res.Remaining >= 0 {
//执行业务流程
} else {
//已耗尽，处理失败逻辑
}
```

## HTTP实例
请尝试使用 `github.com/ilam01/limits-go` 目录下的:

```sh
go run example/main.go
```
访问: http://127.0.0.1:8080/

```go
package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"time"

	ratelimiter "github.com/ilam01/limits-go"
	"github.com/go-redis/redis/v8"
)

// Implements RedisClient for redis.Client
type redisClient struct {
	*redis.Client
}

func (c *redisClient) RateDel(key string) error {
	return c.Del(key).Err()
}
func (c *redisClient) RateEvalSha(sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return c.EvalSha(sha1, keys, args...).Result()
}
func (c *redisClient) RateScriptLoad(script string) (string, error) {
	return c.ScriptLoad(script).Result()
}

func main() {
	// use memory
	// limiter := ratelimiter.New(ratelimiter.Options{
	// 	Max:      10,
	// 	Duration: time.Minute, // limit to 1000 requests in 1 minute.
	// })

	// or use redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	limiter := ratelimiter.New(ratelimiter.Options{
		Max:      10,
		Duration: time.Minute, // limit to 10 requests in 1 minute.
		Client:   &redisClient{client},
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res, err := limiter.Get(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		header := w.Header()
		header.Set("X-Ratelimit-Limit", strconv.FormatInt(int64(res.Total), 10))
		header.Set("X-Ratelimit-Remaining", strconv.FormatInt(int64(res.Remaining), 10))
		header.Set("X-Ratelimit-Reset", strconv.FormatInt(res.Reset.Unix(), 10))

		if res.Remaining >= 0 {
			w.WriteHeader(200)
			fmt.Fprintf(w, "Path: %q\n", html.EscapeString(r.URL.Path))
			fmt.Fprintf(w, "Remaining: %d\n", res.Remaining)
			fmt.Fprintf(w, "Total: %d\n", res.Total)
			fmt.Fprintf(w, "Duration: %v\n", res.Duration)
			fmt.Fprintf(w, "Reset: %v\n", res.Reset)
		} else {
			after := int64(res.Reset.Sub(time.Now())) / 1e9
			header.Set("Retry-After", strconv.FormatInt(after, 10))
			w.WriteHeader(429)
			fmt.Fprintf(w, "Rate limit exceeded, retry in %d seconds.\n", after)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## License
`ratelimiter-go` is licensed under the [MIT](https://github.com/ilam01/limits-go/blob/master/LICENSE) license.
