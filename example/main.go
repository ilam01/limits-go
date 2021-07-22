// The ratelimiter-go HTTP Demo

package main

import (
	"context"
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

func (c *redisClient) RateDel(ctx context.Context, key string) error {
	return c.Del(ctx, key).Err()
}

func (c *redisClient) RateEvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return c.EvalSha(ctx, sha1, keys, args...).Result()
}

func (c *redisClient) RateScriptLoad(ctx context.Context, script string) (string, error) {
	return c.ScriptLoad(ctx, script).Result()
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
		Duration: time.Minute, // limit to 1000 requests in 1 minute.
		Client:   &redisClient{client},
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res, err := limiter.Get(r.Context(), r.URL.Path)
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
			_, _ = fmt.Fprintf(w, "Path: %q\n", html.EscapeString(r.URL.Path))
			_, _ = fmt.Fprintf(w, "Remaining: %d\n", res.Remaining)
			_, _ = fmt.Fprintf(w, "Total: %d\n", res.Total)
			_, _ = fmt.Fprintf(w, "Duration: %v\n", res.Duration)
			_, _ = fmt.Fprintf(w, "Reset: %v\n", res.Reset)
		} else {
			after := int64(res.Reset.Sub(time.Now())) / 1e9
			header.Set("Retry-After", strconv.FormatInt(after, 10))
			w.WriteHeader(429)
			_, _ = fmt.Fprintf(w, "Rate limit exceeded, retry in %d seconds.\n", after)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
