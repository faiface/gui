package event

import (
	"strings"
	"sync"
)

const Sep = "/"

type Dispatch struct {
	mu       sync.Mutex
	handlers []func(event string)
	trie     map[string]*Dispatch
}

func (d *Dispatch) Event(pattern string, handler func(event string)) {
	if pattern == "" {
		d.event(nil, handler)
		return
	}
	d.event(strings.Split(pattern, Sep), handler)
}

func (d *Dispatch) Happen(event string) {
	if event == "" {
		d.happen(nil)
		return
	}
	d.happen(strings.Split(event, Sep))
}

func (d *Dispatch) event(pattern []string, handler func(event string)) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(pattern) == 0 {
		d.handlers = append(d.handlers, handler)
		return
	}

	if d.trie == nil {
		d.trie = make(map[string]*Dispatch)
	}
	if d.trie[pattern[0]] == nil {
		d.trie[pattern[0]] = &Dispatch{}
	}
	d.trie[pattern[0]].event(pattern[1:], handler)
}

func (d *Dispatch) happen(event []string) {
	d.mu.Lock()
	handlers := d.handlers
	d.mu.Unlock()

	for _, handler := range handlers {
		handler(strings.Join(event, Sep))
	}

	if len(event) == 0 {
		return
	}

	d.mu.Lock()
	next := d.trie[event[0]]
	d.mu.Unlock()

	if next != nil {
		next.happen(event[1:])
	}
}
