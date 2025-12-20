BPF_SOURCES := $(wildcard bpf/*.bpf.c)
BPF_OBJECTS := $(patsubst bpf/%.bpf.c,bpf/%.bpf.o,$(BPF_SOURCES))

CLANG ?= clang
CFLAGS := -O2 -g -target bpf -D__TARGET_ARCH_x86_64

.PHONY: all generate build clean

# Compile BPF programs
bpf/%.bpf.o: bpf/%.bpf.c
	$(CLANG) $(CFLAGS) -c $< -o $@ -I./bpf/headers

# Generate Go bindings from compiled .o files
generate: $(BPF_OBJECTS)
	go generate ./internal/kernel/...

build: generate
	go build -o bin/cli ./cmd/cli.go

clean:
	rm -f bpf/*.o
	rm -f internal/kernel/*_bpfe*.go internal/kernel/*_bpfe*.o
	rm -rf bin/

all: clean build