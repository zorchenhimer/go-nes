
EXT=
ifeq ($(OS),Windows_NT)
EXT=.exe
endif

UTILS := chrutil
EXES := $(addsuffix .exe,$(addprefix bin/,$(UTILS)))
SRCS := $(addsuffix .go,$(addprefix cmd/,$(UTILS)))

.PHONY: all clean fmt

all: fmt $(EXES)

fmt:
	gofmt -w .

clean:
	-rm cmd/*.exe
	-rm bin/*.*

bin/%.exe: cmd/%.go
	go build -o $@ $<

bin/%: cmd/%.go
	go build -o $@ $<

