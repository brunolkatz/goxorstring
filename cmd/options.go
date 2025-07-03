package main

type Options struct {
	PackageName string   `json:"package_name"` // Package name for the generated code
	Output      string   `json:"output"`       // Optional output name to save processed files, default is current directory if empty with the name goxorstrings.go
	Files       []string `json:"files"`        // List of files to process
}
