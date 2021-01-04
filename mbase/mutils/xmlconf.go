/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:10:37
 * @LastEditTime: 2020-12-17 09:10:37
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

func XmlConfSave(fpath string, v interface{}) error {
	bs, _ := xml.Marshal(v)
	err := ioutil.WriteFile(fpath, bs, os.ModePerm)
	return err
}
func XmlConfLoad(fpath string, v interface{}) error {
	bs, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(bs, v)
	return err
}

func XmlPrint(v interface{}) string {
	if v == nil {
		return "<></>"
	}
	data, _ := xml.Marshal(v)
	return string(data)
}
func XmlPrintPretty(v interface{}) string {
	if v == nil {
		return "<></>"
	}
	data, _ := xml.MarshalIndent(v, "", "\t")
	return string(data)
}
