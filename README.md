# Library and utilities for working with NES data and ROMS

A library for working with NES CHR data and ROM files.

## chrutil

Utility to work with CHR files and related data.

- Converting bitmaps to the CHR format
- Converting CHR files to PNG
- Selecting which tiles to export (Range or individual IDs)
- Tile de-duplication (with ID mappings exported)
- Multiple input bitmaps into a single CHR
- Multiple inputs to multiple outputs
- 8x16 sprite mode (per-input file)
- Target start ID for given input
- Allow CHR files as input for concatenation
- Output as binary CHR file or assembly source
- Destination tile ID override

### Custom rolled flag parsing (only chrutil)

Three main sections (or "targets") of data: default, global, and per-input.

Default is defined at compile-time, while the global and per-input are defined
at runtime.

All command line options have a target, with the default (first) target
"global".  Retrieving a value for a target will first look in the current
target scope, fall back to the global target if nothing was found, then finally
fall back to the default target.

Each target should have a reference to the higher target scope.  The default
target will not have a reference anywhere as all options will be expected to
have been defined in the default target scope.

### Multiple input files

The order of commands and input files matters.  Options given before an input
file are global for all input files, unless overwritten later.

    $ chrutil input_a.bmp input_b.bmp --b-option
    $ chrutil input_a.bmp --a-option input_b.bmp
    $ chrutil input_a.bmp --a-option input_b.bmp --b-option

    $ chrutil \
        --global-option \
        input_a.bmp \
        input_b.bmp --b-option

    $ chrutil \
        --global-option \
        input_a.bmp --global-override-for-a \
        input_b.bmp --b-option

For this command, `main.chr` will not contain the data from `input_b.bmp`.

    $ chrutil \
        --output main.chr \
        input_a.bmp \
        input_b.bmp --output only_b.chr \
        input_c.bmp

Options are per-input file or global.  If an option is defined once globally
and once for an input file, the global value is overwritten with the input
file's value.

If an option is given more than once for the same scope the last value will be
used.

### Multiple input and output files

Multiple input files with the same output file will be combined into that
output file.

    $ chrutil \
        input_a.bmp --output ab.chr \
        input_b.bmp --output ab.chr \
        input_c.bmp --output cd.chr \
        input_d.bmp --output cd.chr

### Selecting which tiles to export

Either a range or a list of tile IDs.  Accept either decimal or hex ($##)
notation.

    $ chrutil --tile-ids 2-14
    $ chrutil --tile-ids 2,4,8,10

    $ chrutil --tile-ids $02-$0D
    $ chrutil --tile-ids $02,$04,$08,$0A

### CHR as input

This will append `font.chr` to the end of `main.chr` after `input_a.bmp` has
been converted and written.

    $ chrutil \
        --output main.chr \
        input_a.bmp \
        font.chr

This will append the converted `input_a.bmp` CHR data to the end of `main.chr`
after `font.chr`.

    $ chrutil \
        --output main.chr \
        font.chr \
        input_a.bmp

## fontutil

Takes a font in an image and removes blank tiles.  This can also output ca65
remapping commands so ascii typed in source files will be able to use the
generated font without any modification.

    fontutil --input font.bmp --output converted.chr \
             --remap character-remappings.i \
             --widths character-widths.i \
             --space-width 5 \
             --input-offset 0 \
             --input-length 0

`--input` and `--output` are the only required options.

An `--input-length` of zero disables the length check and will process all the
characters in the font.

## metatiles

Convert metatiles in an image to tile-reduced CHR data and metadata that can be
used to re-construct the metatiles.  The positional arguments are required, all
the others are optional.

    $ metatiles input.bmp output.chr metadata.i
    $ metatiles input.bmp output.chr metadata.i \
                --tile-size 2x2 \
                --count 2 \
                --offeset 1 \
                --pad 16

`--tile-size` defaults to 2x2.  This option accepts a single number (square
metatile) or a WxH value.

`--count` will process only the given number of metatiles.  A value of `0` will
process all metatiles found in the input file.

`--offset` will start processing metatiles after the given number of metatiles
(not individual 8x8 tiles).

`--pad` will pad the output CHR to contain at least the given number of 8x8
tiles.

### Output Metadata

This data is used to reconstruct the metatiles from individual 8x8 pixel tiles.
The data consists of two tables.  The first is a list of addresses (pointers)
to data for each metatile.

The data for each mitatile consists of a few vaules:

    ; width, height
    .byte 2, 2
    ; palette, total tiles (W*H)
    .byte 0, 4
    ; list of tile IDs
    .byte 128, 129, 128, 129

## romutil

Utility to work directly with ROM files.

- Unpack ROM into PRG and CHR
- Option to split PRG and CHR into banks
- Pack ROM from unpacked data
- ROM info printout (header info, hashes, etc)

### Command line

General command format.

    $ romutil <command> <input> [options]

Unpack a ROM into PRG and CHR binary files and a `header.json` file.

    $ romutil unpack input.nes

Re-pack an unpacked ROM.

    $ romutil pack unpacked_data_directory/

Print mapper info and CRC32 hashes

    $ romutil info input.nes

## sbutil

An (unfinished) utility to pack and unpack StudyBox rom files.

## text2chr

Create a tile-reduced text image.  Letters are assumed to be variable width.

    $ text2chr --font font.bmp --metadata text.i --input "Hello world" --chr output.chr

The file `font.bmp` is the font to use for the conversion and the text to
convert is provided with the `--input` option.  Metadata is written to
`text.i` and CHR data to `output.chr`.  The metadata consists of a data
length (one byte) followed by that number tile IDs.

## usage

Create an image showing the usage of a ROM.

    $ usage --chr-size 8 input.nes output.png

The output image is made up of columns of data.  Each column represents 16k and
is 16 bytes wide.  Each pixel in the image is a single bit in the ROM.

CHR data is also written out to a set of images.  By default, the images are
split into 8k chunks.  Valid sizes are: 8, 4, 2, 1, & 0.  A value of "0" will
not split the CHR data into chunks and will write a single image.

Currently, the output filenames for the CHR data conforms to "chr_%04d.png".
This will be configurable, eventually.
