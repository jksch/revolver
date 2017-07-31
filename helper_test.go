package revolver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSetupDirs(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		dirs   string
		err    string
	}{
		{
			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},
			dirs:   "",
			err:    "mkdir : no such file or directory",
		},
		{
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dirs: ".",
			err:  "",
		},
		{
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dirs: "./test",
			err:  "",
		},
		{
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			dirs: "test",
			err:  "",
		},
		{
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dirs: "test/log",
			err:  "",
		},
		{
			before: func(t *testing.T) {
				file, err := os.Create("test")
				logErr(err, t)
				logErr(file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dirs: "test/log",
			err:  "stat " + filepath.FromSlash("test/log") + ": not a directory",
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. exp dir: %s err: %v", index, test.dirs, test.err), func(t *testing.T) {
			test.before(t)
			defer test.after(t)

			if err := errStr(setupDirs(test.dirs)); err != test.err {
				t.Fatalf("%d. exp err: '%s' got: '%v'", index, test.err, err)
			}
			if test.err != "" || test.dirs == "" {
				return // done testing
			}

			info, err := os.Stat(test.dirs)
			logErrAt(err, index, t)
			if !info.IsDir() {
				t.Errorf("%d. %s schould be dir", index, test.dirs)
			}

		})
	}
}

func TestNewCreateFile(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		dir    string
		suffix string
		prefix string
		middle func() string
		expSuf string
		err    string
	}{
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:    "test",
			prefix: "dd_",
			suffix: ".json",
			middle: testMiddlePartFunc,
			expSuf: ".json",
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				file, err := os.Create(filepath.FromSlash("test/dd_" + testMiddlePart + ".json"))
				logErr(err, t)
				logErr(file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:    "test",
			prefix: "dd_",
			suffix: ".json",
			middle: testMiddlePartFunc,
			expSuf: "_0.json",
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				file, err := os.Create(filepath.FromSlash("test/dd_" + testMiddlePart + ".json"))
				logErr(err, t)
				logErr(file.Close(), t)
				file, err = os.Create(filepath.FromSlash("test/dd_" + testMiddlePart + "_0" + ".json"))
				logErr(err, t)
				logErr(file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:    "test",
			prefix: "dd_",
			suffix: ".json",
			middle: testMiddlePartFunc,
			expSuf: "_1.json",
		},
		{
			before: func(t *testing.T) {
				file, err := os.Create("test")
				logErr(err, t)
				logErr(file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			dir:    "test",
			prefix: "dd_",
			suffix: ".json",
			middle: testMiddlePartFunc,
			err:    "stat " + filepath.FromSlash("test/dd_"+testMiddlePart+".json") + ": not a directory",
		},
	}
	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. create file", index), func(t *testing.T) {
			test.before(t)
			defer test.after(t)

			file, err := createFile(test.dir, test.prefix, test.suffix, test.middle)
			if test.err != "" && test.err != errStr(err) {
				t.Errorf("%d. exp err: '%s' got: '%v'", index, test.err, err)
			}
			if test.err != "" {
				return // test done
			}
			name := file.Name()
			prefix := filepath.FromSlash(test.dir + "/" + test.prefix)
			if !strings.HasPrefix(name, prefix) {
				t.Errorf("%d. name '%s' should have prefix '%s'", index, name, test.prefix)
			}
			if !strings.Contains(name, testMiddlePart) {
				t.Errorf("%d. name '%s' should contain date in format '%s'", index, name, testMiddlePart)
			}
			if !strings.HasSuffix(name, test.expSuf) {
				t.Errorf("%d. name '%s' should have suffix '%s'", index, name, test.expSuf)
			}
		})
	}
}

func TestNewFileCount(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		dir    string
		prefix string
		count  int
		err    string
	}{
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
			},
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			dir:    "test",
			prefix: "log_",
			count:  0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				file, err := os.Create(filepath.FromSlash("test/log_1"))
				logErr(err, t)
				logErr(file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:    "test",
			prefix: "log_",
			count:  1,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for files := 0; files < 4; files++ {
					file, err := os.Create(filepath.FromSlash(fmt.Sprintf("test/log_%d", files)))
					logErr(err, t)
					logErr(file.Close(), t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:    "test",
			prefix: "log_",
			count:  4,
		},
		{
			before: func(t *testing.T) {
				file, err := os.Create("test")
				logErr(err, t)
				logErr(file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			dir:    "test",
			prefix: "log_",
			err:    "readdirent:",
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. count exp %d err %v", index, test.count, test.err), func(t *testing.T) {
			test.before(t)
			defer test.after(t)

			count, err := fileCount(test.dir, test.prefix)
			if test.err != "" && !strings.HasPrefix(errStr(err), test.err) {
				t.Errorf("%d. exp err starts: '%s' got: '%v'", index, test.err, err)
			}
			if count != test.count {
				t.Errorf("%d. exp count: %d got %d", index, test.count, count)
			}
		})
	}
}

func TestNewRemoveOlderst(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		files  []string
		dir    string
		prefix string
		count  int
		err    string
	}{
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
			},
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			dir:    "test",
			prefix: "_",
			count:  0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for files := 0; files < 1; files++ {
					file, err := os.Create(filepath.FromSlash(fmt.Sprintf("test/_%d", files)))
					logErr(err, t)
					logErr(file.Close(), t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:    "test",
			prefix: "_",
			count:  0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for files := 0; files < 3; files++ {
					file, err := os.Create(filepath.FromSlash(fmt.Sprintf("test/_%d", files)))
					logErr(err, t)
					logErr(file.Close(), t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:    "test",
			prefix: "_",
			files:  []string{"_1", "_2"},
			count:  2,
		},
		{
			before: func(t *testing.T) {
				file, err := os.Create("test")
				logErr(err, t)
				logErr(file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			dir:    "test",
			prefix: "_",
			err:    "readdirent: not a directory",
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				file, err := os.Create(filepath.FromSlash("test/_log"))
				logErr(err, t)
				logErr(file.Close(), t)
				logErr(os.Chmod("test", 0544), t)
			},
			after: func(t *testing.T) {
				logErr(os.Chmod("test", 0755), t)
				logErr(os.RemoveAll("test"), t)
			},
			dir:    "test",
			prefix: "_",
			err:    "remove " + filepath.FromSlash("test/_log") + ": permission denied",
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. remove oldest", index), func(t *testing.T) {
			test.before(t)
			defer test.after(t)

			if err := errStr(removeOldestFile(test.dir, test.prefix)); err != test.err {
				t.Errorf("%d. exp err: '%s' got: '%s'", index, test.err, err)
			}
			if test.err != "" {
				return //Test done
			}
			files, err := ioutil.ReadDir(test.dir)
			logErrAt(err, index, t)
			for position, name := range test.files {
				if !containsFileName(name, files) {
					t.Errorf("%d. exp file: %s at %d to remain in folder", index, name, position)
				}
			}
			count := len(files)
			if count != test.count {
				t.Errorf("%d. exp count: %d got: %d", index, test.count, count)
			}
		})
	}
}

func TestNewCountAndRemove(t *testing.T) {
	var tests = []struct {
		before   func(t *testing.T)
		after    func(t *testing.T)
		dir      string
		prefix   string
		maxFiles int
		count    int
		err      string
	}{
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
			},
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			dir:      "test",
			prefix:   "log_",
			maxFiles: 1,
			count:    0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				file, err := os.Create(filepath.FromSlash("test/log_" + testMiddlePart))
				logErr(err, t)
				logErr(file.Close(), t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:      "test",
			prefix:   "log_",
			maxFiles: 1,
			count:    0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for file := 0; file < 3; file++ {
					file, err := os.Create(filepath.FromSlash("test/log_" + testMiddlePart + "_" + strconv.Itoa(file)))
					logErr(err, t)
					logErr(file.Close(), t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dir:      "test",
			prefix:   "log_",
			maxFiles: 2,
			count:    1,
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. count and remove", index), func(t *testing.T) {
			test.before(t)
			defer test.after(t)

			if err := errStr(countAndRemoveFiles(test.dir, test.prefix, test.maxFiles)); err != test.err {
				t.Errorf("%d. exp err: '%s' got: '%s'", index, test.err, err)
			}
			if test.err != "" {
				return //Test done
			}
			files, err := ioutil.ReadDir(test.dir)
			logErrAt(err, index, t)

			count := len(files)
			if count != test.count {
				t.Errorf("%d. exp count: %d got: %d", index, test.count, count)
			}
		})
	}
}

func BenchmarkSetupDirs(b *testing.B) {
	defer func() {
		logBenchmarkErr(os.RemoveAll("test"), b)
	}()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		logBenchmarkErr(setupDirs("test/log"), b)
		b.StopTimer()
		logBenchmarkErr(os.RemoveAll("test"), b)
	}
}

func BenchmarkCreateFile(b *testing.B) {
	defer func() {
		logBenchmarkErr(os.RemoveAll("test"), b)
	}()
	dir := "test"
	prefix := "log_"
	suffix := ".txt"
	middle := func() string {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	logBenchmarkErr(os.Mkdir("test", 0755), b)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		file, err := createFile(dir, prefix, suffix, middle)
		b.StopTimer()
		logBenchmarkErr(err, b)
		logBenchmarkErr(file.Close(), b)
	}
}

func BenchmarkFileCount(b *testing.B) {
	defer func() {
		logBenchmarkErr(os.RemoveAll("test"), b)
	}()
	dir := "test"
	prefix := "log_"
	logBenchmarkErr(os.Mkdir("test", 0755), b)
	for file := 0; file < 3; file++ {
		file, err := os.Create(filepath.FromSlash("test/log_file_" + strconv.Itoa(file)))
		logBenchmarkErr(err, b)
		logBenchmarkErr(file.Close(), b)
	}
	for i := 0; i < b.N; i++ {
		_, err := fileCount(dir, prefix)
		logBenchmarkErr(err, b)
	}
}

// Testing only name creation without io
// bytes.Buffer  10000000  154 ns/op
// fmt.Sprintf   3000000   440 ns/o
func BenchmarkNameCreationOnly(b *testing.B) {
	dir := "test"
	prefix := "log_"
	suffix := ".txt"
	middle := func() string { return "file" }
	doStuff := func(name string) { /* do nothing */ }
	for i := 0; i < b.N; i++ {
		name := filepath.FromSlash(dir + "/" + prefix + middle())
		doStuff(name)
		_ = name + suffix
	}
}

func containsFileName(name string, files []os.FileInfo) bool {
	for _, info := range files {
		if info.Name() == name {
			return true
		}
	}
	return false
}
