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
		return err
	}
	return os.MkdirAll(dirs, 0755)
}

func createFile(conf Conf) (*os.File, error) {
	name := filepath.FromSlash(conf.Dir + "/" + conf.Prefix + conf.Middle())
	try := 0
	file := name
	for {
		file = file + conf.Suffix
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				return os.Create(file)
			}
			return nil, err
		}
		file = name + "_" + strconv.Itoa(try)
		try++
	}
}

func fileCount(conf Conf) (int, error) {
	files, err := ioutil.ReadDir(filepath.FromSlash(conf.Dir))
	if err != nil {
		return 0, err
	}

	count := 0
	for _, info := range files {
		if strings.HasPrefix(info.Name(), conf.Prefix) {
			count++
		}
	}
	return count, nil
}

func removeOldestFile(conf Conf) error {
	dir := filepath.FromSlash(conf.Dir + "/")
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	var oldest os.FileInfo
	for _, info := range files {
		if oldest == nil || oldest.ModTime().After(info.ModTime()) {
			oldest = info
		}
	}
	if oldest != nil {
		if err := os.Remove(dir + oldest.Name()); err != nil {
			return err
		}
	}
	return nil
}

func countAndRemoveFiles(conf Conf) error {
	count, err := fileCount(conf)
	if err != nil {
		return err
	}
	for conf.MaxFiles <= count {
		if err := removeOldestFile(conf); err != nil {
			return err
		}
		count--
	}
	return nil
}
