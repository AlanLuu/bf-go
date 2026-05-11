package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/AlanLuu/bf-go"
)

const (
	PROGRAM_NAME  = "bf-go"
	ttyPipeErrStr = "cannot enable -%%v flag when stdin is from pipe"
)

var (
	stdinIsTerminal = false
	useTTY          = false
)

func set_stdinIsTerminal_var() {
	stat, _ := os.Stdin.Stat()
	if stat != nil {
		stdinIsTerminal = stat.Mode()&os.ModeCharDevice != 0
	}
}

func addQuotes(s string) string {
	if isWindows() {
		return "\"" + s + "\""
	}
	return "'" + s + "'"
}

func anyFlagProvided(names ...string) bool {
	found := false
	for _, name := range names {
		flag.Visit(func(f *flag.Flag) {
			if f.Name == name {
				found = true
			}
		})
		if found {
			break
		}
	}
	return found
}

func errorf(format string, a ...any) error {
	if len(os.Args) > 0 && os.Args[0] != "" {
		return fmt.Errorf(
			filepath.Base(os.Args[0])+": "+format,
			a...,
		)
	}
	return fmt.Errorf(bf.ErrMsgHeader+format, a...)
}

func gitInfo() (hash string, modified bool) {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		foundHash := false
		foundModified := false
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				foundHash = true
				hash = setting.Value
			case "vcs.modified":
				foundModified = true
				modified = setting.Value != "false"
			}
			if foundHash && foundModified {
				return
			}
		}
	}
	return
}

func gitInfoShort() (hash string, modified bool) {
	const upper = 8
	if hash, modified = gitInfo(); len(hash) > upper {
		hash = hash[:upper]
	}
	return
}

func perrorf(format string, a ...any) {
	if len(os.Args) > 0 && os.Args[0] != "" {
		fmt.Fprintf(
			os.Stderr,
			filepath.Base(os.Args[0])+": "+format,
			a...,
		)
	} else {
		fmt.Fprintf(os.Stderr, bf.ErrMsgHeader+format, a...)
	}
}

func perrorfln(format string, a ...any) {
	perrorf(format+"\n", a...)
}

func perrorln(s string) {
	if len(os.Args) > 0 && os.Args[0] != "" {
		fmt.Fprintln(
			os.Stderr,
			filepath.Base(os.Args[0])+": "+s,
		)
	} else {
		fmt.Fprintln(os.Stderr, bf.ErrMsgHeader+s)
	}
}

func getBfProgram() (string, bool, error) {
	args := flag.Args()
	var reader io.Reader
	if len(args) == 0 {
		set_stdinIsTerminal_var()
		if stdinIsTerminal && !useTTY {
			// stdin is from terminal and useTTY flag is disabled
			return "", false, nil
		}
		if useTTY {
			if !stdinIsTerminal {
				return "", false, errorf(ttyPipeErrStr)
			}
			perrorln("Reading bf program from terminal...")
			if isWindows() {
				perrorln("Press Ctrl+Z followed by Enter to stop reading.")
			} else {
				perrorln("Press Ctrl+D to stop reading.")
			}
		}
		reader = os.Stdin
	} else {
		file, fileErr := os.Open(args[0])
		if fileErr != nil {
			return "", false, fileErr
		}
		defer file.Close()
		reader = file
	}
	builder := new(strings.Builder)
	_, copyErr := io.Copy(builder, reader)
	if copyErr != nil {
		return "", false, copyErr
	}
	return builder.String(), true, nil
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func printVersion() {
	var builder strings.Builder
	builder.WriteString(PROGRAM_NAME)
	if hash, modified := gitInfoShort(); hash != "" {
		builder.WriteByte('-')
		builder.WriteString(hash)
		if modified {
			builder.WriteString(" (modified)")
		}
	} else {
		builder.WriteString(" (no git hash)")
	}
	builder.WriteByte(' ')
	builder.WriteByte('(')
	builder.WriteString(runtime.Version())
	builder.WriteString("-" + runtime.GOOS + "-" + runtime.GOARCH)
	builder.WriteByte(')')
	fmt.Println(builder.String())
}

func realMain() int {
	const (
		cFlagName             = "c"
		arrSizeName           = "a"
		programSizeName       = "p"
		eofFlagName           = "eof"
		noEofFlagName         = "no-eof"
		noOptsFlagName        = "no-opts"
		arrNegAddrName        = "neg"
		arrStaticSizeName     = "static-arr"
		programStaticSizeName = "static-program"
		arrBlockName          = "block"
		loopName              = "loop"
		loopInfName           = "loop-inf"
		arrWrapAroundName     = "wrap"
		sepInputName          = "sep"
		ttyFlagName           = "tty"
		bufferInputName       = "buffer-input"
		bufferOutputName      = "buffer-output"
		versionName1          = "version"
		versionName2          = "v"
	)
	var (
		cFlag = flag.String(
			cFlagName,
			"",
			"Execute bf code from command line argument",
		)
		arrSize = flag.Int(
			arrSizeName,
			0,
			"Specify initial size of bf array",
		)
		programSize = flag.Int(
			programSizeName,
			0,
			"Specify initial size of bf program",
		)
		eofFlag = flag.Int(
			eofFlagName,
			0,
			"Specify value returned to bf array on EOF; "+
				"default is 0",
		)
		noEofFlag = flag.Bool(
			noEofFlagName,
			false,
			"Disable returning value to bf array on EOF; "+
				"overrides -"+eofFlagName,
		)
		noOptsFlag = flag.Bool(
			noOptsFlagName,
			false,
			"Disable bytecode optimizations that improve runtime speed",
		)
		arrNegAddr = flag.Bool(
			arrNegAddrName,
			false,
			"Allow negative addresses in bf array",
		)
		arrStaticSize = flag.Bool(
			arrStaticSizeName,
			false,
			"Exit with error upon going over or under size of bf array; "+
				"overrides -"+arrNegAddrName,
		)
		programStaticSize = flag.Bool(
			programStaticSizeName,
			false,
			"Exit with error upon going over size of program",
		)
		arrBlock = flag.Bool(
			arrBlockName,
			false,
			"Block array address from going over or under size of bf array; "+
				"overrides -"+arrNegAddrName+", -"+arrStaticSizeName+
				", and -"+arrWrapAroundName,
		)
		loopTimes = flag.Int64(
			loopName,
			1,
			"Number of times to run bf program; "+
				"negative number means infinite",
		)
		loopInf = flag.Bool(
			loopInfName,
			false,
			"Run bf program infinitely; overrides -"+loopName,
		)
		arrWrapAround = flag.Bool(
			arrWrapAroundName,
			false,
			"Wrap bf array address upon going over or under size "+
				"of bf array; overrides -"+arrNegAddrName+", -"+
				arrStaticSizeName+", and -"+arrBlockName,
		)
		sepInput = flag.Bool(
			sepInputName,
			false,
			"Allow specifying input to bf program in the program itself, "+
				"separated by '!'\n"+
				"Example: "+addQuotes(",[.,]!Hello world"),
		)
		ttyFlag = flag.Bool(
			ttyFlagName,
			false,
			"Read in bf code from terminal if stdin is connected to a terminal",
		)
		bufferInput = flag.Bool(
			bufferInputName,
			false,
			"Enable buffering of standard input",
		)
		bufferOutput = flag.Bool(
			bufferOutputName,
			false,
			"Enable buffering of standard output",
		)
		version1 = flag.Bool(
			versionName1,
			false,
			"Print version information and exit",
		)
		version2 = flag.Bool(
			versionName2,
			false,
			"Alias for -"+versionName1,
		)
	)
	flag.Parse()
	if *version1 || *version2 {
		printVersion()
		return 0
	}
	useTTY = *ttyFlag

	var onEOF int16 = 0
	if *noEofFlag {
		onEOF = bf.DoNothingOnEOF
	} else if *eofFlag != 0 {
		const (
			bits  = 16
			lower = -(1 << (bits - 1))
			upper = (1 << (bits - 1)) - 1
		)
		if *eofFlag < lower || *eofFlag > upper {
			perrorfln(
				"value of -%v flag must be from %v to %v",
				eofFlagName,
				lower,
				upper,
			)
			return 1
		}
		onEOF = int16(*eofFlag)
	}

	var bfStr string
	if anyFlagProvided(cFlagName) {
		bfStr = *cFlag
	} else {
		var ok bool
		var err error
		bfStr, ok, err = getBfProgram()
		if err != nil {
			switch errStr := err.Error(); {
			case strings.HasSuffix(
				errStr,
				strings.ReplaceAll(ttyPipeErrStr, "%%", "%"),
			):
				fmt.Fprintf(os.Stderr, errStr+"\n", ttyFlagName)
			default:
				fmt.Fprintln(os.Stderr, err)
			}
			return 1
		}
		if !ok {
			perrorfln(
				"missing -%v flag or file name argument",
				cFlagName,
			)
			perrorfln(
				"pass the -%v flag to read from terminal if "+
					"stdin is connected to one",
				ttyFlagName,
			)
			flag.Usage()
			return 1
		}
	}

	if *arrStaticSize {
		*arrNegAddr = false
	}
	if *arrBlock {
		*arrNegAddr = false
		*arrStaticSize = false
		*arrWrapAround = false
	}
	if *arrWrapAround {
		*arrNegAddr = false
		*arrStaticSize = false
		*arrBlock = false
	}
	var reader io.Reader
	if *bufferInput {
		reader = bufio.NewReader(os.Stdin)
	}
	var writer io.Writer
	if *bufferOutput {
		writer = bufio.NewWriter(os.Stdout)
	}
	var loopTimesNum int64
	if *loopInf {
		loopTimesNum = -1
	} else {
		loopTimesNum = *loopTimes
	}
	bfOptions := &bf.Options{
		ArrSize:              *arrSize,
		ProgramSize:          *programSize,
		LoopTimes:            loopTimesNum,
		ArrHasNegAddress:     *arrNegAddr,
		ArrHasStaticSize:     *arrStaticSize,
		ProgramHasStaticSize: *programStaticSize,
		ArrBlock:             *arrBlock,
		ArrWrapAround:        *arrWrapAround,
		SepInput:             *sepInput,
		NoOptimizations:      *noOptsFlag,
		OnEOF:                onEOF,
		Reader:               reader,
		Writer:               writer,
	}
	if err := bf.Execute(bfStr, bfOptions); err != nil {
		fmt.Fprintln(os.Stderr, err)
		notStatic := !*arrStaticSize
		errStr := err.Error()
		const negArrErrMsg = bf.NegativeArrErrMsg
		if notStatic && strings.HasSuffix(errStr, negArrErrMsg) {
			fmt.Fprint(os.Stderr, bf.ErrMsgHeader)
			fmt.Fprintf(
				os.Stderr,
				"pass the -%v flag to allow negative addresses in bf array\n",
				arrNegAddrName,
			)
		}
		return 1
	}
	return 0
}

func main() {
	os.Exit(realMain())
}
