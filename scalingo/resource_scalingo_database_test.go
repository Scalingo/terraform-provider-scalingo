package scalingo

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Scalingo/go-scalingo/v11"
)

type databasePreviewStub struct {
	scalingo.DatabasesPreviewService

	show func(context.Context, string) (scalingo.DatabaseNG, error)
}

func (s databasePreviewStub) DatabaseShow(ctx context.Context, id string) (scalingo.DatabaseNG, error) {
	return s.show(ctx, id)
}

func TestWaitUntilDatabaseProvisioned(t *testing.T) {
	for _, transientNotFound := range []bool{false, true} {
		t.Run(fmt.Sprintf("transient_not_found=%t", transientNotFound), func(t *testing.T) {
			// The create endpoint supplies the app ID in ID, but leaves App empty.
			created := scalingo.DatabaseNG{ID: "app-id", Name: "test-db"}
			calls := 0
			client := databasePreviewStub{show: func(_ context.Context, id string) (scalingo.DatabaseNG, error) {
				calls++
				if id != created.ID {
					t.Fatalf("lookup ID = %q, want %q", id, created.ID)
				}
				if transientNotFound && calls == 1 {
					return scalingo.DatabaseNG{}, fmt.Errorf("search database: %w", scalingo.ErrDatabaseNotFound)
				}
				return scalingo.DatabaseNG{ID: created.ID, Database: scalingo.Database{Status: scalingo.DatabaseStatusRunning}}, nil
			}}
			ctx, cancel := context.WithTimeout(t.Context(), 2*defaultWaitInterval)
			defer cancel()
			got, err := waitUntilDatabaseProvisioned(ctx, client, created)
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != created.ID || got.Database.Status != scalingo.DatabaseStatusRunning {
				t.Fatalf("unexpected database: %+v", got)
			}
			wantCalls := 1
			if transientNotFound {
				wantCalls = 2
			}
			if calls != wantCalls {
				t.Fatalf("calls = %d, want %d", calls, wantCalls)
			}
		})
	}
}

func TestWaitUntilDatabaseProvisionedPreservesResourceOnError(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(fmt.Sprintf("canceled=%t", canceled), func(t *testing.T) {
			created := scalingo.DatabaseNG{ID: "app-id", Name: "test-db"}
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			wantErr := errors.New("API unavailable")
			if canceled {
				wantErr = context.Canceled
			}
			client := databasePreviewStub{show: func(_ context.Context, id string) (scalingo.DatabaseNG, error) {
				if id != created.ID {
					t.Fatalf("lookup ID = %q, want %q", id, created.ID)
				}
				if canceled {
					cancel()
					return scalingo.DatabaseNG{}, scalingo.ErrDatabaseNotFound
				}
				return scalingo.DatabaseNG{}, wantErr
			}}
			got, err := waitUntilDatabaseProvisioned(ctx, client, created)
			if !errors.Is(err, wantErr) {
				t.Fatalf("error = %v, want %v", err, wantErr)
			}
			if got.ID != created.ID || got.Name != created.Name {
				t.Fatalf("lost resource on error: %+v", got)
			}
		})
	}
}
