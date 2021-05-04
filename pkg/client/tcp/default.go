package xtcp

import "context"

func NewDefaultReceiver(buffer int) (func(context.Context) ([]byte, error), func([]byte) []byte) {
	var ch = make(chan []byte, buffer)
	return func(ctx context.Context) ([]byte, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()

			case msg := <-ch:
				return msg, nil
			}
		},
		func(data []byte) []byte {
			ch <- data
			return nil
		}
}
