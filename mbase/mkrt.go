/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:14:12
 * @LastEditTime: 2020-12-17 09:14:12
 * @LastEditors: Chen Long
 * @Reference:
 */

package mbase

import (
	"os"
	"sync"
)

var mkrtExitWait sync.WaitGroup
var mkrtExitCode int
var mkrtRunning bool = false

func MKrtRun() int {
	mkrtExitWait.Add(1)
	mkrtRunning = true
	mkrtExitWait.Wait()
	return mkrtExitCode
}
func MKrtEixt(code int) {
	if !mkrtRunning {
		os.Exit(code)
	} else {
		mkrtRunning = false
		mkrtExitCode = code
		mkrtExitWait.Done()
	}
}
