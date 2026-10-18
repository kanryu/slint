ROOT := ../..
GO ?= go
RM := rm -f

UI_DIR := ui
GENERATED_DIR := generated
BUNDLE := $(GENERATED_DIR)/ui_bundle.slint

ifeq ($(OS),Windows_NT)
BUNDLER := $(ROOT)/bin/slint-bundler.exe
else
BUNDLER := $(ROOT)/bin/slint-bundler
endif

.PHONY: all bundle run clean

all: bundle

bundle: $(BUNDLE)

$(BUNDLER): $(ROOT)/cmd/slint-bundler/main.go
	mkdir -p "$(dir $@)"
	$(GO) build -o "$(BUNDLER)" $(ROOT)/cmd/slint-bundler

$(BUNDLE): $(wildcard $(UI_DIR)/*.slint) $(wildcard $(UI_DIR)/pages/*.slint) $(BUNDLER)
	"$(BUNDLER)" "$(UI_DIR)" -o "$(BUNDLE)"

run: bundle
	$(GO) run .

clean:
	$(RM) "$(BUNDLE)"
