
.PHONY: img bmp all

all: img bmp unpack_rom.exe pack_rom.exe

bmp: bmp2chr.exe

img: chr2img.exe

bmp2chr.exe: bmp2chr.go ../image/*.go
	go build -o bmp2chr.exe bmp2chr.go

chr2img.exe: chr2img.go ../ines/*.go
	go build -o chr2img.exe chr2img.go

unpack_rom.exe: unpack_rom.go ../ines/*.go
	go build -o unpack_rom.exe unpack_rom.go

pack_rom.exe: pack_rom.go ../ines/*.go
	go build -o pack_rom.exe pack_rom.go

clean:
	-rm *.exe

