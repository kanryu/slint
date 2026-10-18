//go:build windows && amd64

package cflatnative

// #cgo windows,amd64 LDFLAGS: -lgdi32 -lopengl32 -limm32 -lntdll -luser32 -lkernel32 -lole32 -loleaut32 -lcomctl32 -ldwmapi -ldwrite -lshlwapi -lshell32 -luxtheme -lpropsys -lmsimg32 -lusp10 -lfontsub -lurlmon -loleacc -luiautomationcore -ladvapi32 -lwinhttp -lws2_32 -lucrt -lucrtbase -lm -lgcc -lgcc_eh -lmingwex -lmingw32
import "C"

// Native system libraries are deferred to the final Go external link.
