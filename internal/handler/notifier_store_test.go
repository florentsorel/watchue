package handler_test

import (
	"context"
	"sync"
	"testing"

	"github.com/florentsorel/watchue/internal/handler"
)

type stubNotifier struct{}

func (stubNotifier) Send(ctx context.Context, resourceName string, on bool) error { return nil }
func (stubNotifier) SendTest(ctx context.Context) error                           { return nil }

func TestNotifierStore_GetSetRoundtrip(t *testing.T) {
	store := handler.NewNotifierStore()
	if got := store.Get(); got != nil {
		t.Fatalf("Get() on a fresh store = %v, want nil", got)
	}

	n := stubNotifier{}
	store.Set(n)
	if got := store.Get(); got != n {
		t.Errorf("Get() = %v, want %v", got, n)
	}
}

func TestNotifierStore_ConcurrentGetSet(t *testing.T) {
	store := handler.NewNotifierStore()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			store.Set(stubNotifier{})
		}()
		go func() {
			defer wg.Done()
			store.Get()
		}()
	}
	wg.Wait()
}
