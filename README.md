# Revolver
[![Build Status](https://travis-ci.org/jksch/revolver.svg?branch=master)](https://travis-ci.org/jksch/revolver)
[![Coverage Status](https://coveralls.io/repos/github/jksch/revolver/badge.svg)](https://coveralls.io/github/jksch/revolver)
[![Go Report Card](https://goreportcard.com/badge/github.com/jksch/revolver)](https://goreportcard.com/report/github.com/jksch/revolver)
[![GoDoc](https://godoc.org/github.com/jksch/revolver?status.svg)](https://godoc.org/github.com/jksch/revolver)
[![License](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/jksch/revolver/blob/master/LICENSE)

Is a simple revolving file writer for go.

Revolver allows for a simple log file rotation setup:

* Writing dir
* File prefix
* File name
* File suffix
* Max file size
* Max files 

...can be customized.
### Basic usage
```go
func main() {
	w, err := revolver.NewQuick(
		"logs",
		"log_",
		".txt",
		func() string { return "" },
		1024*1024,
		3,
	)
	if err != nil {
		panic(err)
	}
	Log := log.New(w, "", log.Ldate|log.Ltime|log.Lshortfile)
	Log.Printf("Ready to use...")
}
```
Alternatively must can be used to get a writer or panic:
```go
var (
	count  = 0
	writer = revolver.Must(revolver.NewQuick(
		"exports",
		"export_",
		".json",
		func() string {
			count++
			return strconv.Itoa(count)
		},
		1024*1024,
		10,
	))
)

func main() {
	writer.Write([]byte("Some export data..."))
}
```
### Compatibility
Revolver is tested on Linux and Mac. On Windows the package seems to work. However the tests won't pass and since the returned errors are windows language specific there is no point in fixing them.
