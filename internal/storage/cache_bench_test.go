package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func BenchmarkCacheSetWithBudget(b *testing.B) {
	for _, size := range []int{1000, 10000} {
		b.Run(fmt.Sprintf("entries=%d", size), func(b *testing.B) {
			ctx := context.Background()
			store, err := Open(ctx, filepath.Join(b.TempDir(), "cache.db"), "secret")
			if err != nil {
				b.Fatal(err)
			}
			defer store.Close()
			cache := store.Cache()
			for i := 0; i < size; i++ {
				if err := cache.Set(ctx, "bench", fmt.Sprint(i), []byte("payload"), "", 0); err != nil {
					b.Fatal(err)
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := cache.Set(ctx, "bench", "updated", []byte("payload"), "", 0); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
