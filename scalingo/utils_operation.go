package scalingo

import (
	"context"
	"fmt"
	"time"

	"github.com/Scalingo/go-scalingo/v8"
)

func waitOperation(ctx context.Context, client *scalingo.Client, location string) error {
	var err error

	op := &scalingo.Operation{}
	return waitUntil(ctx, WaitOptions{
		Timeout:    5 * time.Minute,
		Interval:   10 * time.Second,
		Immediate:  true,
		TimeoutErr: fmt.Errorf("restart operation timeout"),
	}, func() (bool, error) {
		op, err = client.OperationsShowFromURL(ctx, location)
		if err != nil {
			return false, err
		}
		if op.Status == scalingo.OperationStatusDone {
			return true, nil
		}
		if op.Status == scalingo.OperationStatusError {
			return false, fmt.Errorf("restart operation failed %v", op.Error)
		}
		return false, nil
	})
}
