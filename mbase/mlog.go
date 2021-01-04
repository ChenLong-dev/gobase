/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:14:37
 * @LastEditTime: 2020-12-17 09:14:37
 * @LastEditors: Chen Long
 * @Reference:
 */

package mbase

import (
	"fmt"
	"github.com/ChenLong-dev/gobase/mlog"
)

func Trace(v ...interface{}) {
	if (GetMBaseConf().CanTrace != 0) && (mlog.GetLevel() <= mlog.TRACE) {
		mlog.Output(mlog.TRACE, 3, "", "%s", fmt.Sprint(v...))
	}
}
func Tracef(format string, v ...interface{}) {
	if (GetMBaseConf().CanTrace != 0) && (mlog.GetLevel() <= mlog.TRACE) {
		mlog.Output(mlog.TRACE, 3, "", format, v...)
	}
}
