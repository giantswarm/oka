// Package service provides a simple way to manage and wait for goroutines.
package service

import "sync"

// wg is a WaitGroup used to wait for all running services to complete.
var wg sync.WaitGroup

// Run starts a new goroutine that executes the given function.
// It uses a WaitGroup to track the number of running services.
func Run(f func()) {
	wg.Add(1)
	go func(s func()) {
		defer wg.Done()
		s()
	}(f)
}

// Wait blocks until all running services have completed.
func Wait() {
	wg.Wait()
}
