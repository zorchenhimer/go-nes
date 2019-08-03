package common

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Valid options can only contain alphanumerics and dashes.  Because I said so.
var validLongOption *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
var validShortOption *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9-]$`)

type CommandParser struct {
	// Option definitions and their default values
	AvailableOptions []CommandOption

	// Options before any input file on the command line
	GlobalOptions map[string]string

	// List of files, in the order they appear.  Map of options,
	// starting with "input" for the input filename.
	FileOptions []map[string]string

	// Index for the current FileOptions element. A value of -1
	// means global options.
	currentInput int

	// Has `Parse()` already been called?
	parsed bool
}

// NewCommandParser returns a new empty CommandParser.  Options need to be
// added with `AddOption()`.
func NewCommandParser() *CommandParser {
	return &CommandParser{
		AvailableOptions: []CommandOption{},
		GlobalOptions:    map[string]string{},
		FileOptions:      []map[string]string{},
		currentInput:     -1,
		parsed:           false,
	}
}

// Dump debugging info to console.
func (cp CommandParser) Debug() {
	fmt.Println("Available options:")
	for _, opt := range cp.AvailableOptions {
		fmt.Printf("    %s\n", opt.String())
	}

	fmt.Println("Global options:")
	for key, val := range cp.GlobalOptions {
		fmt.Printf("    %s: %s\n", key, val)
	}

	fmt.Println("Input files:")
	for i, file := range cp.FileOptions {
		fmt.Printf("    %d\n", i)
		for key, val := range file {
			fmt.Printf("        %s: %s\n", key, val)
		}
	}
}

// Parse reads the input args and verifies that only available options
// are supplied.  Validation is only done on argument names and weather
// or not they accept values.  No validation is done on the values themselves.
func (cp *CommandParser) Parse() error {
	// current input scope.  -1 is global.
	ci := -1
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		// If the current element starts with a dash, it's an option.
		// Filenames cannot start with a dash.  Because I said so.
		if strings.HasPrefix(args[i], "-") {
			name := ""
			value := ""
			hasValue := false

			// Check for --opt=val
			if spi := strings.Index(args[i], "="); spi > -1 {
				name = args[i][0:spi]
				value = args[i][spi+1:]
			} else {
				name = args[i]
			}

			ok := false

			// long name
			if strings.HasPrefix(name, "--") {
				if ok, hasValue = cp.isValidLongOption(name[2:]); !ok {
					return fmt.Errorf("Invalid option provided: %q", args[i])
				}

			} else { // short name
				if ok, hasValue, name = cp.isValidShortOption(name[1:]); !ok {
					return fmt.Errorf("Invalid option provided: %q", args[i])
				}
			}

			// hasValue is if the option expects a value, not if one is present.
			if hasValue {
				if i+1 >= len(args) {
					return fmt.Errorf("Missing value for option %q", name)
				}

				// grab the value if it was in the format --opt=val
				if value == "" {
					i += 1
					value = args[i]
				}

			} else {
				// default to "true" for options that do not expect a value.
				value = "true"
			}

			name = strings.TrimLeft(name, "-")

			// Add to global options if no input file has been given so far.
			if ci == -1 {
				cp.GlobalOptions[name] = value
			} else {
				cp.FileOptions[ci][name] = value
			}

		} else {
			ci += 1
			cp.FileOptions = append(cp.FileOptions, map[string]string{
				"input-filename": args[i],
			})
		}
	}

	if len(cp.FileOptions) == 0 {
		return fmt.Errorf("No input files given!")
	}

	cp.parsed = true
	return nil
}

// NextInput moves to the next input file.  Returns false when there is no input
// to move to, true otherwise.
func (cp *CommandParser) NextInput() bool {
	if cp.currentInput+1 >= len(cp.FileOptions) {
		return false
	}
	cp.currentInput += 1
	return true
}

// GetOption returns the named option for the current input file.
// This will get options from the Global scope if `NextInput()` was not called
// before this.
func (cp CommandParser) GetOption(name string) (string, error) {
	if !cp.parsed {
		return "", fmt.Errorf("GetOption called before Parse()")
	}

	if cp.currentInput > -1 {
		if val, ok := cp.FileOptions[cp.currentInput][name]; ok {
			return val, nil
		}
	}

	if val, ok := cp.GlobalOptions[name]; ok {
		return val, nil
	}

	for _, opt := range cp.AvailableOptions {
		if opt.LongName == name {
			return opt.DefaultValue, nil
		}
	}

	return "", fmt.Errorf("Invalid option in GetOption(): %q", name)
}

func (cp CommandParser) GetBoolOption(name string) (bool, error) {
	str, err := cp.GetOption(name)
	if err != nil {
		return false, err
	}

	var val bool
	_, err = fmt.Sscanf(str, "%t", &val)
	if err != nil {
		return false, err
	}

	return val, nil
}

// AddOption adds an available option.  Calling this after Parse() will
// throw an error.  Attempting to add a duplicate or invalid option will
// cause a panic.
func (cp *CommandParser) AddOption(longName, shortName string, expectsValue bool, value string) {
	if cp.parsed {
		panic("AddOption called after Parsed")
	}

	if !validLongOption.MatchString(longName) {
		panic(fmt.Sprintf("Invalid long option format cannot be added: %q", longName))
	}

	// Don't require short names, but validate them if given.
	if shortName != "" && !validShortOption.MatchString(shortName) {
		panic(fmt.Sprintf("Invalid option short format cannot be added: %q", shortName))
	}

	// Don't allow duplicate options
	for _, opt := range cp.AvailableOptions {
		if opt.LongName == longName || opt.LongName == shortName || (shortName != "" && opt.ShortName == shortName) {
			panic(fmt.Sprintf("Duplicate options canont be added: %q, %q", longName, shortName))
		}
	}

	cp.AvailableOptions = append(cp.AvailableOptions, CommandOption{
		LongName:     longName,
		ShortName:    shortName,
		ExpectsValue: expectsValue,
		DefaultValue: value,
	})
}

// isValidLongOption loos for the LongName `name` in AvailableOptions and returns
// true if it is found and if it expects a value.
func (cp CommandParser) isValidLongOption(name string) (bool, bool) {
	for _, opt := range cp.AvailableOptions {
		if opt.LongName == name {
			return true, opt.ExpectsValue
		}
	}
	return false, false
}

// isValidShortOption looks for the ShortName `name` in AvailableOptions and
// returns true if it is found and if it expects a value.  The LongName is also
// returned for the option.
func (cp CommandParser) isValidShortOption(name string) (bool, bool, string) {
	for _, opt := range cp.AvailableOptions {
		if opt.ShortName == name {
			return true, opt.ExpectsValue, opt.LongName
		}
	}
	return false, false, ""
}

type CommandOption struct {
	LongName     string
	ShortName    string
	ExpectsValue bool
	DefaultValue string
	Description  string
}

func (co CommandOption) String() string {
	return fmt.Sprintf("LongName:%q ShortName:%q ExpectsValue:%t DefaultValue:%q",
		co.LongName,
		co.ShortName,
		co.ExpectsValue,
		co.DefaultValue,
		co.Description,
	)
}
