package tool

import "sync"

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		defer w.Done()
		cb()
	}()
}

func (w *WaitGroupWrapper) Wait() {
	w.WaitGroup.Wait()
}

func (w *WaitGroupWrapper) WaitFn(fn func()) {
	w.WaitGroup.Wait()
	fn()
}
