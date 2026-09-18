package scalingo

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/Scalingo/go-scalingo/v11"
)

func TestWaitUntilDatabaseProvisioned(t *testing.T) {
	for _, transientNotFound := range []bool{false, true} {
		t.Run(fmt.Sprintf("transient_not_found=%t", transientNotFound), func(t *testing.T) {
			// The create endpoint supplies the app ID in ID, but leaves App empty.
			created := scalingo.DatabaseNG{ID: "app-id", Name: "test-db"}
			client := NewMockDatabasesPreviewService(gomock.NewController(t))
			var notFound *gomock.Call
			if transientNotFound {
				notFound = client.EXPECT().DatabaseShow(gomock.Any(), created.ID).
					Return(scalingo.DatabaseNG{}, fmt.Errorf("search database: %w", scalingo.ErrDatabaseNotFound))
			}
			running := client.EXPECT().DatabaseShow(gomock.Any(), created.ID).
				Return(scalingo.DatabaseNG{ID: created.ID, Database: scalingo.Database{Status: scalingo.DatabaseStatusRunning}}, nil)
			if transientNotFound {
				gomock.InOrder(notFound, running)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 2*defaultWaitInterval)
			defer cancel()
			got, err := waitUntilDatabaseProvisioned(ctx, client, created)
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != created.ID || got.Database.Status != scalingo.DatabaseStatusRunning {
				t.Fatalf("unexpected database: %+v", got)
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
			client := NewMockDatabasesPreviewService(gomock.NewController(t))
			show := client.EXPECT().DatabaseShow(gomock.Any(), created.ID)
			if canceled {
				show.DoAndReturn(func(context.Context, string) (scalingo.DatabaseNG, error) {
					cancel()
					return scalingo.DatabaseNG{}, scalingo.ErrDatabaseNotFound
				})
			} else {
				show.Return(scalingo.DatabaseNG{}, wantErr)
			}
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
