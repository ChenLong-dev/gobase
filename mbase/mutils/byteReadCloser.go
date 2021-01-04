/*
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 18:04:34
 * @LastEditTime: 2020-12-16 18:04:35
 * @LastEditors: Chen Long
 * @Reference: 
 */


 package mutils

import (
	"io"
	"bytes"
	"fmt"
)

type ByteReadCloser struct {
	*bytes.Reader
}
func (brc *ByteReadCloser) Read(p []byte) (n int, err error) {
	reader := brc.Reader
	if reader == nil {
		return 0, fmt.Errorf("read after close")
	}
	return reader.Read(p)
}
func (brc *ByteReadCloser) Close() error {
	brc.Reader = nil
	return nil
}

func NewByteReadCloser(data []byte) io.ReadCloser {
	brc := &ByteReadCloser{}
	brc.Reader = bytes.NewReader(data)

	return brc
}