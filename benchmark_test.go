package ratelimiter_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"

	ratelimiter "github.com/ilam01/limits-go"
)

func BenchmarkGet(b *testing.B) {

	limiter := ratelimiter.New(ratelimiter.Options{})
	policy := []int{1000000, 1000}
	id := getUniqueID()

	b.N = 100000
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Get(context.Background(), id, policy...)
	}
}
func BenchmarkGetAndEexceeding(b *testing.B) {

	limiter := ratelimiter.New(ratelimiter.Options{})
	policy := []int{100, 1000}
	id := getUniqueID()

	b.N = 100000
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Get(context.Background(), id, policy...)
	}
}
func BenchmarkGetAndParallel(b *testing.B) {
	limiter := ratelimiter.New(ratelimiter.Options{})
	policy := []int{1000000, 1000}
	id := getUniqueID()

	b.N = 100000
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Get(context.Background(), id, policy...)
		}
	})
}
func BenchmarkGetAndClean(b *testing.B) {
	limiter := ratelimiter.New(ratelimiter.Options{})
	policy := []int{1000000, 1000}
	id := getUniqueID()

	b.N = 100000
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Get(context.Background(), id, policy...)
		}
		//limiter.Clean()
	})
}
func BenchmarkGetForDifferentUser(b *testing.B) {
	limiter := ratelimiter.New(ratelimiter.Options{})
	policy := []int{1, 10000}

	b.N = 10000
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := getUniqueID()
			limiter.Get(context.Background(), id, policy...)
		}
		//limiter.Clean()
	})
}

func getUniqueID() string {
	buf := make([]byte, 12)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
