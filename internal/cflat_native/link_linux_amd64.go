//go:build linux && amd64

package cflatnative

// #cgo linux,amd64 LDFLAGS: -ldl -lpthread -lm -lrt -lutil -lfontconfig
import "C"
