// Package revolver provides a revolving file writer.
package revolver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type revWriter struct {
	dir      string
	prefix   string
	suffix   string
	middle   func() string
	maxBytes int
	maxFiles int
	size     int
	file     *os.File
	lock     *sync.Mutex // synchronizes file operations
}

// Must wraps the call to NewWriter and returns a io.WriteCloser or panics
func Must(w io.WriteCloser, err error) io.WriteCloser {
	if err != nil {
		panic(fmt.Errorf("could not create revolving log writer, %v", err))
	}
	return w
}

// New  returns a io.WriteCloser that writes revolving files as specified by the given conf.
// Calling New will always create a new file even if there is space left in other files.
// If the configured directory doesn't exist it will be created.
func New(conf Conf) (io.WriteCloser, error) {
	if err := ValidConf(conf); err != nil {
		return nil, err
	}
	conf = clean(conf)
	return NewQuick(conf.Dir, conf.Prefix, conf.Suffix, conf.Middle, conf.MaxBytes, conf.MaxFiles)
}

// NewQuick returns a io.WriteCloser that writes revolving files.
// Calling New will always create a new file even if there is space left in other files.
// If the configured directory doesn't exist it will be created.
func NewQuick(dir, prefix, suffix string, middle func() string, maxBytes, maxFiles int) (io.WriteCloser, error) {
	if prefix == "" {
		return nil, fmt.Errorf("revolver, prefix can not be empty")
	}
	if middle == nil {
		middle = func() string { return "" }
	}
	if maxBytes < 1 {
		return nil, fmt.Errorf("revolver, maxBytes must be > 0")
	}
	if maxFiles < 1 {
		return nil, fmt.Errorf("revolver, maxFiles must be > 0")
	}

	if err := setupDirs(dir); err != nil {
		return nil, fmt.Errorf("revolver setup, %v", err)
	}
	if err := countAndRemoveFiles(dir, prefix, maxFiles); err != nil {
		return nil, fmt.Errorf("revolver, remove, %v", err)
	}

	file, err := createFile(dir, prefix, suffix, middle)
	if err != nil {
		return nil, fmt.Errorf("revolver, create, %v", err)
	}

	return &revWriter{
		dir:      filepath.Clean(dir),
		prefix:   filepath.Clean(prefix),
		suffix:   suffix,
		middle:   middle,
		maxBytes: maxBytes,
		maxFiles: maxFiles,
		file:     file,
		lock:     &sync.Mutex{},
	}, nil
}

// Write the given bytes into the log file specified by the given conf.
// If there is not enough file space left,surplus files will be deleted and a new file will be created.
func (l *revWriter) Write(p []byte) (n int, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	size := len(p)
	if size > l.maxBytes {
		return 0, fmt.Errorf("revolver, bytes to write %d over max file size %d", size, l.maxBytes)
	}
	if l.file == nil || l.size+size > l.maxBytes {
		if err := l.close(); err != nil {
			return 0, fmt.Errorf("revolver, close, %v", err)
		}

		if err := countAndRemoveFiles(l.dir, l.prefix, l.maxFiles); err != nil {
			return 0, fmt.Errorf("revolver, remove, %v", err)
		}

		file, err := createFile(l.dir, l.prefix, l.suffix, l.middle)
		if err != nil {
			return 0, fmt.Errorf("revolver, create, %v", err)

		}
		l.file = file
		l.size = 0
	}

	l.size += size
	return l.file.Write(p)

}

// Close closes the current log file and sets the writer reference to nil.
// If the file reference is nil, the returned err is always be nil.
// Writing to a nil referencing writer cleans up surplus files and creates a new file.
func (l *revWriter) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.close()
}

func (l *revWriter) close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err

}
