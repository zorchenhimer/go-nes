
EXT=
ifeq ($(OS),Windows_NT)
EXT=.exe
endif

UTILS := chrutil romutil
EXES := $(addsuffix .exe,$(addprefix bin/,$(UTILS)))
SRCS := $(addsuffix .go,$(addprefix cmd/,$(UTILS)))

.PHONY: all clean fmt

all: fmt $(EXES)

fmt:
	gofmt -w .

clean:
	rm -f cmd/*.exe
	rm -f bin/*.*

bin/%.exe: cmd/%.go image/*.go common/*.go ines/*.go
	go build -o $@ $<

bin/%: cmd/%.go image/*.go common/*.go ines/*.go
	go build -o $@ $<

