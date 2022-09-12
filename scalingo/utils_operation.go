package scalingo

import (
	"fmt"
	"time"

	"github.com/Scalingo/go-scalingo/v5"
)

func waitOperation(client *scalingo.Client, location string) error {
	var err error

	op := &scalingo.Operation{}
	timer := time.NewTimer(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for op.Status != scalingo.OperationStatusDone {
		op, err = client.OperationsShowFromURL(location)
		if err != nil {
			return err
		}
		// Don't wait next tick
		if op.Status == scalingo.OperationStatusDone {
			break
		}
		if op.Status == scalingo.OperationStatusError {
			return fmt.Errorf("restart operation failed %v", op.Error)
		}
		select {
		case <-timer.C:
			return fmt.Errorf("restart operation timeout")
		case <-ticker.C:
			<-ticker.C
		}
	}
	return nil
}
