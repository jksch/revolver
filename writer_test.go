package revolver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

const (
	testMiddlePart = "log_file"
)

var (
	testMiddlePartFunc = func() string {
		return testMiddlePart
	}
)

func TestNew(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		conf   Conf
		err    string
	}{
		{
			conf: Conf{},
			err:  "revolver conf can not be empty",
		},
		{
			before: func(t *testing.T) {
				_, err := os.Create("test")
				logErr(err, t)
			},
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 1024,
			},
			err: "revolver setup, test is not a directory",
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0000), t)
			},
			after: func(t *testing.T) {
				logErr(os.Chmod("test", 0755), t)
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 1,
				MaxBytes: 1024,
			},
			err: "revolver remove, open test: permission denied",
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0555), t)
			},
			after: func(t *testing.T) {
				logErr(os.Chmod("test", 0755), t)
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 1024,
			},
			err: "revolver create, open " + filepath.FromSlash("test/log_"+testMiddlePart) + ": permission denied",
		},
		{
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 1024,
			},
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for file := 0; file < 3; file++ {
					_, err := os.Create(filepath.FromSlash("test/log_" + strconv.Itoa(file)))
					logErr(err, t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 1,
				MaxBytes: 1024,
			},
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. New err: %v", index, test.err), func(t *testing.T) {
			if test.before != nil {
				test.before(t)
			}
			if test.after != nil {
				defer test.after(t)
			}

			_, err := New(test.conf)
			if errStr(err) != test.err {
				t.Fatalf("%d. exp err: '%s' got: '%v'", index, test.err, err)
			}

			if test.err != "" {
				return
			}

			files, err := ioutil.ReadDir("test")
			logErrAt(err, index, t)
			count := len(files)
			if count > test.conf.MaxFiles {
				t.Errorf("%d. exp file count: %d got: %d", index, test.conf.MaxFiles, count)
			}

		})
	}
}

func TestWrite(t *testing.T) {
	var tests = []struct {
		before func(w *revWriter, t *testing.T)
		after  func(t *testing.T)
		conf   Conf
		bytes  []byte
		err    string
	}{
		{
			before: func(w *revWriter, t *testing.T) {},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 5,
			},
			bytes: []byte{1, 2, 3, 4, 5, 6},
			err:   "revolver bytes to write 6 over max file size 5",
		},
		{
			before: func(w *revWriter, t *testing.T) {
				logErr(os.Chmod("test", 0000), t)
				w.size = 5
			},
			after: func(t *testing.T) {
				logErr(os.Chmod("test", 0755), t)
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 10,
			},
			bytes: []byte{1, 2, 3, 4, 5, 6},
			err:   "revolver remove, open test: permission denied",
		},
		{
			before: func(w *revWriter, t *testing.T) {
				_, err := os.Create(filepath.FromSlash("test/log_test"))
				logErr(err, t)
				logErr(os.Chmod("test", 0555), t)
				w.size = 5
			},
			after: func(t *testing.T) {
				logErr(os.Chmod("test", 0755), t)
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 1,
				MaxBytes: 10,
			},
			bytes: []byte{1, 2, 3, 4, 5, 6},
			err:   "revolver remove, remove " + filepath.FromSlash("test/log_"+testMiddlePart) + ": permission denied",
		},
		{
			before: func(w *revWriter, t *testing.T) {
				_, err := os.Create(filepath.FromSlash("test/log_test"))
				logErr(err, t)
				w.size = 5
				logErr(w.file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 1,
				MaxBytes: 10,
			},
			bytes: []byte{1, 2, 3, 4, 5, 6},
			err:   "revolver close, invalid argument",
		},
		{
			before: func(w *revWriter, t *testing.T) {
				w.size = 5
				logErr(os.Chmod("test", 0555), t)
			},
			after: func(t *testing.T) {
				logErr(os.Chmod("test", 0755), t)
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 10,
			},
			bytes: []byte{1, 2, 3, 4, 5, 6},
			err:   "revolver create, open " + filepath.FromSlash("test/log_"+testMiddlePart+"_0") + ": permission denied",
		},
		{
			before: func(w *revWriter, t *testing.T) {
				w.size = 5
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 10,
			},
			bytes: []byte("This..."),
		},
		{
			before: func(w *revWriter, t *testing.T) {
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 7,
			},
			bytes: []byte("This..."),
		},
		{
			before: func(w *revWriter, t *testing.T) {
				for file := 0; file < 3; file++ {
					_, err := os.Create("test/log_" + testMiddlePart + "_" + strconv.Itoa(file) + ".txt")
					logErr(err, t)
				}
				w.size = 8
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				Suffix:   ".txt",
				MaxFiles: 1,
				MaxBytes: 9,
			},
			bytes: []byte{0, 1, 2, 3, 4, 5, 6, 7},
		},
	}
	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. Write err: %v", index, test.err), func(t *testing.T) {
			w, err := New(test.conf)
			logErrAt(err, index, t)
			defer w.Close()

			log := w.(*revWriter)
			test.before(log, t)
			defer test.after(t)

			n, err := log.Write(test.bytes)
			if errStr(err) != test.err {
				t.Errorf("%d. exp err: '%s' got: '%v'", index, test.err, err)
			}
			if test.err != "" {
				return // test done
			}

			count, err := fileCount(test.conf)
			logErrAt(err, index, t)
			if count > test.conf.MaxFiles {
				t.Errorf("%d. exp file count: %d got: %d", index, test.conf.MaxFiles, count)
			}

			length := len(test.bytes)
			if n != length {
				t.Errorf("%d. exp to write %d bytes got %d", index, length, n)
			}
			got, err := ioutil.ReadFile(log.file.Name())
			logErrAt(err, index, t)
			if !bytes.Equal(test.bytes, got) {
				t.Errorf("%d. exp content: '%v' got: '%v'", index, test.bytes, got)
			}
		})
	}
}

func TestCloseTwice(t *testing.T) {
	defer func() {
		logErr(os.RemoveAll("test"), t)
	}()
	conf := Conf{
		Dir:      "test",
		Prefix:   "log_",
		Middle:   testMiddlePartFunc,
		MaxFiles: 2,
		MaxBytes: 1024,
	}
	w, err := New(conf)
	logErr(err, t)
	logErr(w.Close(), t)
	if err := w.Close(); err != nil {
		t.Errorf("expected close to return nil but got: '%v'", err)
	}
}

func TestMust(t *testing.T) {
	var tests = []struct {
		conf  Conf
		panic bool
	}{
		{
			conf:  Conf{},
			panic: true,
		},
		{
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				Middle:   testMiddlePartFunc,
				MaxFiles: 2,
				MaxBytes: 1024,
			},
			panic: false,
		},
	}
	for index, test := range tests {
		t.Run(fmt.Sprintf("%d. must panic: %v", index, test.panic), func(t *testing.T) {
			defer func() {
				got := recover() != nil
				if test.panic != got {
					t.Errorf("%d. exp panic: %v got: %v", index, test.panic, got)
				}
				if test.panic {
					return
				}
				logErrAt(os.RemoveAll("test"), index, t)
			}()
			w := Must(New(test.conf))
			if test.panic {
				t.Errorf("%d. exp panic: true  got: false", index, test.panic)
			}
			if w != nil {
				w.Close()
			}
		})
	}
}

func TestRace(t *testing.T) {
	defer func() {
		logErr(os.RemoveAll("test"), t)
	}()
	conf := Conf{
		Dir:      "test",
		Prefix:   "log_",
		Middle:   testMiddlePartFunc,
		Suffix:   ".txt",
		MaxFiles: 2,
		MaxBytes: 1024,
	}
	w, err := New(conf)
	logErr(err, t)

	wg := sync.WaitGroup{}
	runner := func(runner int, wg *sync.WaitGroup) {
		for mes := 0; mes < 4; mes++ {
			fmt.Fprintf(w, "Runner %d, log %d", runner, mes)
		}
		wg.Done()
	}
	wg.Add(4)
	for worker := 0; worker < 4; worker++ {
		worker := worker
		go runner(worker, &wg)
	}
	wg.Wait()
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func logErr(err error, t *testing.T) {
	if err != nil {
		t.Fatalf("unexpected error, %+v", err)
	}
}

func logErrAt(err error, index int, t *testing.T) {
	if err != nil {
		t.Fatalf("%d. unexpected error, %+v", index, err)
	}
}
