# Revolver
[![Build Status](https://travis-ci.org/jksch/revolver.svg?branch=master)](https://travis-ci.org/jksch/revolver)
[![Coverage Status](https://coveralls.io/repos/github/jksch/revolver/badge.svg)](https://coveralls.io/github/jksch/revolver)

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
	w, err := revolver.New(revolver.DefaultConf())
	if err != nil {
		log.Println(err)
	}
	log.SetOutput(w)
	log.Println("Ready to use!")
}
```
Alternatively must can be used to get a writer or panic:
```go
var (
	rev    = revolver.Must(revolver.New(revolver.DefaultConf()))
	logger = log.New(rev, "", log.Ldate|log.Ltime|log.Lshortfile)
)

func main() {
	logger.Println("Ready to use!")
}
```
It is also possible to set up multiple writers for different purposes. Perhaps to separate the app log form the serial log which tends to be spamy:
```go
var (
	appConf = revolver.Conf{
		Dir:    "log/app",
		Prefix: "app-",
		Middle: func() string {
			return strings.Replace(time.Now().Format("02-01-2006-15:04"), ":", "_", -1)
		},
		Suffix:   ".txt",
		MaxFiles: 5,
		MaxBytes: 1024 * 1024 * 10,
	}
	serialConf = revolver.Conf{
		Dir:    "log/serial",
		Prefix: "serial-",
		Middle: func() string {
			return strings.Replace(time.Now().Format("02-01-2006-15:04:00"), ":", "_", -1)
		},
		Suffix:   ".txt",
		MaxFiles: 3,
		MaxBytes: 1024 * 1024 * 5,
	}
	applog    = log.New(io.MultiWriter(os.Stdout, revolver.Must(revolver.New(appConf))), "", 0)
	seriallog = log.New(revolver.Must(revolver.New(serialConf)), "", 0)
)

func main() {
	setupApp(applog)
	setupSerial(seriallog)
}

func setupApp(logger *log.Logger) {
	logger.Printf("Setting up app...")
}

func setupSerial(logger *log.Logger) {
	logger.Printf("Setting up serial com...")
}
```
### Compatibility
Revolver is tested on Linux and Mac. On Windows the package seems to work. However the tests won't pass and since the returned errors are windows language specific there is no point in fixing them.
