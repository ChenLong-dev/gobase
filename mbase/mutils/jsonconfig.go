/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 20:59:48
 * @LastEditTime: 2020-12-16 20:59:48
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func JsonConfSave(fpath string, v interface{}) error {
	bs, _ := json.Marshal(v)
	err := ioutil.WriteFile(fpath, bs, os.ModePerm)
	return err
}
func JsonConfLoad(fpath string, v interface{}) error {
	bs, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bs, v)
	return err
}

func JsonPrint(v interface{}) string {
	if v == nil {
		return "{}"
	}
	data, _ := json.Marshal(v)
	return string(data)
}
func JsonPrintPretty(v interface{}) string {
	if v == nil {
		return "{}"
	}
	data, _ := json.MarshalIndent(v, "", "\t")
	return string(data)
}
