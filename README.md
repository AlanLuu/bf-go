# bf-go
A Go implementation of the bf language

# Usage
```
Usage: bf [OPTIONS] [FILE]

OPTIONS:
  -a int
        Specify initial size of bf array
  -block
        Block array address from going over or under size of bf array; overrides -neg, -static-arr, and -wrap
  -buffer-input
        Enable buffering of standard input
  -buffer-output
        Enable buffering of standard output
  -c string
        Execute bf code from command line argument
  -eof int
        Specify value returned to bf array on EOF; default is 0
  -loop int
        Number of times to repeat bf program; negative number means infinite
  -neg
        Allow negative addresses in bf array
  -no-eof
        Disable returning value to bf array on EOF; overrides -eof
  -no-opts
        Disable bytecode optimizations that improve runtime speed
  -p int
        Specify initial size of bf program
  -sep
        Allow specifying input to bf program in the program itself, separated by '!'
        Example: ',[.,]!Hello world'
  -static-arr
        Exit with error upon going over or under size of bf array; overrides -neg
  -static-program
        Exit with error upon going over size of program
  -tty
        Read in bf code from terminal if stdin is connected to a terminal
  -v    Alias for -version
  -version
        Print version information and exit
  -wrap
        Wrap bf array address upon going over or under size of bf array; overrides -neg, -static-arr, and -block
```

# Installation
First, [install Go](https://go.dev/doc/install) if it's not installed already. Then run the following commands to build:
```
git clone https://github.com/AlanLuu/bf-go.git
cd bf-go/cmd/bf
go build
```
This will create an executable binary called `bf` on Linux/macOS and `bf.exe` on Windows.

# License
This bf interpreter is distributed under the terms of the [MIT License](https://github.com/AlanLuu/bf-go/blob/main/LICENSE).
