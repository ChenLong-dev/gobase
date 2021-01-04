/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:49:51
 * @LastEditTime: 2020-12-16 14:49:51
 * @LastEditors: Chen Long
 * @Reference:
 */

package msys

import (
	"github.com/fsnotify/fsnotify"
	"path/filepath"
)

type FSNOTIFY_TYPE int

const (
	FSNOTIFY_CREATE FSNOTIFY_TYPE = 1
	FSNOTIFY_WRITE  FSNOTIFY_TYPE = 3
	FSNOTIFY_REMOVE FSNOTIFY_TYPE = 5
	FSNOTIFY_RENAME FSNOTIFY_TYPE = 7
	FSNOTIFY_CHMOD  FSNOTIFY_TYPE = 9
)

type WatchCallback func(fsntype FSNOTIFY_TYPE)

type Fsnotify struct {
	path  string
	watch *fsnotify.Watcher
}

func (fsn *Fsnotify) Path() string {
	return fsn.path
}
func (fsn *Fsnotify) Watch(cb WatchCallback) error {
	for {
		select {
		case ev := <-fsn.watch.Events:
			{
				if ev.Name != fsn.path {
					break
				}

				if ev.Op&fsnotify.Create == fsnotify.Create {
					cb(FSNOTIFY_CREATE)
				}
				if ev.Op&fsnotify.Write == fsnotify.Write {
					cb(FSNOTIFY_WRITE)
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					cb(FSNOTIFY_REMOVE)
				}
				if ev.Op&fsnotify.Rename == fsnotify.Rename {
					cb(FSNOTIFY_RENAME)
				}
				if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
					cb(FSNOTIFY_CHMOD)
				}
			}
		case err := <-fsn.watch.Errors:
			{
				//mlog.Errorf("error : ", err);
				return err
			}
		}
	}
}
func NewFsnotify(path string) (fsn *Fsnotify, err error) {
	watch, werr := fsnotify.NewWatcher()
	if werr != nil {
		return nil, werr
	}
	err = watch.Add(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	return &Fsnotify{path: path, watch: watch}, nil
}
