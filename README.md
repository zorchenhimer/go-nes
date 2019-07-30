# Library and utilities for working with NES data and ROMS

A library for working with NES CHR data and ROM files.

## chrutil

Universal utility to work with CHR files and related data.

- Converting bitmaps to the CHR format
- Converting CHR files to bitmap (or other formats?)
- Selecting which tiles to export (Range or individual IDs)
- Tile de-duplication
- Multiple input bitmaps into a single CHR
- Multiple inputs to multiple outputs?
- 8x16 sprite mode (per-input file)
- Target start ID for given input
- Allow CHR files as input for concatenation
- Output as binary CHR file or assembly source
- Convert bitmap fonts to CHR
- Automatic tile de-duplication
- Destination tile ID override
- Output a re-mapping file that can be read by `bmp2chr` to translate tile IDs.

### Multiple input files

Command line options should probably be ordered.  Something like input file
followed by options.

    $ bmp2chr input_a.bmp input_b.bmp --b-option
    $ bmp2chr input_a.bmp --a-option input_b.bmp
    $ bmp2chr input_a.bmp --a-option input_b.bmp --b-option

    $ bmp2chr \
        --global-option \
        input_a.bmp \
        input_b.bmp --b-option

    $ bmp2chr \
        --global-option \
        input_a.bmp --global-override-for-a \
        input_b.bmp --b-option

For this command, `main.chr` will not contain the data from `input_b.bmp`.

    $ bmp2chr \
        --output main.chr \
        input_a.bmp \
        input_b.bmp --output only_b.chr \
        input_c.bmp

Options will be per-input file or global.  If an option is defined once
globally and once for an input file, the global value is overwritten with the
input file's value.

If an option is given more than once for the same scope, use the last set
value.

### Multiple input and output files

Multiple input files with the same output file will be combined into that
output file.

    $ bmp2chr \
        input_a.bmp --output ab.chr \
        input_b.bmp --output ab.chr \
        input_c.bmp --output cd.chr \
        input_d.bmp --output cd.chr

### Selecting which tiles to export

Either a range or a list of tile IDs.  Accept either decimal or hex ($##)
notation.

    $ bmp2chr --tile-ids 2-14
    $ bmp2chr --tile-ids 2,4,8,10

    $ bmp2chr --tile-ids $02-$0D
    $ bmp2chr --tile-ids $02,$04,$08,$0A

Option to exclude specific tiles or a range?

    $ bmp2chr --tile-ids 2-14 --exclude-ids 4-6
    $ bmp2chr --tile-ids 2-14 --exclude-ids 4,6,10

    $ bmp2chr --tile-ids $02-$0D --exclude-ids $04-$06
    $ bmp2chr --tile-ids $02-$0D --exclude-ids $04,$06,$0A

### Allowing CHR as input

This will append `font.chr` to the end of `main.chr` after `input_a.bmp` has
been converted and written.

    $ bmp2chr \
        --output main.chr \
        input_a.bmp \
        font.chr

This will append the converted `input_a.bmp` CHR data to the end of `main.chr`
after `font.chr`.

    $ bmp2chr \
        --output main.chr \
        font.chr \
        input_a.bmp

## romutil

Utility to work directly with ROM files.

- Unpack ROM into PRG and CHR
- Option to split PRG and CHR into banks
- Pack ROM from unpacked data
- Apply/create patches (IPS and NINJA)
- ROM info printout (header info, hashes, etc)
- Overwrite default page offsets for splitting
- Page usage report
- Read Nestopia (or Mesen's formatted version) database and correct mapper numbers

### Command line

General command format.

    $ romutil <command> <input> [options]

Unpack a ROM into PRG and CHR binary files and a `header.json` file.

    $ romutil unpack input.rom

Re-pack an unpacked ROM.

    $ romutil pack unpacked_data_directory/

Print mapper info and hashes (CRC, MD5, SHA, etc) to the console.

    $ romutil info input.rom

Generate image visualizing ROM usage.  Defaults to replacing the input file's
`.rom` extension with `.png`.

Visualize both total usage, and "data type" stuff.  Eg, make CHR data
distinguishable at a glance (data as shades of grey?).

    $ romutil usage input.rom [options]

Convert to NES2.0 ROM format.  If ROM is already a NES2.0 ROM, verify header
matches data (sizes, etc).

    $ romutil nes2 input.rom
