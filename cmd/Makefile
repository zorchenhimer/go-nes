
.PHONY: img bmp all

all: img bmp

bmp: bmp2chr.exe

img: chr2img.exe

bmp2chr.exe: bmp2chr.go ../image/*.go
	go build -o bmp2chr.exe bmp2chr.go

chr2img.exe: chr2img.go ../ines/*.go
	go build -o chr2img.exe chr2img.go

clean:
	-rm *.exe

