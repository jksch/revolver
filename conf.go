package revolver

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

const (
	defaultDir      = "log"
	defaultPrefix   = "log-"
	defaultSuffix   = ".txt"
	defaultMaxFiles = 3
	defaultMaxBytes = 1024 * 1024 * 10
)

// DateStringMiddle returns a date string which dose not contain any reserved
// characters (e. g. i':' on Windows). Therefor it is save to used in filenames.
func DateStringMiddle() string {
	return strings.Replace(time.Now().Format("02-01-2006-15:04:05"), ":", "_", -1)
}

// Conf is deprecated. Use NewQuick instead.
// Conf holds the conf for the revolving file writer.
type Conf struct {
	Dir      string        // CAUTION all files in this dir with the Prefix will eventually be deleted
	Prefix   string        // CAUTION this is used to identify surplus files to delete
	Middle   func() string // A function that returns the middle of the file name part e. g. a date
	Suffix   string        // optional
	MaxFiles int           // min 1
	MaxBytes int           // min 1
}

// DefaultConf returns a ready to use revolver conf.
// Dir: log
// Prefix: log-
// Middle: Now().format("02-01-2006-15:04:05") replaced : with _
// Suffix: .txt
// MaxFiles: 3
// MaxBytes: 1024 * 1024 * 10
func DefaultConf() Conf {
	return Conf{
		Dir:      defaultDir,
		Prefix:   defaultPrefix,
		Middle:   DateStringMiddle,
		Suffix:   defaultSuffix,
		MaxFiles: defaultMaxFiles,
		MaxBytes: defaultMaxBytes,
	}
}

// ValidConf is deprecated. Use NewQuick instead.
// ValidConf checks if the given conf is valid.
// Calling revolver.New() will also validate the conf.
func ValidConf(conf Conf) error {
	switch {
	case reflect.DeepEqual(conf, Conf{}):
		return fmt.Errorf("revolver conf can not be empty")
	case conf.Prefix == "":
		return fmt.Errorf("revolver conf.Prefix can not be empty")
	case conf.Middle == nil:
		return fmt.Errorf("revolver conf.Middle can not be nil")
	case conf.MaxFiles < 1:
		return fmt.Errorf("revolver conf.MaxFiles must be > 0")
	case conf.MaxBytes < 1:
		return fmt.Errorf("revolver conf.MaxBytes must be > 0")
	}
	return nil
}

func clean(conf Conf) Conf {
	conf.Dir = filepath.Clean(conf.Dir)
	return conf
}
