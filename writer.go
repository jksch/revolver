// Package revolver provides a revolving file writer.
package revolver

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type revWriter struct {
	conf Conf
	size int
	file *os.File
	lock sync.Mutex // synchronizes file operations
}

// Must wraps the call to NewWriter and returns a io.WriteCloser or panics
func Must(w io.WriteCloser, err error) io.WriteCloser {
	if err != nil {
		panic(fmt.Errorf("could not create revolving log writer, %v", err))
	}
	return w
}

// New returns a io.WriteCloser that writes revolving files as specified by the given conf.
// Calling New will always create a new file even if there is space left in other files.
// If the configured directory doesn't exist it will be created.
func New(conf Conf) (io.WriteCloser, error) {
	if err := ValidConf(conf); err != nil {
		return nil, err
	}
	conf = clean(conf)

	if err := setupDirs(conf.Dir); err != nil {
		return nil, fmt.Errorf("revolver setup, %v", err)
	}

	if err := countAndRemoveFiles(conf); err != nil {
		return nil, fmt.Errorf("revolver remove, %v", err)
	}

	file, err := createFile(conf)
	if err != nil {
		return nil, fmt.Errorf("revolver create, %v", err)
	}

	return &revWriter{
		conf: conf,
		file: file,
	}, nil
}

// Write the given bytes into the log file specified by the given conf.
// If there is not enough file space left,surplus files will be deleted and a new file will be created.
func (l *revWriter) Write(p []byte) (n int, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	size := len(p)
	if size > l.conf.MaxBytes {
		return 0, fmt.Errorf("revolver bytes to write %d over max file size %d", size, l.conf.MaxBytes)
	}
	if l.file == nil || l.size+size > l.conf.MaxBytes {
		if err := l.close(); err != nil {
			return 0, fmt.Errorf("revolver close, %v", err)
		}

		if err := countAndRemoveFiles(l.conf); err != nil {
			return 0, fmt.Errorf("revolver remove, %v", err)
		}

		file, err := createFile(l.conf)
		if err != nil {
			return 0, fmt.Errorf("revolver create, %v", err)

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
