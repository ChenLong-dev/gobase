/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 21:00:44
 * @LastEditTime: 2020-12-16 21:00:44
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func MPrintIp(ip uint32) string {
	var bytes [4]byte
	bytes[0] = byte(ip & 0xFF)
	bytes[1] = byte((ip >> 8) & 0xFF)
	bytes[2] = byte((ip >> 16) & 0xFF)
	bytes[3] = byte((ip >> 24) & 0xFF)

	return net.IPv4(bytes[0], bytes[1], bytes[2], bytes[3]).String()
}
func ParseIpv4(ipstr string) uint32 {
	ipstr = strings.TrimSpace(ipstr)
	if len(ipstr) == 0 {
		return 0
	}

	ip := net.ParseIP(ipstr)
	if ip == nil {
		return 0
	}

	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
		//return BytesReadUint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
	//return BytesReadUint32(ip)
}
func NetIpToUint32(ip net.IP) uint32 {
	if ipv4 := ip.To4(); ipv4 == nil {
		return 0
	} else {
		return binary.BigEndian.Uint32(ipv4)
	}
}
func AddNetIp(ip net.IP, add int) net.IP {
	if len(ip) < 4 {
		return nil
	}
	prefixLen := len(ip) - 4
	hip := binary.BigEndian.Uint32(ip[prefixLen:])
	hip += uint32(add)
	tmpIp := make([]byte, prefixLen, prefixLen+4)
	copy(tmpIp, ip[:prefixLen])
	return append(tmpIp, BytesBigEndianUint32(hip)...)
}
func CompareNetIp(lip, rip net.IP) int {
	l, r := lip.To4(), rip.To4()
	if len(l) != len(r) {
		return len(l) - len(r)
	}
	if l == nil {
		l, r = lip, rip
	}
	for i, n := 0, len(l); i < n; i++ {
		if l[i] < r[i] {
			return -1
		} else if l[i] > r[i] {
			return 1
		}
	}

	return 0
}

type IpRange struct {
	BeginIp net.IP
	EndIp   net.IP
}

func (ipr *IpRange) Contains(ip net.IP) bool {
	return CompareNetIp(ipr.BeginIp, ip) <= 0 && CompareNetIp(ipr.EndIp, ip) >= 0
}
func (ipr *IpRange) Distance() uint32 {
	return NetIpToUint32(ipr.EndIp) - NetIpToUint32(ipr.BeginIp) + 1
}
func genEndIp(cidr *net.IPNet) net.IP {
	ip := cidr.IP
	mask := cidr.Mask
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
		if len(mask) == net.IPv6len {
			mask = mask[12:]
		}
	}

	n := len(ip)
	out := make(net.IP, len(ip))
	for i := 0; i < n; i++ {
		out[i] = ip[i] | (^mask[i])
	}
	return out
}

/*
 * 参数：
 *      @str：描述范围的字符串,格式如下:
 *              类型 1  (1) 192.168.0.0
 *                      (2) ::2
 *              类型 2  (3) 192.168.0.0/16
 *                      (4) ::/16
 *              类型 3  (5) 192.168.0.0/255.255.255.0
 *						(6) ff::/ffff:ffff:ff00::
 *				类型 4  (7) 192.168.0.0-192.168.255.255
 *                      (8) ::2-::ff
 */
func ParseIpRange(str string) (ipr IpRange, err error) {
	if ind := strings.Index(str, "-"); ind != -1 {
		strBeginIp := strings.TrimSpace(str[:ind])
		strEndIp := strings.TrimSpace(str[ind+1:])
		if ipr.BeginIp = net.ParseIP(strBeginIp); ipr.BeginIp == nil {
			return ipr, fmt.Errorf("parse beginIp(%s) error", strBeginIp)
		}
		if ipr.EndIp = net.ParseIP(strEndIp); ipr.EndIp == nil {
			return ipr, fmt.Errorf("parse endIp(%s) error", strEndIp)
		}
		return ipr, nil
	} else {
		if ind := strings.Index(str, "/"); ind == -1 {
			ipr.BeginIp = net.ParseIP(str)
			ipr.EndIp = ipr.BeginIp
		} else {
			if ss := strings.Split(str, "/"); len(ss) > 1 && strings.Index(ss[1], ".") == -1 {
				_, cidr, cerr := net.ParseCIDR(strings.TrimSpace(str))
				if cerr != nil {
					return ipr, cerr
				}
				ipr.BeginIp = cidr.IP
				ipr.EndIp = genEndIp(cidr)
			} else {
				ip := net.ParseIP(ss[0])
				mask := net.ParseIP(ss[1])
				if len(ip) == 0 || len(mask) == 0 {
					return ipr, fmt.Errorf("%s format error", str)
				}
				ipr.BeginIp = ip.Mask(net.IPMask(mask))
				ipr.EndIp = genEndIp(&net.IPNet{IP: ip, Mask: net.IPMask(mask)})
			}
		}
	}
	return ipr, nil
}

func ParseInt64(str string, def int64) int64 {
	if v, e := strconv.ParseInt(str, 0, 0); e == nil {
		return v
	} else {
		return def
	}
}
func ParseUint64(str string, def uint64) uint64 {
	if v, e := strconv.ParseUint(str, 0, 0); e == nil {
		return v
	} else {
		return def
	}
}
func ParseFloat64(str string, def float64) float64 {
	if v, e := strconv.ParseFloat(str, 64); e == nil {
		return v
	} else {
		return def
	}
}

func spaceTerminator(str string) bool {
	terminators := []string{" ", "\t"}
	for _, term := range terminators {
		if strings.HasPrefix(str, term) {
			return true
		}
	}
	return false
}
func ReadWord(str string) (word string, termpos int) {
	return ReadWordFunc(str, spaceTerminator)
}
func ReadWordFunc(str string, termFunc func(restr string) bool) (word string, termpos int) {
	str = strings.TrimSpace(str)
	strLen := len(str)
	if strLen == 0 {
		return "", 0
	}

	prevQuota := false
	prevSlash := false
	for i, c := range str {
		if c == '\\' {
			if prevSlash {
				word += str[i : i+1]
				prevSlash = false
			} else {
				prevSlash = true
			}
		} else {
			if prevSlash {
				word += str[i : i+1]
			} else if c == '"' {
				if prevQuota {
					prevQuota = false
				} else {
					prevQuota = true
				}
			} else {
				if prevQuota || !termFunc(str[i:]) {
					word += str[i : i+1]
				} else {
					return word, i
				}
			}
			prevSlash = false
		}
	}
	return word, strLen
}

type Mml map[string]string

func NewMml(s string, sep string) (Mml, error) {

	if len(s) == 0 {
		return nil, fmt.Errorf("no data")
	}
	mml := make(map[string]string)

	for len(s) > 0 {
		key, kn := ReadWordFunc(s, func(str string) bool { return strings.HasPrefix(str, "=") })
		if kn == len(s) {
			return mml, fmt.Errorf("no value found!")
		}
		s = s[kn+1:]
		val, vn := ReadWordFunc(s, func(str string) bool { return strings.HasPrefix(str, ",") })
		if vn == len(s) {
			s = s[vn:]
		} else {
			s = s[vn+1:]
		}
		mml[key] = val
	}

	return mml, nil
}

func (mml Mml) Write(sep string) string {

	var buf bytes.Buffer

	for k, v := range mml {
		if buf.Len() == 0 {
			buf.WriteString(fmt.Sprintf("%s=%s", k, v))
		} else {
			buf.WriteString(fmt.Sprintf("%s%s=%s", sep, k, v))
		}
	}
	return buf.String()
}

/* 会自动解析是8 10 16 进制 */
func (mml Mml) GetInt64(k string, def int64) int64 {
	if val, ok := mml[k]; ok {
		if v, e := strconv.ParseInt(val, 0, 0); e == nil {
			return v
		} else {
			return def
		}
	}
	return def
}

func (mml Mml) GetUint64(k string, def uint64) uint64 {

	if val, ok := mml[k]; ok {
		if v, e := strconv.ParseUint(val, 0, 0); e == nil {
			return v
		} else {
			return def
		}
	}
	return def
}

func (mml Mml) GetInt(k string, def int) int {

	if val, ok := mml[k]; ok {
		if v, e := strconv.ParseInt(val, 0, 0); e == nil {
			return int(v)
		} else {
			return def
		}
	}
	return def
}

func (mml Mml) GetUint(k string, def uint) uint {

	if val, ok := mml[k]; ok {
		if v, e := strconv.ParseInt(val, 0, 0); e == nil {
			return uint(v)
		} else {
			return def
		}
	}
	return def
}

func (mml Mml) GetUint32(k string, def uint32) uint32 {

	if val, ok := mml[k]; ok {
		if v, e := strconv.ParseInt(val, 10, 0); e == nil {
			return uint32(v)
		} else {
			return def
		}
	}
	return def
}

func (mml Mml) GetInt32(k string, def int32) int32 {

	if val, ok := mml[k]; ok {
		if v, e := strconv.ParseInt(val, 10, 0); e == nil {
			return int32(v)
		} else {
			return def
		}
	}
	return def
}

func (mml Mml) GetString(k string, def string) string {

	if val, ok := mml[k]; ok {
		return val
	}
	return def
}

func (mml Mml) GetBool(k string, def bool) bool {
	if val, ok := mml[k]; ok {
		if b, err := strconv.ParseBool(val); err != nil {
			return def
		} else {
			return b
		}
	}
	return def
}

func (mml Mml) GetFloat64(k string, def float64) float64 {
	if val, ok := mml[k]; ok {
		if v, e := strconv.ParseFloat(val, 64); e == nil {
			return v
		} else {
			return def
		}
	}
	return def
}

func (mml Mml) GetFloat32(k string, def float32) float32 {
	if val, ok := mml[k]; ok {
		if v, e := strconv.ParseFloat(val, 32); e == nil {
			return float32(v)
		} else {
			return def
		}
	}

	return def
}

func (mml Mml) GetIp(k string, def net.IP) net.IP {
	return net.ParseIP(k)
	/*if val, ok := mml[k] ; ok {
		ipSegs := strings.Split(val, ".")
		var ipInt uint32 = 0
		var pos uint = 24
		for _, ipSeg := range ipSegs {
			tempInt, _ := strconv.Atoi(ipSeg)
			tempInt = tempInt << pos
			ipInt = ipInt | uint32(tempInt)
			pos -= 8
		}

		return  ipInt, true
	}
	return def, false*/
}
