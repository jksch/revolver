package revolver

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	exp := Conf{
		Dir:      defaultDir,
		Prefix:   defaultPrefix,
		Middle:   logDate,
		Suffix:   defaultSuffix,
		MaxFiles: defaultMaxFiles,
		MaxBytes: defaultMaxBytes,
	}
	got := DefaultConf()
	if exp.Dir != got.Dir {
		t.Errorf("exp config.Middle: %v got: %v", exp.Dir, got.Dir)
	}
	if exp.Prefix != got.Prefix {
		t.Errorf("exp config.Prefix: %v got: %v", exp.Prefix, got.Prefix)
	}
	if reflect.DeepEqual(exp.Middle, got.Middle) {
		t.Errorf("exp config.Middle: %v got: %v", exp.Middle, got.Middle)
	}
	if exp.Suffix != got.Suffix {
		t.Errorf("exp config.Suffix: %v got: %v", exp.Suffix, got.Suffix)
	}
	if exp.MaxFiles != got.MaxFiles {
		t.Errorf("exp config.MaxFiles: %v got: %v", exp.MaxFiles, got.MaxFiles)
	}
	if exp.MaxBytes != got.MaxBytes {
		t.Errorf("exp config.MaxBytes: %v got: %v", exp.MaxBytes, got.MaxBytes)
	}
}

func TestValidConf(t *testing.T) {
	var tests = []struct {
		conf Conf
		err  string
	}{
		{
			conf: Conf{},
			err:  "conf can not be empty",
		},
		{
			conf: Conf{Dir: "log/"},
			err:  "conf.Prefix can not be empty",
		},
		{
			conf: Conf{
				Dir:    "log/",
				Prefix: "log-",
			},
			err: "conf.Middle can not be nil",
		},
		{
			conf: Conf{
				Dir:    "log/",
				Prefix: "log-",
				Middle: logDate,
			},
			err: "conf.MaxFiles must be > 0",
		},
		{
			conf: Conf{
				Dir:      "log/",
				Prefix:   "log-",
				Middle:   logDate,
				MaxFiles: 1,
			},
			err: "conf.MaxBytes must be > 0",
		},
		{
			conf: Conf{
				Dir:      "log/",
				Prefix:   "log-",
				Middle:   logDate,
				MaxFiles: 1,
				MaxBytes: 1,
			},
			err: "",
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. valid config err: %v", index, test.err), func(t *testing.T) {
			if err := errStr(ValidConf(test.conf)); err != test.err {
				t.Fatalf("%d. exp err: '%s' got: '%v'", index, test.err, err)
			}

		})
	}
}

func TestClean(t *testing.T) {
	var tests = []struct {
		conf Conf
		exp  string
	}{
		{conf: Conf{Dir: ""}, exp: "."},
		{conf: Conf{Dir: "."}, exp: "."},
		{conf: Conf{Dir: "test"}, exp: "test"},
		{conf: Conf{Dir: "test/"}, exp: "test"},
		{conf: Conf{Dir: "./test"}, exp: "test"},
		{conf: Conf{Dir: "./test/log/"}, exp: "test/log"},
		{conf: Conf{Dir: "/test/log/"}, exp: "/test/log"},
	}
	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. dir: '%s' to '%s'", index, test.conf.Dir, test.exp), func(t *testing.T) {
			t.Parallel()
			conf := clean(test.conf)
			if conf.Dir != test.exp {
				t.Errorf("%d. exp dir: '%s' got: '%s'", index, conf.Dir, test.exp)
			}
		})
	}
}

func TestLogDate(t *testing.T) {
	formatOnly := time.Now().Format("02-01-2006-15:04:05")
	replaced := logDate()
	if formatOnly == replaced {
		t.Errorf("exp logDate: '%s' !=  '%s'", replaced, formatOnly)
	}
}
