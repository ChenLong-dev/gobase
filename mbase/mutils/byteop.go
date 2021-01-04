/*
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 18:03:28
 * @LastEditTime: 2020-12-16 18:03:28
 * @LastEditors: Chen Long
 * @Reference: 
 */


 package mutils

import "encoding/binary"

func bytesReadUint(data []byte) uint64 {
	len := len(data)
	if len > 8 {
		len = 8
	}

	var v uint64 = 0
	for i := 0; i < len; i++ {
		v += (uint64(data[i]) << uint(i * 8))
	}

	return v
}

func BytesReadUint16(data []byte) uint16 {
	var v uint16 = 0

	v += (uint16(data[0]) << 0)
	v += (uint16(data[1]) << 8)

	return v
}
func BytesReadUint32(data []byte) uint32 {
	var v uint32 = 0

	v += (uint32(data[0]) << 0)
	v += (uint32(data[1]) << 8)
	v += (uint32(data[2]) << 16)
	v += (uint32(data[3]) << 24)

	return v
}
func BytesReadUint64(data []byte) uint64 {
	var v uint64 = 0

	v += (uint64(data[0]) << 0)
	v += (uint64(data[1]) << 8)
	v += (uint64(data[2]) << 16)
	v += (uint64(data[3]) << 24)
	v += (uint64(data[4]) << 32)
	v += (uint64(data[5]) << 40)
	v += (uint64(data[6]) << 48)
	v += (uint64(data[7]) << 56)

	return v
}
func BytesWriteUint16(data []byte, v uint16) {
	data[0] = byte(v >> 0)
	data[1] = byte(v >> 8)
}
func BytesWriteUint32(data []byte, v uint32) {
	data[0] = byte(v >> 0)
	data[1] = byte(v >> 8)
	data[2] = byte(v >> 16)
	data[3] = byte(v >> 24)
}
func BytesWriteUint64(data []byte, v uint64) {
	data[0] = byte(v >> 0)
	data[1] = byte(v >> 8)
	data[2] = byte(v >> 16)
	data[3] = byte(v >> 24)
	data[4] = byte(v >> 32)
	data[5] = byte(v >> 40)
	data[6] = byte(v >> 48)
	data[7] = byte(v >> 56)
}
func BytesLittleEndianUint16Read(bs []byte) uint16 {
	if len(bs) < 2 {
		return 0
	}
	return binary.LittleEndian.Uint16(bs)
}
func BytesLittleEndianUint32Read(bs []byte) uint32 {
	if len(bs) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(bs)
}
func BytesLittleEndianUint64Read(bs []byte) uint64 {
	if len(bs) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(bs)
}
func BytesBigEndianUint16Read(bs []byte) uint16 {
	if len(bs) < 2 {
		return 0
	}
	return binary.BigEndian.Uint16(bs)
}
func BytesBigEndianUint32Read(bs []byte) uint32 {
	if len(bs) < 4 {
		return 0
	}
	return binary.BigEndian.Uint32(bs)
}
func BytesBigEndianUint64Read(bs []byte) uint64 {
	if len(bs) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bs)
}
func BytesLittleEndianUint16(v uint16) []byte {
	var bs [2]byte
	binary.LittleEndian.PutUint16(bs[:], v)
	return bs[:]
}
func BytesLittleEndianUint32(v uint32) []byte {
	var bs [4]byte
	binary.LittleEndian.PutUint32(bs[:], v)
	return bs[:]
}
func BytesLittleEndianUint64(v uint64) []byte {
	var bs [8]byte
	binary.LittleEndian.PutUint64(bs[:], v)
	return bs[:]
}

func BytesBigEndianUint16(v uint16) []byte {
	var bs [2]byte
	binary.BigEndian.PutUint16(bs[:], v)
	return bs[:]
}
func BytesBigEndianUint32(v uint32) []byte {
	var bs [4]byte
	binary.BigEndian.PutUint32(bs[:], v)
	return bs[:]
}
func BytesBigEndianUint64(v uint64) []byte {
	var bs [8]byte
	binary.BigEndian.PutUint64(bs[:], v)
	return bs[:]
}

func swap16(h uint16) uint16 {
	var n uint16 = 0

	n |= ( (h << 8) & 0xff00 )
	n |= ( (h >> 8) & 0xff )

	return n
}
func swap32(h uint32) uint32 {
	var n uint32 = 0

	n |= ( (h << 24) & 0xff000000 )
	n |= ( (h << 8) & 0xff0000 )
	n |= ( (h >> 8) & 0xff00 )
	n |= ( (h >> 24) & 0xff )

	return n
}
func swap64(h uint64) uint64 {
	var n uint64 = 0

	n |= ( (h << 56) & 0xff00000000000000 )
	n |= ( (h << 40) & 0xff000000000000 )
	n |= ( (h << 24) & 0xff0000000000 )
	n |= ( (h << 8) & 0xff00000000 )
	n |= ( (h >> 8) & 0xff000000 )
	n |= ( (h >> 24) & 0xff0000 )
	n |= ( (h >> 40) & 0xff00 )
	n |= ( (h >> 56) & 0xff )

	return n
}
func HtoN16(h uint16) uint16 {
	return swap16(h)
}
func HtoN32(h uint32) uint32 {
	return swap32(h)
}
func HtoN64(h uint64) uint64 {
	return swap64(h)
}
func NtoH16(n uint16) uint16 {
	return swap16(n)
}
func NtoH32(n uint32) uint32 {
	return swap32(n)
}
func NtoH64(n uint64) uint64 {
	return swap64(n)
}
func ReverseBytes(s []byte) []byte {
	n := len(s)
	if n == 0 {
		return nil
	}
	ns := make([]byte, n)
	for i := range s {
		ns[n-i-1] = s[i]
	}
	return ns
}