package scalingo

import (
	"context"
	"fmt"
	"time"
)

type WaitOptions struct {
	Timeout    time.Duration
	Interval   time.Duration
	Immediate  bool
	TimeoutErr error
}

func waitUntil(ctx context.Context, opts WaitOptions, check func() (bool, error)) error {
	if opts.Interval <= 0 {
		return fmt.Errorf("wait interval must be positive")
	}

	if opts.Immediate {
		done, err := check()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	var timeout <-chan time.Time
	var timer *time.Timer
	if opts.Timeout > 0 {
		timer = time.NewTimer(opts.Timeout)
		defer timer.Stop()
		timeout = timer.C
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			if opts.TimeoutErr != nil {
				return opts.TimeoutErr
			}
			return fmt.Errorf("timed out waiting for condition")
		case <-ticker.C:
			done, err := check()
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}
