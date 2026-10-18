# Minimal Slint example.
#
# Keep native link settings here instead of relying on directives in the
# generated binding package.

GO ?= go
RM ?= rm -f
ROOT := ../..

ifeq ($(OS),Windows_NT)
NATIVE_TARGET := syso-windows-amd64
GO_LDFLAGS := -linkmode=external -extldflags=-Wl,-u,slint_c_create,-lntdll,-lopengl32,-limm32,-lws2_32,-lgdi32,-luser32,-lkernel32
EXE := hello.exe
else
GO_LDFLAGS :=
EXE := hello
ifeq ($(shell uname -s),Linux)
ifeq ($(shell uname -m),x86_64)
NATIVE_TARGET := syso-linux-amd64
GO_LDFLAGS := -linkmode=external -extldflags=-Wl,-u,slint_c_create
endif
endif
endif

.PHONY: all native build run clean

all: build

native:
ifneq ($(NATIVE_TARGET),)
	$(MAKE) -C "$(ROOT)" $(NATIVE_TARGET)
endif

build: native
	$(GO) build -ldflags "$(GO_LDFLAGS)" -o "$(EXE)" .

run: build
	./$(EXE)

clean:
	$(RM) "$(EXE)"
