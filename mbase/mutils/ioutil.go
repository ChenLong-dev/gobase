/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 20:59:23
 * @LastEditTime: 2020-12-16 20:59:23
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"encoding/binary"
	"io"
)

func WriteN(wd io.Writer, data []byte) (n int, err error) {
	dataLen := len(data)
	for n < dataLen {
		wn, werr := wd.Write(data[n:])
		if wn > 0 {
			n += wn
		}
		if werr != nil {
			return n, werr
		}
	}
	return n, nil
}
func ReadN(rd io.Reader, data []byte) (rn int, err error) {
	rPos := 0
	for rPos < len(data) {
		buff := data[rPos:]
		rn, err = rd.Read(buff)
		if rn > 0 {
			rPos += rn
		}
		if err != nil {
			return rPos, err
		}
	}
	return rPos, nil
}

func WriteUint16(wd io.Writer, d uint16) (err error) {
	bsD := BytesLittleEndianUint16(d)

	_, err = WriteN(wd, bsD)
	return err
}
func WriteUint32(wd io.Writer, d uint32) (err error) {
	bsD := BytesLittleEndianUint32(d)

	_, err = WriteN(wd, bsD)
	return err
}
func WriteUint64(wd io.Writer, d uint64) (err error) {
	bsD := BytesLittleEndianUint64(d)

	_, err = WriteN(wd, bsD)
	return err
}
func ReadUint16(rd io.Reader) (d uint16, err error) {
	var bs [2]byte
	_, err = ReadN(rd, bs[:])
	return binary.LittleEndian.Uint16(bs[:]), err
}
func ReadUint32(rd io.Reader) (d uint32, err error) {
	var bs [4]byte
	_, err = ReadN(rd, bs[:])
	return binary.LittleEndian.Uint32(bs[:]), err
}
func ReadUint64(rd io.Reader) (d uint64, err error) {
	var bs [8]byte
	_, err = ReadN(rd, bs[:])
	return binary.LittleEndian.Uint64(bs[:]), err
}

func WriteBigEndianUint16(wd io.Writer, d uint16) (err error) {
	bsD := BytesBigEndianUint16(d)

	_, err = WriteN(wd, bsD)
	return err
}
func WriteBigEndianUint32(wd io.Writer, d uint32) (err error) {
	bsD := BytesBigEndianUint32(d)

	_, err = WriteN(wd, bsD)
	return err
}
func WriteBigEndianUint64(wd io.Writer, d uint64) (err error) {
	bsD := BytesBigEndianUint64(d)

	_, err = WriteN(wd, bsD)
	return err
}
func ReadBigEndianUint16(rd io.Reader) (d uint16, err error) {
	var bs [2]byte
	_, err = ReadN(rd, bs[:])
	return binary.BigEndian.Uint16(bs[:]), err
}
func ReadBigEndianUint32(rd io.Reader) (d uint32, err error) {
	var bs [4]byte
	_, err = ReadN(rd, bs[:])
	return binary.BigEndian.Uint32(bs[:]), err
}
func ReadBigEndianUint64(rd io.Reader) (d uint64, err error) {
	var bs [8]byte
	_, err = ReadN(rd, bs[:])
	return binary.BigEndian.Uint64(bs[:]), err
}
