
#EXT=
#
#ifeq ($(OS),Windows_NT)
#	EXT=.exe
#endif

#COMMANDS= bmp2chr

.PHONY: all clean fmt

all: fmt cmd/bmp2chr.exe

fmt:
	gofmt -w .

clean:
	rm cmd/*.exe

cmd/bmp2chr.exe: cmd/bmp2chr.go
	go build -o cmd/bmp2chr.exe cmd/bmp2chr.go

