//go:build !windows && (!linux || !amd64)

package slint

import "errors"

var errUnsupported = errors.New("slint: cgo-free SlintCFlat binding is not generated for this platform")

type Context struct{}
type Value struct{}
type Message struct {
	Msg    uint32
	WParam uintptr
	LParam int64
}
type CallbackEvent struct{ Name, StringArg string }

func New(string) (*Context, error)                      { return nil, errUnsupported }
func (*Context) Close()                                 {}
func (*Context) IsRunning() bool                        { return false }
func (*Context) SetStringProperty(string, string) error { return errUnsupported }
func LastErrorCode() uint32                             { return 0 }
func ClearError()                                       {}
