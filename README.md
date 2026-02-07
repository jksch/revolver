# Revolver
[![Go Report Card](https://goreportcard.com/badge/github.com/jksch/revolver)](https://goreportcard.com/report/github.com/jksch/revolver)
[![GoDoc](https://godoc.org/github.com/jksch/revolver?status.svg)](https://godoc.org/github.com/jksch/revolver)
[![License](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/jksch/revolver/blob/master/LICENSE)

Is a simple revolving file writer for go.

Revolver allows for a simple log file rotation setup:

* Writing dir
* File prefix
* File middle name
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
		revolver.DateStringMiddle,
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
	count = 0
	w     = revolver.Must(revolver.NewQuick(
		"logs",
		"log_",
		".txt",
		revolver.DateStringMiddle,
		1024*1024,
		3,
	))
	Log = log.New(w, "", log.Ldate|log.Ltime|log.Lshortfile)
)

func main() {
	Log.Printf("Ready to use...")
}
```
A different use case would be:
```go
func main() {
	count := 0
	w, err := revolver.NewQuick(
		"exports",
		"export_",
		".json",
		func() string {
			count++
			return strconv.Itoa(count)
		},
		1024*1024,
		10,
	)
	if err != nil {
		panic(err)
	}
	defer w.Close()
	w.Write([]byte("Some json data..."))
}
```
### Parameters
###### Dir
Specifies the directory to write to. If the directory dose not exist, it and all parents will be created.
###### Prefix
The prefix is mandatory and will be used to determent which files can be deleted.
###### Suffix
In case the next generated filename already exists revolver will append a number to this filename. E. g. the file export.json already exists export.json_1 will be created. But now the file extension would be broken. To remedy that a filename suffix can be specified. All files are guaranteed to end with this suffix.
###### Middle
The middle name of the file can be customized in this function.
###### MaxBytes
Specifies the maximum bytes size a file can have. If the data to be written is larger then the remaining file size, a new file will be created.
###### MaxFiles
This is the limit of files that will be created.
### Compatibility
Revolver is tested on Linux and Mac. On Windows the package seems to work. However the tests won't pass and since the returned errors are windows language specific there is no point in fixing them.
