package authobserve

import (
	"context"
	"io"
	"sync"
)

type logoutBody struct {
	next        io.ReadCloser
	ctx         context.Context
	c           *Collector
	index, slot int
	mu          sync.Mutex
	data        []byte
	read        string
	closed      bool
}

// Called with collector lock held. Neither reservation nor copying reads any
// additional bytes from the underlying response.
func (c *Collector) wrapLogout(ctx context.Context, body io.ReadCloser, index int) io.ReadCloser {
	if c.data.Logout == nil || index < 0 || c.data.State != "complete" {
		return body
	}
	o := c.data.Logout
	if o.State == "budget_exceeded" {
		return body
	}
	if len(o.Responses) == 2 {
		o.State = "budget_exceeded"
		o.Responses = []LogoutResponse{}
		return body
	}
	slot := len(o.Responses)
	o.State = "observed"
	o.Responses = append(o.Responses, unknownLogout(index, "incomplete"))
	c.pending++
	b := &logoutBody{next: body, ctx: ctx, c: c, index: index, slot: slot, read: "incomplete"}
	c.logoutBodies = append(c.logoutBodies, b)
	return b
}
func (b *logoutBody) discard() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.read = "budget_exceeded"
	clear(b.data)
	b.data = nil
}
func (b *logoutBody) Read(p []byte) (int, error) {
	n, e := b.next.Read(p)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return n, e
	}
	if b.ctx.Err() != nil {
		b.read = "canceled"
		clear(b.data)
		b.data = nil
	}
	if b.read == "incomplete" || b.read == "complete" {
		if n < 0 || n > len(p) {
			b.read = "read_error"
		} else if len(b.data)+n > logoutLimit {
			b.read = "too_large"
		} else {
			b.c.mu.Lock()
			if b.c.data.Logout.State == "budget_exceeded" || b.c.logoutBytes+n > 2*logoutLimit {
				b.c.data.Logout.State = "budget_exceeded"
				b.c.data.Logout.Responses = []LogoutResponse{}
				b.read = "budget_exceeded"
			} else {
				b.c.logoutBytes += n
			}
			b.c.mu.Unlock()
			if b.read != "budget_exceeded" {
				b.data = append(b.data, p[:n]...)
			}
		}
		if b.read == "incomplete" || b.read == "complete" {
			if e == io.EOF {
				b.read = "complete"
			} else if e != nil {
				b.read = "read_error"
			} else {
				b.read = "incomplete"
			}
		}
	}
	if b.read != "incomplete" && b.read != "complete" {
		clear(b.data)
		b.data = nil
	}
	return n, e
}
func (b *logoutBody) Close() error {
	e := b.next.Close()
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return e
	}
	b.closed = true
	data, read := b.data, b.read
	b.data = nil
	b.mu.Unlock()
	defer clear(data)
	if b.ctx.Err() != nil {
		read = "canceled"
	} else if e != nil && read == "complete" {
		read = "read_error"
	}
	r := unknownLogout(b.index, read)
	if read == "complete" {
		r = analyzeLogout(b.ctx, data, b.index)
	}
	b.c.mu.Lock()
	defer b.c.mu.Unlock()
	b.c.pending--
	if b.c.data.State == "complete" && b.c.data.Logout.State == "observed" && b.slot < len(b.c.data.Logout.Responses) {
		b.c.data.Logout.Responses[b.slot] = r
	}
	return e
}
