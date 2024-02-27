
EXT=
ifeq ($(OS),Windows_NT)
EXT=.exe
endif

UTILS := chrutil romutil fontutil sbutil metatiles text2chr ws2da usage
EXES := $(addsuffix $(EXT), $(addprefix bin/,$(UTILS)))
SRCS := $(addsuffix .go,$(addprefix cmd/,$(UTILS)))

.PHONY: all clean fmt

all: fmt $(EXES)

fmt:
	gofmt -w .

clean:
	-rm -f cmd/*.exe bin/*.* bin/*

bin/chrutil$(EXT): cmd/chrutil.go common/*.go image/*.go
	go build -o $@ $<

bin/romutil$(EXT): cmd/romutil.go rom/*.go
	go build -o $@ $<

bin/sbutil$(EXT): cmd/sbutil.go studybox/*.go
	go build -o $@ $<

bin/fontutil$(EXT): cmd/fontutil.go image/*.go
	go build -o $@ $<

bin/metatiles$(EXT): cmd/metatiles.go image/*.go
	go build -o $@ $<

bin/text2chr$(EXT): cmd/text2chr.go image/*.go
	go build -o $@ $<

bin/ws2da$(EXT): cmd/ws2da.go mesen/*.go
	go build -o $@ $<

bin/usage$(EXT): cmd/usage.go rom/*.go
	go build -o $@ $<
