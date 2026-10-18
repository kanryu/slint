# Minimal Slint example.
#
# The SlintCFlat archive is supplied as a .syso file, so Go uses its
# external linker on Windows.  Keep the native libraries here instead of
# relying on cgo directives in the generated binding package.

GO ?= go
RM ?= rm -f

ifeq ($(OS),Windows_NT)
GO_LDFLAGS := -linkmode=external -extldflags=-Wl,-u,slint_c_create,-lntdll,-lopengl32,-limm32,-lws2_32,-lgdi32,-luser32,-lkernel32
EXE := hello.exe
else
GO_LDFLAGS :=
EXE := hello
endif

.PHONY: all build run clean

all: build

build:
	$(GO) build -ldflags "$(GO_LDFLAGS)" -o "$(EXE)" .

run: build
	./$(EXE)

clean:
	$(RM) "$(EXE)"
