/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:13:36
 * @LastEditTime: 2020-12-17 09:13:36
 * @LastEditors: Chen Long
 * @Reference:
 */

package mbase

import (
	"fmt"
)

type MError struct {
	errno int32
}

func (merr *MError) Error() string {
	if merr == nil {
		return ""
	}
	if merr.errno == 0 {
		return "ok(0)"
	}

	return fmt.Sprintf("error(%d)", merr.errno)
}
func (merr *MError) Errno() int32 {
	if merr == nil {
		return 0
	}
	return merr.errno
}
func NewMError(errno int32) error {
	if errno == 0 {
		return nil
	}
	return &MError{errno: errno}
}
