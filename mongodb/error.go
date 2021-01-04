/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:19:05
 * @LastEditTime: 2020-12-17 09:19:05
 * @LastEditors: Chen Long
 * @Reference:
 */

package mongodb

import "errors"

var (
	ErrNilValue       = errors.New("params can't be nil")
	ErrLength         = errors.New("params must be equal in length")
	ErrFileNotFound   = errors.New("file not found")
	ErrFileExist      = errors.New("file already exist")
	ErrUserExist      = errors.New("user already exist")
	ErrBucketNotFound = errors.New("not found bucket")
	ErrPartExist      = errors.New("part already exist")
	ErrBucketExist    = errors.New("bucket exist")
)
