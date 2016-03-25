package util

import (
	"fmt"
)

// Forwards messages from source channel to target channnel with prefix
// Target <- <Prefix> + Source
// Return func could be used to forward remaining messages to the output channel before closing it
// Example:
//   defer close(target)
//   source := someFunc()
//   defer util.ForwardInBackground("Prefix", source, target)()
func ForwardInBackground(prefix string, source <-chan string, target chan<- []string) func() {
	f := func() {
		for data := range source {
			prefixed := fmt.Sprintf("%s%s", prefix, data)
			target <- []string{prefixed}
		}
	}
	go func() {
		defer func() { recover() }()
		f()
	}()

	return f
}

// Concatenate two channels in async way, i.e., any message comming from a or b passes to result without any change
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
