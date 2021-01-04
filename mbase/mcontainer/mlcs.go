/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:44:40
 * @LastEditTime: 2020-12-16 14:44:40
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcontainer

import "sort"

func mlcsRune(ls []rune, rs []rune) int {
	lsize := len(ls) + 1
	rsize := len(rs) + 1

	dp := make([]int32, lsize*rsize)

	for i := 1; i < lsize; i++ {
		for j := 1; j < rsize; j++ {
			if ls[i-1] == rs[j-1] {
				dp[i*rsize+j] = dp[(i-1)*rsize+(j-1)] + 1
			} else {
				lv := dp[(i-1)*rsize+j]
				rv := dp[i*rsize+(j-1)]
				if lv > rv {
					dp[i*rsize+j] = lv
				} else {
					dp[i*rsize+j] = rv
				}
			}
		}
	}
	return int(dp[(lsize-1)*rsize+(rsize-1)])
}
func mlcsByte(ls []byte, rs []byte) int {
	lsize := len(ls) + 1
	rsize := len(rs) + 1

	dp := make([]int32, lsize*rsize)

	for i := 1; i < lsize; i++ {
		for j := 1; j < rsize; j++ {
			if ls[i-1] == rs[j-1] {
				dp[i*rsize+j] = dp[(i-1)*rsize+(j-1)] + 1
			} else {
				lv := dp[(i-1)*rsize+j]
				rv := dp[i*rsize+(j-1)]
				if lv > rv {
					dp[i*rsize+j] = lv
				} else {
					dp[i*rsize+j] = rv
				}
			}
		}
	}
	return int(dp[(lsize-1)*rsize+(rsize-1)])
}
func MLcsString(lstr string, rstr string) int {
	ls := []byte(lstr)
	rs := []byte(rstr)
	return mlcsByte(ls, rs)
}
func MLcsStringSort(lstr string, rstr string) int {
	ls := []byte(lstr)
	rs := []byte(rstr)

	sort.Slice(ls, func(i, j int) bool {
		return ls[i] < ls[i]
	})

	sort.Slice(rs, func(i, j int) bool {
		return rs[i] < rs[i]
	})

	return mlcsByte(ls, rs)
}
func MLcsByte(lr []byte, rr []byte) int {
	return mlcsByte(lr, rr)
}
func MLcsByteSort(lr []byte, rr []byte) int {
	sort.Slice(lr, func(i, j int) bool {
		return lr[i] < lr[i]
	})

	sort.Slice(rr, func(i, j int) bool {
		return rr[i] < rr[i]
	})

	return mlcsByte(lr, rr)
}
func MLcsRune(lr []rune, rr []rune) int {
	return mlcsRune(lr, rr)
}
func MLcsRuneSort(lr []rune, rr []rune) int {
	sort.Slice(lr, func(i, j int) bool {
		return lr[i] < lr[i]
	})

	sort.Slice(rr, func(i, j int) bool {
		return rr[i] < rr[i]
	})

	return mlcsRune(lr, rr)
}
