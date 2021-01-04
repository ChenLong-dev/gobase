/*
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 14:53:37
 * @LastEditTime: 2020-12-16 14:53:37
 * @LastEditors: Chen Long
 * @Reference: 
 */


 package msys
 import (
	 "errors"
	 "fmt"
	 "time"
 )
 
 type MTimerOverFunc		func(pridata interface{})
 type MTimer struct {
	 start 		int64
	 elapsed 	int64
	 overNotify	MTimerOverFunc
	 pridata		interface{}
	 timer 		*time.Timer
 }
 
 func (mt *MTimer) String() string {
	 if mt == nil {
		 return "{}"
	 }
	 return fmt.Sprintf("{start:%d,elapsed:%d,expire:%d,timer:%v}", mt.start, mt.elapsed, mt.Expire(), mt.timer)
 }
 func (mt *MTimer) Expire() int64 {
	 if mt == nil {
		 return 0
	 }
	 now := int64(NowMillisecond())
	 over := mt.start + mt.elapsed
	 return over - now
 }
 func (mt *MTimer) Elapsed() int64 {
	 return mt.elapsed
 }
 func (mt *MTimer) run() {
	 mt.start = int64(NowMillisecond())
	 f := func() {
		 mt.timer = nil
		 mt.overNotify(mt.pridata)
	 }
	 mt.timer = time.AfterFunc(time.Duration(mt.elapsed)*time.Millisecond, f)
 }
 func (mt *MTimer)	Stop() {
	 if mt == nil {
		 return
	 }
	 timer := mt.timer
	 mt.timer = nil
	 if timer != nil {
		 timer.Stop()
	 }
 }
 func (mt *MTimer)	Start() error {
	 if mt.elapsed == 0 {
		 return errors.New(fmt.Sprintf("params error:elapsed=%d", mt.elapsed))
	 }
 
	 mt.run()
	 return nil
 }
 func (mt *MTimer)	SetOverNotifyFunc(overNotify MTimerOverFunc) {
	 mt.overNotify = overNotify
 }
 
 func NewMTimer(elapsed int64, overNotify MTimerOverFunc, pridata interface{}) *MTimer {
	 return &MTimer{elapsed: elapsed, overNotify: overNotify, pridata: pridata}
 }
 
 func StartMTimer(elapsed int64, overNotify MTimerOverFunc, pridata interface{}) *MTimer {
	 mt := NewMTimer(elapsed, overNotify, pridata)
	 if err := mt.Start(); err != nil {
		 return nil
	 }
 
	 return mt
 }
 
 