/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:05:00
 * @LastEditTime: 2020-12-17 09:05:00
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/ChenLong-dev/gobase/mbase/msys"
	"github.com/ChenLong-dev/gobase/mlog"
)

var (
	currentTime      = time.Now
	backupTimeFormat = "2006-01-02T15-04-05.000"
)

func openFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0744)
	if err != nil {
		return nil, fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	t := currentTime()
	timestamp := t.Format(backupTimeFormat)

	name := filepath.Join(dir, fmt.Sprintf("%s-%s-%d%s", prefix, timestamp, msys.ProcessId(), ext))

	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
}

func cleanOld(path string, maxFileNum int) {
	dir := filepath.Dir(path)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		mlog.Debugf("can't read log file directory: %s", err)
		return
	}

	key := make([]int64, 0, len(files))
	m := make(map[int64]os.FileInfo)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		timestamp := f.ModTime().UnixNano()
		for {
			if _, ok := m[timestamp]; ok {
				timestamp++
			} else {
				break
			}
		}
		key = append(key, timestamp)
		m[timestamp] = f
	}
	sort.Slice(key, func(i, j int) bool {
		return key[i] > key[j]
	})

	if len(key) <= maxFileNum {
		return
	}

	key = key[maxFileNum:]

	for _, ts := range key {
		f, _ := m[ts]
		os.Remove(filepath.Join(dir, f.Name()))
	}

}

func RedirectStdToFile(path string, maxFileNum int) error {
	f, err := openFile(path)
	if err != nil {
		return err
	}

	err = syscall.Dup2(int(f.Fd()), 1)
	err = syscall.Dup2(int(f.Fd()), 2)
	os.Stdout = f
	os.Stderr = f

	go cleanOld(path, maxFileNum)
	return err
}
