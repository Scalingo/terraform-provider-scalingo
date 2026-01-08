package scalingo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Scalingo/go-scalingo/v9"
)

func waitOperation(ctx context.Context, client *scalingo.Client, location string) error {
	var err error

	op := &scalingo.Operation{}
	return waitUntil(ctx, waitOptions{
		timeout:    5 * time.Minute,
		timeoutErr: errors.New("restart operation timeout"),
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
