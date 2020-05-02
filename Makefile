
EXT=
ifeq ($(OS),Windows_NT)
EXT=.exe
endif

UTILS := chrutil romutil sbutil
EXES := $(addsuffix .exe,$(addprefix bin/,$(UTILS)))

.PHONY: all clean fmt

all: fmt $(EXES)

fmt:
	gofmt -w .

clean:
	rm -f cmd/*.exe
	rm -f bin/*.*

bin/chrutil$(EXT): cmd/chrutil.go common/*.go image/*.go
	go build -o $@ $<

bin/romutil$(EXT): cmd/romutil.go ines/*.go
	go build -o $@ $<

bin/sbutil$(EXT): cmd/sbutil.go studybox/*.go
	go build -o $@ $<
