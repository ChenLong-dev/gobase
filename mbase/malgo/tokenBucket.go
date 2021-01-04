/*
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 14:17:10
 * @LastEditTime: 2020-12-16 14:17:11
 * @LastEditors: Chen Long
 * @Reference: 
 */


 package malgo

import (
	"time"
	"fmt"
)

/*
	频率控制令牌桶
 */
type cmdTokenBucketOp int
const (
	TB_OP_RESET cmdTokenBucketOp = 1
//	TB_OP_STOP  cmdTokenBucketOp = 10
)
type TokenBucketParams struct {
	Max 		int
	Interval 	int
}
type cmdTokenBucket struct {
	op		cmdTokenBucketOp
	param   interface{}
}
type TokenBucket struct {
	c 			chan int
	cmdC 		chan *cmdTokenBucket
	max 		int
	interval	int		//ms
}
func (tb *TokenBucket) String() string {
	if tb == nil {
		return "{}"
	}

	return fmt.Sprintf("{len:%d,cap:%d,interval:%d}", len(tb.c), cap(tb.c), tb.interval)
}
func (tb *TokenBucket) Len() int {
	return len(tb.c)
}
func (tb *TokenBucket) Cap() int {
	return cap(tb.c)
}
func (tb *TokenBucket) Reset(max int, interval int) error {
	if max <= 0 || interval <= 0 {
		return fmt.Errorf("params error!")
	}

	tb.cmdC <- &cmdTokenBucket{op:TB_OP_RESET, param:&TokenBucketParams{Max:max,Interval:interval}}
	return nil
}
func (tb *TokenBucket) Get() {
	<- tb.c
}
func (tb *TokenBucket) GetUnblock() bool {
	select {
	case <- tb.c:
		return true
	default:
		return false
	}
}
func (tb *TokenBucket) put() {
	for i := len(tb.c); i < tb.max; i++ {
		select {
		case tb.c <- 1:
		default:
			return
		}
	}
}
func (tb *TokenBucket) Run() {
	if tb.cmdC != nil {
		return
	}

	tb.cmdC = make(chan *cmdTokenBucket)
	tb.c = make(chan int, tb.max)

	go func() {
		tick := time.NewTicker(time.Duration(tb.interval)*time.Millisecond)
		for {
			select {
			case <-tick.C:
				tb.put()
			case cmd := <-tb.cmdC:
				if cmd == nil {
					close(tb.c)
					tick.Stop()
					return
				}
				if cmd.op == TB_OP_RESET {
					param := cmd.param.(*TokenBucketParams)

					tb.max = param.Max
					tb.interval = param.Interval
					tick.Stop()
					tick = time.NewTicker(time.Duration(tb.interval)*time.Millisecond)
					c := tb.c
					tb.c = make(chan int, tb.max)
					close(c)
				}
			}
		}
	} ()
}
func (tb *TokenBucket) Stop() {
	close(tb.cmdC)
}
func NewTokenBucket(max int, interval int) *TokenBucket {
	tb := &TokenBucket{max:max, interval:interval}
	tb.Run()

	return tb
}

/*
	并发控制令牌桶
 */
 type ConcurrentBucket struct {
 	c 	chan int
 }
 func (cb *ConcurrentBucket)	Get() {
 	<-cb.c
 }
 func (cb *ConcurrentBucket)	Put() {
 	cb.c<-1
 }
 func NewConcurrentBucket(max int) *ConcurrentBucket {
 	cb := &ConcurrentBucket{}
 	cb.c = make(chan int, max)
 	for i :=0 ; i < max; i++ {
 		cb.Put()
	}
	return cb
 }