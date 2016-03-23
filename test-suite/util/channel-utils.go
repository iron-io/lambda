package util

import (
	"fmt"
	"time"
)

// Forwards messages from source channel to target channnel with prefix
// Target <- <Prefix> + Source
// The last message is repeated to the target channel every minute if there is no message in the source
func Forward(prefix string, from <-chan string, to chan<- []string) {
	go func() {
		defer func() { recover() }()
		var lastData *string = nil
		for {
			select {
			case data, ok := <-from:
				{
					if !ok {
						return
					}
					prefixed := fmt.Sprintf("%s%s", prefix, data)
					lastData = &prefixed
					to <- []string{prefixed}
				}
			case <-time.After(time.Minute):
				if lastData != nil {
					to <- []string{*lastData}
				}
			}
		}
	}()
}

// Concatinate two channels in async way, i.e., any message comming from a or b passes to result without any change
func JoinChannels(a <-chan []string, b <-chan []string) <-chan []string {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	r := make(chan []string)
	go func() {
		defer close(r)
		for a != nil || b != nil {
			select {
			case item, ok := <-a:
				if ok {
					r <- item
				} else {
					a = nil
				}
			case item, ok := <-b:
				if ok {
					r <- item
				} else {
					b = nil
				}
			}
		}
	}()
	return r
}
