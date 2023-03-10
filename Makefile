
EXT=
ifeq ($(OS),Windows_NT)
EXT=.exe
endif

UTILS := chrutil romutil fontutil sbutil metatiles text2chr
EXES := $(addprefix bin/,$(UTILS))
SRCS := $(addsuffix .go,$(addprefix cmd/,$(UTILS)))

.PHONY: all clean fmt

all: fmt $(EXES)

fmt:
	gofmt -w .

clean:
	-rm -f cmd/*.exe bin/*.* bin/*

bin/chrutil$(EXT): cmd/chrutil.go common/*.go image/*.go
	go build -o $@ $<

bin/romutil$(EXT): cmd/romutil.go ines/*.go
	go build -o $@ $<

bin/sbutil$(EXT): cmd/sbutil.go studybox/*.go
	go build -o $@ $<

bin/fontutil$(EXT): cmd/fontutil.go image/*.go
	go build -o $@ $<

bin/metatiles$(EXT): cmd/metatiles.go image/*.go
	go build -o $@ $<

bin/text2chr$(EXT): cmd/text2chr.go image/*.go
	go build -o $@ $<
