/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 21:03:42
 * @LastEditTime: 2020-12-17 01:20:07
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"runtime"
	"sort"
	"strconv"

	"github.com/ChenLong-dev/gobase/mlog"
)

func Upbound(x, n int) int {
	return (x + (n - 1)) & (^(n - 1))
}
func Lowbound(x, n int) int {
	return x & (^(n - 1))
}
func ReadVersion(v string) int64 {
	if v == "" {
		return int64(0)
	}
	var a, b, c int64
	_, err := fmt.Sscanf(v, "%d.%d.%d", &a, &b, &c)
	if err != nil {
		mlog.Debugf("v=%s.err=%v", v, err)
	}
	return (a << 48) | (b << 32) | (c & 0xffffffff)

}

func WriteVersion(v int64) string {
	return fmt.Sprintf("%d.%d.%d", (v>>48)&0xffffffff, (v>>32)&0xffffffff, v&0xffffffff)
}

func CompareVersion(v1 string, v2 string) int64 {
	return ReadVersion(v1) - ReadVersion(v2)
}

func ToInt64(v interface{}, def ...int64) int64 {
	var defval int64 = 0
	if len(def) > 0 {
		defval = def[0]
	}
	switch val := v.(type) {
	case int64:
		return val
	case uint64:
		return int64(val)
	case int:
		return int64(val)
	case uint:
		return int64(val)
	case float32:
		return int64(val)
	case float64:
		return int64(val)
	case int32:
		return int64(val)
	case uint32:
		return int64(val)
	case int16:
		return int64(val)
	case uint16:
		return int64(val)
	case int8:
		return int64(val)
	case uint8:
		return int64(val)
	case bool:
		if val {
			return 1
		} else {
			return 0
		}
	default:
		return defval
	}
}
func ToFloat64(v interface{}, def ...float64) float64 {
	var defval float64 = 0
	if len(def) > 0 {
		defval = def[0]
	}

	switch val := v.(type) {
	case float32:
		return float64(val)
	case float64:
		return float64(val)
	case int:
		return float64(val)
	case uint:
		return float64(val)
	case int64:
		return float64(val)
	case uint64:
		return float64(val)
	case int32:
		return float64(val)
	case uint32:
		return float64(val)
	case int16:
		return float64(val)
	case uint16:
		return float64(val)
	case int8:
		return float64(val)
	case uint8:
		return float64(val)
	case bool:
		if val {
			return 1.0
		} else {
			return 0.0
		}
	default:
		return defval
	}
}

func ToString(v interface{}, def ...string) string {
	var defval string = ""
	if len(def) > 0 {
		defval = def[0]
	}
	if val, ok := v.(string); ok {
		return val
	}
	return defval
}

func MaxInt(lhs, rhs int64) int64 {
	if lhs >= rhs {
		return lhs
	}
	return rhs
}
func MinInt(lhs, rhs int64) int64 {
	if lhs <= rhs {
		return lhs
	}
	return rhs
}
func DistanceUint64(lhs, rhs uint64) uint64 {
	if lhs < rhs {
		return rhs - lhs
	} else {
		return lhs - rhs
	}
}
func DistanceInt64(lhs, rhs int64) int64 {
	if lhs < rhs {
		return rhs - lhs
	} else {
		return lhs - rhs
	}
}
func DistanceUint32(lhs, rhs uint32) uint32 {
	if lhs < rhs {
		return rhs - lhs
	} else {
		return lhs - rhs
	}
}
func DistanceInt32(lhs, rhs int32) int32 {
	if lhs < rhs {
		return rhs - lhs
	} else {
		return lhs - rhs
	}
}
func DistanceUint16(lhs, rhs uint16) uint16 {
	if lhs < rhs {
		return rhs - lhs
	} else {
		return lhs - rhs
	}
}
func DistanceInt16(lhs, rhs int16) int16 {
	if lhs < rhs {
		return rhs - lhs
	} else {
		return lhs - rhs
	}
}

/*
 * 返回字符串长度，支持中文字符串
 */
func Strlen(str string) int {
	return len([]rune(str))
}

/*
 * 截取从 from 到 to 的子串，如果超出边界则以边界为准，支持中文字符串
 */
func SubString(str string, from, to int) (substring string) {
	slen := Strlen(str)
	if from < 0 {
		from = 0
	}
	if to > slen {
		to = slen
	}
	return string([]rune(str)[from:to])
}

/*
 * float to int64
 * 2.01 -> 201
 */
func FloatRound(f float64, n int) int64 {
	n10 := math.Pow10(n)
	inst, _ := strconv.ParseInt(fmt.Sprintf("%.0f", f*n10), 10, 64)
	return inst
}

// Convenience types for common cases
type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p Int64Slice) Sort() { sort.Sort(p) }

// Convenience wrappers for common cases

// Int64s sorts a slice of int64s in increasing order.
func SortInt64s(a []int64) { sort.Sort(Int64Slice(a)) }

/*
* uses binary search to find whether x is in slice
* slice must be in increasing order.
 */
func SearchInt64s(data []int64, x int64) bool {
	i := sort.Search(len(data), func(i int) bool { return data[i] >= x })
	if i < len(data) && data[i] == x {
		return true
	} else {
		return false
	}
}

func DumpStack(all bool) string {
	buf := make([]byte, 32*1024)
	for {
		n := runtime.Stack(buf, all)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}

	return string(buf)
}

func GZipCompress(src []byte) (result []byte, err error) {
	buf := new(bytes.Buffer)
	zw := gzip.NewWriter(buf)
	if _, err := zw.Write(src); err != nil {
		mlog.Error(err)
		return nil, err
	}
	if err := zw.Close(); err != nil {
		mlog.Error(err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func GZipUnCompress(src []byte) (result []byte, err error) {
	srcR := bytes.NewReader(src)
	zr, err := gzip.NewReader(srcR)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, zr); err != nil {
		return nil, err
	}
	if err := zr.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
