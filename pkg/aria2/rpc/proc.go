package rpc

import "sync"

type ResponseProcFn func(resp clientResponse) error

type ResponseProcessor struct {
	cbs map[uint64]ResponseProcFn
	mu  *sync.RWMutex
}

func NewResponseProcessor() *ResponseProcessor {
	return &ResponseProcessor{
		make(map[uint64]ResponseProcFn),
		&sync.RWMutex{},
	}
}

func (r *ResponseProcessor) Add(id uint64, fn ResponseProcFn) {
	r.mu.Lock()
	r.cbs[id] = fn
	r.mu.Unlock()
}

func (r *ResponseProcessor) remove(id uint64) {
	r.mu.Lock()
	delete(r.cbs, id)
	r.mu.Unlock()
}

// Process called by recv routine
func (r *ResponseProcessor) Process(resp clientResponse) error {
	id := *resp.Id
	r.mu.RLock()
	fn, ok := r.cbs[id]
	r.mu.RUnlock()
	if ok && fn != nil {
		defer r.remove(id)
		return fn(resp)
	}
	return nil
}
