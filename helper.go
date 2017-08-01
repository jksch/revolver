package revolver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func setupDirs(dirs string) error {
	dirs = filepath.FromSlash(dirs)
	if dirs == "." {
		return nil
	}
	info, err := os.Stat(dirs)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", dirs)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("error in dir setup, %v", err)
	}
	return os.MkdirAll(dirs, 0755)
}

func createFile(dir, prefix, suffix string, filename func() string) (*os.File, error) {
	name := filepath.FromSlash(filepath.Join(dir, prefix+filename()))
	try := 0
	file := name
	for {
		file = file + suffix
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				return os.Create(file)
			}
			return nil, fmt.Errorf("error on create file, %v", err)
		}
		file = name + "_" + strconv.Itoa(try)
		try++
	}
}

func fileCount(dir, prefix string) (int, error) {
	files, err := ioutil.ReadDir(filepath.FromSlash(dir))
	if err != nil {
		return 0, fmt.Errorf("error while counting files, %v", err)
	}

	count := 0
	for _, info := range files {
		if isRevolverFile(prefix, info) {
			count++
		}
	}
	return count, nil
}

func removeOldestFile(dir, prefix string) error {
	dir = filepath.FromSlash(dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error listing oldest file, %v", err)
	}
	var oldest os.FileInfo
	for _, info := range files {
		if isRevolverFile(prefix, info) && isOlder(info, oldest) {
			oldest = info
		}
	}
	if oldest != nil {
		if err := os.Remove(filepath.Join(dir, oldest.Name())); err != nil {
			return fmt.Errorf("error removing oldest file, %v", err)
		}
	}
	return nil
}

func isRevolverFile(prefix string, file os.FileInfo) bool {
	return !file.IsDir() && strings.HasPrefix(file.Name(), prefix)
}

func isOlder(test, old os.FileInfo) bool {
	return old == nil || old.ModTime().After(test.ModTime())
}

func countAndRemoveFiles(dir, prefix string, maxFiles int) error {
	count, err := fileCount(dir, prefix)
	if err != nil {
		return err
	}
	for maxFiles <= count {
		if err := removeOldestFile(dir, prefix); err != nil {
			return err
		}
		count--
	}
	return nil
}
