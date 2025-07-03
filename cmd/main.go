package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/brunolkatz/goxorstring"
	"go/format"
	"os"
	"regexp"
	"strings"
	"text/template"
	"unicode"
)

type Assets struct {
	VarName string
	Path    string
	Value   func(path string) []byte // Will panic when any error occurs
}

func main() {
	config := parseArgs()

	if len(config.Files) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No input files provided.\n\n")
		return
	}

	knownVars := make(map[string]int)
	assets := make([]*Assets, 0)
	for _, file := range config.Files {
		asset := &Assets{
			VarName: safeVarName(file, knownVars),
			Path:    file,
			Value: func(path string) []byte {
				f, err := os.ReadFile(path)
				if err != nil {
					panic(fmt.Sprintf("Error reading file %s: %v", path, err))
				}
				if len(f) == 0 {
					return []byte{}
				}
				return f
			},
		}

		assets = append(assets, asset)
	}

	xors := make(map[string]*goxorstring.XorString)
	for _, asset := range assets {
		xor := goxorstring.NewXorString(string(asset.Value(asset.Path)))
		xors[asset.VarName] = xor
	}

	t, err := template.ParseFiles("cmd/output.tmpl")
	if err != nil {
		panic(fmt.Sprintf("Error parsing template: %v", err))
	}

	req := struct {
		PackageName string
		Xors        map[string]*goxorstring.XorString
	}{
		PackageName: config.PackageName,
		Xors:        xors,
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, req)
	if err != nil {
		panic(fmt.Sprintf("Error executing template: %v", err))
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(fmt.Sprintf("Error formatting code: %v", err))
	}
	
	if config.Output != "" {
		// Ensure the output file exists (create or truncate)
		f, err := os.OpenFile(config.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(fmt.Sprintf("Error opening output file %s: %v", config.Output, err))
		}
		defer f.Close()

		// Write the formatted code to the file
		_, err = f.Write(formatted)
		if err != nil {
			panic(fmt.Sprintf("Error writing to output file %s: %v", config.Output, err))
		}
		fmt.Fprintf(os.Stdout, "Wrote output to \"%s\"\n", config.Output)
	} else {
		// Print to stdout
		_, err = os.Stdout.Write(formatted)
		if err != nil {
			panic(fmt.Sprintf("Error writing to stdout: %v", err))
		}
	}
	return
}

var (
	ArgsFlagsHelper = map[string]string{
		"--help":      "Show this help message",
		"-p|--pkg":    "Package name for the generated code (default: \"main\")",
		"-o|--output": "Output directory for processed files",
	}
)

func parseArgs() *Options {

	config := &Options{
		PackageName: "main",
		Output:      "./goxorstrings.go",
		Files:       []string{},
	}

	flag.Usage = func() {
		fmt.Println("Usage: goprotos7 [options] <input dir[ dir ...])>")
		flag.PrintDefaults()
	}

	flag.StringVar(&config.PackageName, "p", config.PackageName, "Package name for the generated code")
	flag.StringVar(&config.PackageName, "pkg", config.PackageName, "Package name for the generated code")
	flag.StringVar(&config.Output, "o", "", "Output directory for processed files")
	flag.StringVar(&config.Output, "output", "", "Output directory for processed files")

	flag.Parse()

	// Make sure we have input paths.
	if flag.NArg() == 0 {
		_, _ = fmt.Fprintf(os.Stdout, "Missing <input dir>\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create input configurations.
	inputDirectories := []string{}
	for i := range flag.NArg() {
		inputDirectories = append(inputDirectories, flag.Arg(i))
	}
	config.Files = inputDirectories

	return config
}

var regFuncName = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// safeVarName converts the given name into a name
// which qualifies as a valid variable identifier. It
// also compares against a known list of variables to
// prevent conflict based on name translation.
func safeVarName(name string, knownVars map[string]int) string {
	var inBytes, outBytes []byte
	var toUpper bool

	name = strings.ToLower(name)
	inBytes = []byte(name)

	for i := 0; i < len(inBytes); i++ {
		if regFuncName.Match([]byte{inBytes[i]}) {
			toUpper = true
		} else if toUpper {
			outBytes = append(outBytes, []byte(strings.ToUpper(string(inBytes[i])))...)
			toUpper = false
		} else {
			outBytes = append(outBytes, inBytes[i])
		}
	}

	name = string(outBytes)

	// Identifier can't start with a digit.
	if unicode.IsDigit(rune(name[0])) {
		name = "_" + name
	}

	if num, ok := knownVars[name]; ok {
		knownVars[name] = num + 1
		name = fmt.Sprintf("%s%d", name, num)
	} else {
		knownVars[name] = 2
	}

	return name
}
