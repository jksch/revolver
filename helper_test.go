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
			dirs: "",
			err:  "mkdir : no such file or directory",
		},
		{
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dirs: ".",
			err:  "",
		},
		{
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dirs: "./test",
			err:  "",
		},
		{
			after: func(t *testing.T) {
				logErr(os.Remove("test"), t)
			},
			dirs: "test",
			err:  "",
		},
		{
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			dirs: "test/log",
			err:  "",
		},
		{
			before: func(t *testing.T) {
				_, err := os.Create("test")
				logErr(err, t)
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
			if test.before != nil {
				test.before(t)
			}
			if test.after != nil {
				defer test.after(t)
			}
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

func TestCreateFile(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		conf   Conf
		suffix string
		err    string
	}{
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:    "test",
				Prefix: "dd_",
				Suffix: ".json",
				Middle: testMiddlePartFunc,
			},
			suffix: ".json",
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				_, err := os.Create(filepath.FromSlash("test/dd_" + testMiddlePart + ".json"))
				logErr(err, t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:    "test",
				Prefix: "dd_",
				Suffix: ".json",
				Middle: testMiddlePartFunc,
			},
			suffix: "_0.json",
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				_, err := os.Create(filepath.FromSlash("test/dd_" + testMiddlePart + ".json"))
				logErr(err, t)
				_, err = os.Create(filepath.FromSlash("test/dd_" + testMiddlePart + "_0" + ".json"))
				logErr(err, t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:    "test",
				Prefix: "dd_",
				Suffix: ".json",
				Middle: testMiddlePartFunc,
			},
			suffix: "_1.json",
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
				Dir:    "test",
				Prefix: "dd_",
				Suffix: ".json",
				Middle: testMiddlePartFunc,
			},
			err: "stat " + filepath.FromSlash("test/dd_"+testMiddlePart+".json") + ": not a directory",
		},
	}
	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. create file", index), func(t *testing.T) {
			if test.before != nil {
				test.before(t)
			}
			if test.after != nil {
				defer test.after(t)
			}
			file, err := createFile(test.conf)
			if test.err != "" && test.err != errStr(err) {
				t.Errorf("%d. exp err: '%s' got: '%v'", index, test.err, err)
			}
			if test.err != "" {
				return // test done
			}
			name := file.Name()
			prefix := filepath.FromSlash(test.conf.Dir + "/" + test.conf.Prefix)
			if !strings.HasPrefix(name, prefix) {
				t.Errorf("%d. name '%s' should have prefix '%s'", index, name, test.conf.Prefix)
			}
			if !strings.Contains(name, testMiddlePart) {
				t.Errorf("%d. name '%s' should contain date in format '%s'", index, name, testMiddlePart)
			}
			if !strings.HasSuffix(name, test.suffix) {
				t.Errorf("%d. name '%s' should have suffix '%s'", index, name, test.suffix)
			}
		})
	}
}

func TestFileCount(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		conf   Conf
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
			conf: Conf{
				Dir:    "test",
				Prefix: "log_",
			},
			count: 0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				_, err := os.Create(filepath.FromSlash("test/log_1"))
				logErr(err, t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:    "test",
				Prefix: "log_",
			},
			count: 1,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for files := 0; files < 4; files++ {
					_, err := os.Create(filepath.FromSlash(fmt.Sprintf("test/log_%d", files)))
					logErr(err, t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:    "test",
				Prefix: "log_",
			},
			count: 4,
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
				Dir:    "test",
				Prefix: "log_",
			},
			err: "readdirent:",
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. count exp %d err %v", index, test.count, test.err), func(t *testing.T) {
			if test.before != nil {
				test.before(t)
			}
			if test.after != nil {
				defer test.after(t)
			}
			count, err := fileCount(test.conf)
			if test.err != "" && !strings.HasPrefix(errStr(err), test.err) {
				t.Errorf("%d. exp err starts: '%s' got: '%v'", index, test.err, err)
			}
			if count != test.count {
				t.Errorf("%d. exp count: %d got %d", index, test.count, count)
			}
		})
	}
}

func TestRemoveOlderst(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		conf   Conf
		files  []string
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
			conf: Conf{
				Dir:    "test",
				Prefix: "_",
			},
			count: 0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for files := 0; files < 1; files++ {
					_, err := os.Create(filepath.FromSlash(fmt.Sprintf("test/_%d", files)))
					logErr(err, t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:    "test",
				Prefix: "_",
			},
			count: 0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for files := 0; files < 3; files++ {
					_, err := os.Create(filepath.FromSlash(fmt.Sprintf("test/_%d", files)))
					logErr(err, t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:    "test",
				Prefix: "_",
			},
			files: []string{"_1", "_2"},
			count: 2,
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
				Dir:    "test",
				Prefix: "_",
			},
			err: "open " + filepath.FromSlash("test/") + ": not a directory",
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				_, err := os.Create(filepath.FromSlash("test/_log"))
				logErr(err, t)
				logErr(os.Chmod("test", 0544), t)
			},
			after: func(t *testing.T) {
				logErr(os.Chmod("test", 0755), t)
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:    "test",
				Prefix: "_",
			},
			err: "remove " + filepath.FromSlash("test/_log") + ": permission denied",
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. remove oldest", index), func(t *testing.T) {
			if test.before != nil {
				test.before(t)
			}
			if test.after != nil {
				defer test.after(t)
			}
			if err := errStr(removeOldestFile(test.conf)); err != test.err {
				t.Errorf("%d. exp err: '%s' got: '%s'", index, test.err, err)
			}
			if test.err != "" {
				return //Test done
			}
			files, err := ioutil.ReadDir(test.conf.Dir)
			logErrAt(err, index, t)
		check:
			for position, name := range test.files {
				for _, info := range files {
					if info.Name() == name {
						continue check
					}
				}
				t.Errorf("%d. exp file: %s at %d to remain in folder", index, name, position)
			}
			count := len(files)
			if count != test.count {
				t.Errorf("%d. exp count: %d got: %d", index, test.count, count)
			}
		})
	}
}

func TestCountAndRemove(t *testing.T) {
	var tests = []struct {
		before func(t *testing.T)
		after  func(t *testing.T)
		conf   Conf
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
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				MaxFiles: 1,
			},
			count: 0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				_, err := os.Create(filepath.FromSlash("test/log_" + testMiddlePart))
				logErr(err, t)
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				MaxFiles: 1,
			},
			count: 0,
		},
		{
			before: func(t *testing.T) {
				logErr(os.Mkdir("test", 0755), t)
				for file := 0; file < 3; file++ {
					_, err := os.Create(filepath.FromSlash("test/log_" + testMiddlePart + "_" + strconv.Itoa(file)))
					logErr(err, t)
				}
			},
			after: func(t *testing.T) {
				logErr(os.RemoveAll("test"), t)
			},
			conf: Conf{
				Dir:      "test",
				Prefix:   "log_",
				MaxFiles: 2,
			},
			count: 1,
		},
	}

	for index, test := range tests {
		index, test := index, test
		t.Run(fmt.Sprintf("%d. count and remove", index), func(t *testing.T) {
			if test.before != nil {
				test.before(t)
			}
			if test.after != nil {
				defer test.after(t)
			}
			if err := errStr(countAndRemoveFiles(test.conf)); err != test.err {
				t.Errorf("%d. exp err: '%s' got: '%s'", index, test.err, err)
			}
			if test.err != "" {
				return //Test done
			}
			files, err := ioutil.ReadDir(test.conf.Dir)
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
		if err := os.RemoveAll("test"); err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
	}()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		if err := setupDirs("test/log"); err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
		b.StopTimer()
		if err := os.RemoveAll("test"); err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
	}
}

func BenchmarkCreateFile(b *testing.B) {
	defer func() {
		if err := os.RemoveAll("test"); err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
	}()
	var conf = Conf{
		Dir:    "test",
		Prefix: "log_",
		Suffix: ".txt",
		Middle: func() string {
			return strconv.FormatInt(time.Now().UnixNano(), 10)
		},
	}
	if err := os.Mkdir("test", 0755); err != nil {
		b.Fatalf("unexpected error, %v", err)
	}
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		file, err := createFile(conf)
		if err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
		b.StopTimer()
		if err := file.Close(); err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
	}
}

func BenchmarkFileCount(b *testing.B) {
	defer func() {
		if err := os.RemoveAll("test"); err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
	}()
	var conf = Conf{
		Dir:    "test",
		Prefix: "log_",
	}
	if err := os.Mkdir("test", 0755); err != nil {
		b.Fatalf("unexpected error, %v", err)
	}
	for file := 0; file < 3; file++ {
		file, err := os.Create(filepath.FromSlash("test/log_file_" + strconv.Itoa(file)))
		if err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
		if err := file.Close(); err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
	}
	for i := 0; i < b.N; i++ {
		if _, err := fileCount(conf); err != nil {
			b.Fatalf("unexpected error, %v", err)
		}
	}
}
