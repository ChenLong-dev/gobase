/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 20:36:00
 * @LastEditTime: 2020-12-16 20:58:52
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"
)

func MD5Hex(text string) string {
	md := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", md)
}

func SHA1Hex(text string) string {
	sh := sha1.Sum([]byte(text))
	return fmt.Sprintf("%x", sh)
}

//	92分
func BKDRHash(str string) uint32 {
	seed := uint32(131) // 31 131 1313 13131 131313 etc..
	hash := uint32(0)

	for _, c := range str {
		hash = hash*seed + uint32(c)
	}

	return hash & 0x7FFFFFFF
}

//	86分
func APHash(str string) uint32 {
	hash := uint32(0)

	for i, c := range str {
		if (i & 1) == 0 {
			hash = hash ^ ((hash << 7) ^ uint32(c) ^ (hash >> 3))
		} else {
			hash = hash ^ (^((hash << 11) ^ uint32(c) ^ (hash >> 5)))
		}
	}

	return hash & 0x7FFFFFFF
}

//	83分
func DJBHash(str string) uint32 {
	hash := uint32(5381)
	for _, c := range str {
		hash += (hash << 5) + uint32(c)
	}

	return hash & 0x7FFFFFFF
}

const signatureSalt = "as298Fd.*%fie9"

func Signature(key string, random string, timestamp int64) string {
	bs := []string{signatureSalt + key, random, fmt.Sprint(timestamp)}
	sort.Strings(bs)
	return SHA1Hex(strings.Join(bs, "-"))
}
