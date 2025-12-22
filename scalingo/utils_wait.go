package scalingo

import (
	"context"
	"errors"
	"time"
)

const defaultWaitInterval = 5 * time.Second

type waitOptions struct {
	timeout    time.Duration
	interval   time.Duration
	immediate  bool
	timeoutErr error
}

func waitUntil(ctx context.Context, opts waitOptions, check func() (bool, error)) error {
	if opts.interval == 0 {
		opts.interval = defaultWaitInterval
	} else if opts.interval < 0 {
		return errors.New("wait interval must be positive")
	}

	if opts.immediate {
		done, err := check()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}

	ticker := time.NewTicker(opts.interval)
	defer ticker.Stop()

	var timeout <-chan time.Time
	var timer *time.Timer
	if opts.timeout > 0 {
		timer = time.NewTimer(opts.timeout)
		defer timer.Stop()
		timeout = timer.C
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			if opts.timeoutErr != nil {
				return opts.timeoutErr
			}
			return errors.New("timed out waiting for condition")
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
