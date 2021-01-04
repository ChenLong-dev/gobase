/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:50:45
 * @LastEditTime: 2020-12-16 14:50:45
 * @LastEditors: Chen Long
 * @Reference:
 */

package msys

import (
	"runtime/debug"
	"syscall"
)

func SetMaxFdSize(maxFdSize uint64) error {
	var rLimit syscall.Rlimit

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}

	rLimit.Max = maxFdSize
	rLimit.Cur = maxFdSize
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}

	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}

	return nil
}

func EnableCore() error {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_CORE, &rLimit)
	if err != nil {
		return err
	}

	rLimit.Max = ^uint64(0) //unlimited
	rLimit.Cur = ^uint64(0)

	err = syscall.Setrlimit(syscall.RLIMIT_CORE, &rLimit)
	if err != nil {
		return err
	}

	err = syscall.Getrlimit(syscall.RLIMIT_CORE, &rLimit)
	if err != nil {
		return err
	}

	debug.SetTraceback("crash")

	return nil
}
