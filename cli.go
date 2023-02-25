package badcli

import "io"
import "os"
import "fmt"
import "sort"
import "image"
import "strings"
import "unicode/utf8"

type CLI struct {
	programName string // cli program name to be used when displaying usage or others
	flags map[string]*flag // maps the long names of the flags, without "--", to their *Flag struct
	flagShortAliases map[rune]string
	extraArgs []string
	extraArgsDisallowed bool
	extraUsageSections []string
	helpDescription string
	// TODO: add explicit example usages?
	// TODO: usage pattern / scheme ? like, prog-name [--flags] path/to/file.png
	//       (though I personally prefer usage examples right away)
}

// Creates a [*CLI] struct for parsing command line arguments.
// The programName is used for displaying usage and help, and
// the help description is shown at the start if -h or --help
// are used when invoking the program.
func NewCLI(programName string, helpDescription string) *CLI {
	return &CLI{
		programName: programName,
		flags: make(map[string]*flag),
		flagShortAliases: make(map[rune]string),
		helpDescription: helpDescription,
	}
}

func (self *CLI) DisallowExtraArgs() {
	self.extraArgsDisallowed = true
}

func (self *CLI) ExtraArgs() []string {
	return self.extraArgs
}

// Parse will read command line arguments, parse them, and exit
// with an error code if there's any error during parsing.
func (self *CLI) ParseArguments() {
	const SeeHelp = "Further help: %s --help\n"

	args := os.Args[1:]
	index := 0
	for index < len(args) {
		arg := args[index]
		if arg == "-h" || arg == "--help" || arg == "/?" {
			fmt.Print(self.helpDescription, "\n\n")
			self.PrintUsage(os.Stdout)
			os.Exit(0)
		}

		if strings.HasPrefix(arg, "--") {
			// check if flag is known
			flagName := arg[2 : ]
			flagPtr, found := self.flags[flagName]
			if !found {
				msg := "Failed to parse '%s' argument:\n\tflag name not recognized\n"
				if len(arg) > 2 {
					key := self.FindCloseFlagName(flagName)
					if key != "" {
						msg += "(Maybe you meant '--" + key + "'?)\n"
					}
				}
				fmt.Fprintf(os.Stderr, msg + "\n" + SeeHelp, arg, self.programName)
				os.Exit(2)
			}

			// check redundant flag
			if flagPtr.SetByUser {
				msg := "Duplicated flag '%s'. Program flags can't be repeated.\n"
				fmt.Fprintf(os.Stderr, msg + SeeHelp, arg, self.programName)
				os.Exit(2)
			}

			// get next argument to parse flag
			if index + 1 >= len(args) { // next value is missing
				err := flagPtr.Value.ParseFromArg("")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to parse '%s' argument:\n", arg)
					EachLine(err.Error(), 74, func(line string) error {
						fmt.Fprint(os.Stderr, "\t", line, "\n")
						return nil
					})
					fmt.Fprintf(os.Stderr, SeeHelp, self.programName)
					os.Exit(2)
				}
			} else { // obtain next value
				index += 1
				nextArg := args[index]
				err := flagPtr.Value.ParseFromArg(nextArg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to parse '%s %s' arguments:\n", arg, nextArg)
					EachLine(err.Error(), 74, func(line string) error {
						fmt.Fprint(os.Stderr, "\t", line, "\n")
						return nil
					})
					fmt.Fprintf(os.Stderr, SeeHelp, self.programName)
					os.Exit(2)
				}
			}

			// set flag as parsed
			flagPtr.SetByUser = true
		} else if strings.HasPrefix(arg, "-") {
			// short flag
			if runeLenAbove(arg, 2) {
				msg := "Failed to parse '%s' argument:\n" +
					"\tmulti-letter flags not allowed for single dash flags\n"
				fmt.Fprintf(os.Stderr, msg + SeeHelp, arg, self.programName)
				os.Exit(2)
			}

			// TODO: actually I should translate to long flag and just use it
			panic("TODO")
		} else {
			// extra argument
			if self.extraArgsDisallowed {
				msg := "Unexpected '%s' argument.\n"
				fmt.Fprintf(os.Stderr, msg + SeeHelp, arg, self.programName)
				os.Exit(2)
			} else {
				self.extraArgs = append(self.extraArgs, arg)
			}
		}

		index += 1
	}
}

// Add an extra usage section displayed after the arguments.
func (self *CLI) AddUsageSection(paragraph string) {
	self.extraUsageSections = append(self.extraUsageSections, paragraph)
}

func (self *CLI) RegisterFlag(longFlagName, usage string, value FlagValue, aliases ...rune) {
	// safety checks
	if !runeLenAbove(longFlagName, 1) {
		panic("flag name must have at least 2 characters ('" + longFlagName + "')")
	}

	if longFlagName[0] == '-' {
		panic("flag name can't start with a dash ('" + longFlagName + "')")
	}

	if self.IsFlagRegistered(longFlagName) {
		panic("flag name already registered ('" + longFlagName + "')")
	}
	
	if value == nil {
		panic("can't register flag with nil value")
	}
	
	// actual registration
	self.flags[longFlagName] = &flag{
		Value: value,
		Usage: usage,
	}
}

// Returns whether the given long flag name is registered or not.
// For aliases, check [CLI.AliasToFullFlag]() instead.
func (self *CLI) IsFlagRegistered(longFlagName string) bool {
	_, alreadyRegistered := self.flags[longFlagName]
	return alreadyRegistered
}

// The returned string will be empty if the alias doesn't exist.
func (self *CLI) AliasToFullFlag(alias rune) string {
	return self.flagShortAliases[alias]
}

func (self *CLI) RegisterShortAliases(longFlagName string, aliases ...rune) {
	if !self.IsFlagRegistered(longFlagName) {
		panic("can't register aliases for inexistent '" + longFlagName + "' flag")
	}
	
	for _, alias := range aliases {
		existingFlag, alreadyDefined := self.flagShortAliases[alias]
		
		// safety assertions
		if alreadyDefined {
			a := string(alias)
			if existingFlag == longFlagName {
				panic("repeated registration of alias '" + a + "' to '" + longFlagName + "'")
			} else {
				panic("can't register alias '" + a + "' to '" + longFlagName + "' " +
				      "(already registered to '" + existingFlag + "')")
			}
		}

		// actual registration
		self.flagShortAliases[alias] = longFlagName
	}
}

// Returns the flag value for the given full flag name.
// If [CLI.ParseArguments]() hasn't been called yet, only nil or
// default values can be returned. See also [CLI.FlagSetByUser]().
func (self *CLI) GetFlagValue(fullFlagName string) FlagValue {
	flagPtr, found := self.flags[fullFlagName]
	if !found { return nil }
	if flagPtr == nil { panic("internal code error") }
	return flagPtr.Value
}

// Returns whether a flag has been explicitly set by the
// user or not. A flag may still have a default value even
// if it hasn't been set by the user.
//
// The name passed must be the long form of the flag name.
func (self *CLI) FlagSetByUser(longFlagName string) bool {
	flagPtr, found := self.flags[longFlagName]
	if !found { return false }
	return flagPtr.SetByUser
}

// Returns whether any of the given flags have been passed
// to the program. The names passed must be the long form of
// the flag names.
func (self *CLI) AnyFlagSetByUser(fullFlagNames ...string) bool {
	for _, fullFlagName := range fullFlagNames {
		if self.FlagSetByUser(fullFlagName) {
			return true
		}
	}
	return false
}

// Returns whether all the given flags have been passed
// to the program. The names passed must be the long form
// of the flag names.
func (self *CLI) AllFlagsSetByUser(fullFlagNames ...string) bool {
	for _, fullFlagName := range fullFlagNames {
		if !self.FlagSetByUser(fullFlagName) {
			return false
		}
	}
	return true
}

func (self *CLI) UsageFail(fmtStr string, args ...any) {
	fmt.Fprint(os.Stderr, "Invalid usage:\n")
	EachLine(fmt.Sprintf(fmtStr, args...), 74, func(line string) error {
		fmt.Fprint(os.Stderr, "\t", line, "\n")
		return nil
	})
	fmt.Fprintf(os.Stderr, "\n")
	self.PrintUsage(os.Stderr)
	os.Exit(2)
}

func (self *CLI) PrintUsage(output io.Writer) {
	fmt.Fprintf(output, "Usage of %s:\n", self.programName)

	reverseAliases := make(map[string][]rune)
	for aliasLetter, aliasedFlag := range self.flagShortAliases {
		reverseAliases[aliasedFlag] = append(reverseAliases[aliasedFlag], aliasLetter)
	}

	// find flag usage description lengths
	usageSplits := make([]split , len(self.flags))
	flagNames   := make([]string, len(self.flags))
	flagIndices := make([]int   , len(self.flags))
	flagIndex := 0
	for flagLongName, flagPtr := range self.flags {
		flagNameLen := utf8.RuneCountInString(flagLongName) + 2
		flagNameLen += len(reverseAliases[flagLongName])*4
		descrLen := utf8.RuneCountInString(flagPtr.Usage)
		usageSplits[flagIndex] = split{ leftLen: uint16(flagNameLen), rightLen: uint16(descrLen) }
		flagNames[flagIndex]   = flagLongName
		flagIndices[flagIndex] = flagIndex
		flagIndex += 1
	}
	// TODO: I'm not considering the case of line breaks within Usage descriptions.
	//       I don't know if I should check, but... I guess it's ok to ignore ftm.

	tabSize := uint16(4) // approximate, we don't really know what the terminal does
	flagVsDescrSpacing := uint16(4)
	contentLen := 80 - tabSize - flagVsDescrSpacing
	maxFlagLen := findBreakpointMin(usageSplits, contentLen)
	
	// sort flags alphabetically
	sort.Slice(flagIndices, func(i, j int) bool {
		return flagNames[flagIndices[i]] < flagNames[flagIndices[j]]
	})

	// first flags iteration, print short lines
	var strBuilder strings.Builder
	for _, index := range flagIndices {
		split := usageSplits[index]
		spacesNeeded := int(maxFlagLen) - int(split.leftLen)
		if spacesNeeded >= 0 && split.leftLen + split.rightLen <= contentLen {
			flagName := flagNames[index]
			strBuilder.Reset()
			strBuilder.WriteString("\t--")
			strBuilder.WriteString(flagName)
			for _, letter := range reverseAliases[flagName] {
				strBuilder.WriteString(", -")
				strBuilder.WriteRune(letter)
			}
			for i := spacesNeeded + int(flagVsDescrSpacing); i > 0; i-- {
				strBuilder.WriteByte(' ')
			}
			strBuilder.WriteString(self.flags[flagName].Usage)
			strBuilder.WriteByte('\n')
			fmt.Fprint(output, strBuilder.String())
		}
	}
	
	// second flags iteration, print long lines
	for _, index := range flagIndices {
		split := usageSplits[index]
		if split.leftLen > maxFlagLen || split.leftLen + split.rightLen > contentLen {
			flagName := flagNames[index]
			strBuilder.Reset()
			strBuilder.WriteString("\t--")
			strBuilder.WriteString(flagName)
			for _, letter := range reverseAliases[flagName] {
				strBuilder.WriteString(", -")
				strBuilder.WriteRune(letter)
			}
			strBuilder.WriteByte('\n')
			fmt.Fprint(output, strBuilder.String())
			EachLine(self.flags[flagName].Usage, 70, func(line string) error {
				fmt.Fprint(output, "\t     ", line, "\n")
				return nil
			})
		}
	}
	
	// write additional paragraphs, if relevant
	for _, section := range self.extraUsageSections {
		fmt.Fprint(output, "\n")
		EachLine(section, 80, func(line string) error {
			fmt.Fprint(output, line, "\n")
			return nil
		})
	}
}

// Similar to fmt.Fprintf(os.Stderr, "Warning: ", ...).
// An extra \n is always added at the end, so don't add it yourself.
func (self *CLI) Warn(fmtStr string, args ...any) {
	fmt.Fprint(os.Stderr, "Warning: ")
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	fmt.Fprint(os.Stderr, "\n")
}

// Similar to fmt.Fprintf(os.Stderr, ...) and os.Exit(1).
// An extra \n is always added at the end, so don't add it yourself.
func (self *CLI) Fatal(fmtStr string, args ...any) {
	self.printFatalMsg(fmt.Sprintf(fmtStr, args...))
}

func (self *CLI) FatalErr(err error) {
	if err == nil { panic("nil error") }
	self.printFatalMsg(err.Error())
}

func (self *CLI) printFatalMsg(msg string) {
	if len(msg) <= 67 && utf8.RuneCountInString(msg) <= 67 {
		fmt.Fprint(os.Stderr, "Fatal error: ", msg, "\n")
	} else {
		fmt.Fprint(os.Stderr, "Fatal error:\n")
		EachLine(msg, 74, func(line string) error {
			fmt.Fprint(os.Stderr, "\t", line, "\n")
			return nil
		})
	}
	os.Exit(1)
}

// Equivalent to fmt.Printf. Ignores errors.
func (self *CLI) Printf(fmtStr string, args ...any) {
	fmt.Printf(fmtStr, args...)
}

// Equivalent to fmt.Print. Ignores errors.
func (self *CLI) Print(args ...any) { fmt.Print(args...) }

// Defined for [CLI.ExportImage]().
type ImageExportFunc = func(io.Writer, image.Image) error

func (self *CLI) ExportImage(path string, img image.Image, exportFn ImageExportFunc) {
	file, err := os.Create(path)
   if err != nil {
		self.Fatal("Failed to create '%s' for image export: %s", path, err)
	}
	
	// proceed with export
	err = exportFn(file, img)
	if err != nil {
		// try to perform cleanup
		cleanupErr := os.Remove(path)
		if cleanupErr != nil {
			if strings.Contains(cleanupErr.Error(), path) {
				self.Warn("couldn't clean up failed image export: %s", cleanupErr)
			} else {
				self.Warn("couldn't clean up failed image export at '%s': %s", path, cleanupErr)
			}	
		}

		if strings.Contains(err.Error(), path) {
			self.Fatal("failed to encode image to file: %s", err)
		} else {
			self.Fatal("failed to encode image '%s' to file: %s", path, err)
		}
	}
}

// May return an empty string if no close / good match exists.
func (self *CLI) FindCloseFlagName(longFlagName string) string {
	// Note 1: unicode normalization is probably not unnecessary in theory,
	//         but it should virtually always be unnecessary in practice.
	//         so I guess I'll leave that out for the moment.
	// Note 2: this algorithm is non-deterministic if two best options
	//         exist due to map rng iteration. Should I correct that?
	nearestEditFlag := ""
	lowestEditDist  := 65535
	costCutoff := len(longFlagName)/2 + 1
	if costCutoff < 7 { costCutoff = 7 }
	for candidateFlagName, _ := range self.flags {
		dist := EditDistance(longFlagName, candidateFlagName, costCutoff)
		if dist < lowestEditDist {
			lowestEditDist = dist
			nearestEditFlag = candidateFlagName
			if dist == 0 { break }
		}
	}

	if lowestEditDist < costCutoff {
		// get longest rune length (could optimize by making EditDistance
		// return some extra runeLen info, but not a big deal either way)
		longLen := utf8.RuneCountInString(longFlagName)
		nearLen := utf8.RuneCountInString(nearestEditFlag)
		if nearLen > longLen { longLen = nearLen }
		
		// compute similarity rate
		similarity := float64(longLen - lowestEditDist)/float64(longLen)
		if similarity >= 0.5 || longLen <= 3 {
			return nearestEditFlag
		}
	}
	
	return "" // no match
}
