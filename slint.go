//go:build windows || (linux && amd64)

// Package slint provides a cgo-free Go binding for SlintCFlat.
package slint

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/kanryu/slint/internal/cflat"
)

type Context struct{ ptr *byte }
type Value struct{ ptr *byte }
type Message struct {
	Msg    uint32
	WParam uintptr
	LParam int64
}
type CallbackEvent struct{ Name, StringArg string }

const (
	WMNull       = 0x0000
	WMCreate     = 0x0001
	WMDestroy    = 0x0002
	WMClose      = 0x0010
	WMQuit       = 0x0012
	WMCommand    = 0x0111
	WMUser       = 0x0400
	ArrayInteger = 1
	ArrayFloat   = 2
	ArrayString  = 3
	ArrayImage   = 4
	ArrayBool    = 5
)

func New(code string) (*Context, error) {
	p := cflat.Create(code)
	if p == nil {
		return nil, lastError()
	}
	c := &Context{ptr: p}
	runtime.SetFinalizer(c, (*Context).Close)
	return c, nil
}
func (c *Context) Close() {
	if c != nil && c.ptr != nil {
		// slint_c_free waits for the native UI thread to finish.  Ask the
		// window/event loop to terminate first; otherwise the join can wait
		// forever while the UI thread is still pumping messages.
		_ = cflat.PostMessage(c.ptr, WMQuit, 0, 0)
		cflat.Free(c.ptr)
		c.ptr = nil
		runtime.SetFinalizer(c, nil)
	}
}
func (c *Context) IsRunning() bool { return c != nil && c.ptr != nil && cflat.IsRunning(c.ptr) }
func (c *Context) OpenFileDialog(capacity int) (string, bool, error) {
	if c == nil || c.ptr == nil {
		return "", false, errors.New("slint: closed context")
	}
	if capacity < 2 {
		capacity = 4096
	}
	buf := make([]byte, capacity)
	if !cflat.OpenFileDialog(c.ptr, &buf[0], capacity) {
		return "", false, lastError()
	}
	return cStringValue(buf), true, nil
}
func (c *Context) PostMessage(msg uint32, wparam uintptr, lparam int64) error {
	if !cflat.PostMessage(c.ptr, int(msg), int(wparam), int(lparam)) {
		return lastError()
	}
	return nil
}
func (c *Context) SendMessage(msg uint32, wparam uintptr, lparam int64) (int64, error) {
	return int64(cflat.SendMessage(c.ptr, int(msg), int(wparam), int(lparam))), nil
}
func (c *Context) PollEvent() (Message, bool, error) {
	var event struct {
		Msg    uint32
		WParam uintptr
		LParam int64
	}
	if !cflat.PollEvent(c.ptr, (*byte)(unsafe.Pointer(&event))) {
		return Message{}, false, nil
	}
	return Message{event.Msg, event.WParam, event.LParam}, true, nil
}
func (c *Context) DispatchEvent(m Message) error {
	event := struct {
		Msg    uint32
		WParam uintptr
		LParam int64
	}{m.Msg, m.WParam, m.LParam}
	if !cflat.DispatchEvent(c.ptr, (*byte)(unsafe.Pointer(&event))) {
		return lastError()
	}
	return nil
}
func (c *Context) PollCallback() (CallbackEvent, bool, error) {
	var event struct {
		Name      *byte
		StringArg *byte
	}
	if !cflat.PollCallback(c.ptr, (*byte)(unsafe.Pointer(&event))) {
		return CallbackEvent{}, false, nil
	}
	return CallbackEvent{cStringPointer(event.Name), cStringPointer(event.StringArg)}, true, nil
}

func (c *Context) SetStringProperty(name, value string) error {
	v := cString(value)
	if !cflat.SetProperty(c.ptr, name, &v[0]) {
		return lastError()
	}
	return nil
}
func (c *Context) SetBoolProperty(name string, value bool) error {
	return c.setScalar(name, unsafe.Pointer(&value))
}
func (c *Context) SetNumberProperty(name string, value float64) error {
	return c.setScalar(name, unsafe.Pointer(&value))
}
func (c *Context) SetGlobalStringProperty(g, n, v string) error {
	d := cString(v)
	if !cflat.SetGlobalProperty(c.ptr, g, n, &d[0]) {
		return lastError()
	}
	return nil
}
func (c *Context) SetGlobalBoolProperty(g, n string, v bool) error {
	return c.setGlobalScalar(g, n, unsafe.Pointer(&v))
}
func (c *Context) SetGlobalNumberProperty(g, n string, v float64) error {
	return c.setGlobalScalar(g, n, unsafe.Pointer(&v))
}
func (c *Context) setScalar(name string, p unsafe.Pointer) error {
	if !cflat.SetProperty(c.ptr, name, (*byte)(p)) {
		return lastError()
	}
	return nil
}
func (c *Context) setGlobalScalar(g, n string, p unsafe.Pointer) error {
	if !cflat.SetGlobalProperty(c.ptr, g, n, (*byte)(p)) {
		return lastError()
	}
	return nil
}
func (c *Context) SetPropertyValue(name string, v *Value) error {
	if !cflat.SetPropertyValue(c.ptr, name, v.ptr) {
		return lastError()
	}
	return nil
}
func (c *Context) SetGlobalPropertyValue(g, n string, v *Value) error {
	if !cflat.SetGlobalPropertyValue(c.ptr, g, n, v.ptr) {
		return lastError()
	}
	return nil
}

func NewString(s string) (*Value, error) {
	return newValue(cflat.ValueNewString(s))
}
func NewBool(v bool) (*Value, error)      { return newValue(cflat.ValueNewBool(v)) }
func NewNumber(v float64) (*Value, error) { return newValue(cflat.ValueNewNumber(v)) }
func NewInteger(v int64) (*Value, error)  { return newValue(cflat.ValueNewInteger(int(v))) }
func NewFloat(v float32) (*Value, error)  { return newValue(cflat.ValueNewFloat(v)) }
func NewImageFromFile(s string) (*Value, error) {
	return newValue(cflat.ImageLoadFromFile(s))
}
func NewArray(t uint32) (*Value, error) { return newValue(cflat.ValueNewArray(int(t))) }
func (v *Value) ArrayAppend(t uint32, e *Value) error {
	if !cflat.ValueArrayAppend(v.ptr, int(t), e.ptr) {
		return lastError()
	}
	return nil
}
func (v *Value) Close() {
	if v != nil && v.ptr != nil {
		cflat.ValueFree(v.ptr)
		v.ptr = nil
		runtime.SetFinalizer(v, nil)
	}
}
func newValue(p *byte) (*Value, error) {
	if p == nil {
		return nil, lastError()
	}
	v := &Value{p}
	runtime.SetFinalizer(v, (*Value).Close)
	return v, nil
}
func LastErrorCode() uint32 { return uint32(cflat.LastErrorCode()) }
func ClearError()           { cflat.ClearError() }
func lastError() error {
	message := cflat.LastError()
	if message == "" {
		return errors.New("slint: native operation failed")
	}
	return errors.New(message)
}
func cString(s string) []byte { return append([]byte(s), 0) }
func cStringValue(b []byte) string {
	for i, v := range b {
		if v == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
func cStringPointer(p *byte) string {
	if p == nil {
		return ""
	}
	q := unsafe.Slice(p, 1)
	for len(q) < 1<<20 && q[len(q)-1] != 0 {
		q = unsafe.Slice(p, len(q)+1)
	}
	return cStringValue(q)
}
