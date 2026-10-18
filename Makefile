# SlintCFlat/Hike .syso build.
# The default `all` target follows the host OS. Linux additionally emits the
# ARM64 and ARM32 variants.

BRIDGE_DIR := internal/cflat
RM := rm -f
GENERATED_GO := $(BRIDGE_DIR)/cflat.go

GO         ?= go
CLANG      ?= clang
# hikepack must be installed (or supplied explicitly) because this module has
# its own go.mod; `go run ../cmd/hikepack` would resolve in the wrong module.
HIKEC      ?= hikec
HIKEPACK   ?= hikepack
BUNDLER_OUT ?= bin/slint-bundler$(if $(filter Windows_NT,$(OS)),.exe,)

SLINT_LIB_windows_amd64 ?= deps/slint-cflat-windows/libslint_cflat.a
SLINT_LIB_linux_amd64   ?= deps/slint-cflat-linux/libslint_cflat.a
SLINT_LIB_linux_arm64   ?= deps/slint-cflat-linux-arm64/libslint_cflat.a
SLINT_LIB_linux_arm32   ?= deps/slint-cflat-linux-arm32/libslint_cflat.a
SLINT_LIB_darwin_arm64  ?= deps/slint-cflat-darwin/libslint_cflat.a

ifeq ($(OS),Windows_NT)
HOST_GOOS := windows
ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
HOST_GOARCH := amd64
else ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
HOST_GOARCH := amd64
else ifeq ($(PROCESSOR_ARCHITECTURE),ARM64)
HOST_GOARCH := arm64
else
HOST_GOARCH := 386
endif
else
HOST_UNAME := $(shell uname -s)
HOST_MACHINE := $(shell uname -m)
ifeq ($(HOST_UNAME),Linux)
HOST_GOOS := linux
else ifeq ($(HOST_UNAME),Darwin)
HOST_GOOS := darwin
else
HOST_GOOS := $(HOST_UNAME)
endif
ifeq ($(HOST_MACHINE),x86_64)
HOST_GOARCH := amd64
else ifneq ($(filter aarch64 arm64,$(HOST_MACHINE)),)
HOST_GOARCH := arm64
else ifneq ($(filter armv7l armv7,$(HOST_MACHINE)),)
HOST_GOARCH := arm
else
HOST_GOARCH := $(HOST_MACHINE)
endif
endif

ifeq ($(HOST_GOOS),windows)
HOST_TARGET := windows-amd64
else ifeq ($(HOST_GOOS),linux)
HOST_TARGET := linux-amd64
else ifeq ($(HOST_GOOS),darwin)
HOST_TARGET := darwin-arm64
else
HOST_TARGET := $(HOST_GOOS)-$(HOST_GOARCH)
endif

.PHONY: all syso host slint-bundler linux syso-windows-amd64 syso-linux-amd64 \
	syso-linux-arm64 syso-linux-arm32 syso-darwin-arm64 clean

ifeq ($(HOST_GOOS),linux)
all: linux
else
all: host
endif

host: syso-$(HOST_TARGET)
linux: syso-linux-amd64 syso-linux-arm64 syso-linux-arm32
syso: host

slint-bundler:
	mkdir -p "$(dir $(BUNDLER_OUT))"
	$(GO) build -o "$(BUNDLER_OUT)" ./cmd/slint-bundler

define build_syso
	$(HIKEPACK) syso -v -target $(1) .
endef

syso-windows-amd64:
	$(call build_syso,windows,internal/cflat/cflat_hike_windows_amd64.syso,x86_64-w64-windows-gnu,$(SLINT_LIB_windows_amd64),internal/cflat/cflat_windows_amd64.syso)

syso-linux-amd64:
	$(call build_syso,linux,internal/cflat/cflat_hike_linux_amd64.syso,x86_64-unknown-linux-gnu,$(SLINT_LIB_linux_amd64),internal/cflat/cflat_linux_amd64.syso)

syso-linux-arm64:
	$(call build_syso,linux-arm64,internal/cflat/cflat_hike_linux_arm64.syso,aarch64-unknown-linux-gnu,$(SLINT_LIB_linux_arm64),internal/cflat/cflat_linux_arm64.syso)

syso-linux-arm32:
	$(call build_syso,linux-arm32,internal/cflat/cflat_hike_linux_arm32.syso,armv7-unknown-linux-gnueabihf,$(SLINT_LIB_linux_arm32),internal/cflat/cflat_linux_arm32.syso)

syso-darwin-arm64:
	$(call build_syso,darwin,internal/cflat/cflat_hike_darwin_arm64.syso,arm64-apple-darwin,$(SLINT_LIB_darwin_arm64),internal/cflat/cflat_darwin_arm64.syso)

clean:
	$(RM) internal/cflat/cflat_hike_*.syso internal/cflat/cflat_*_amd64.syso internal/cflat/cflat_*_archive_member*.syso internal/cflat/cflat_linux_arm*.syso internal/cflat/cflat_darwin_arm64.syso
