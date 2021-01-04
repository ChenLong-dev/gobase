/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:42:38
 * @LastEditTime: 2020-12-17 09:42:38
 * @LastEditors: Chen Long
 * @Reference:
 */

package utils

import (
	"crypto/md5"
	"encoding/hex"
	"reflect"
)

func StringMd5(encryKey string) string {
	md5Obj := md5.New()
	md5Obj.Write([]byte(encryKey))
	return hex.EncodeToString(md5Obj.Sum(nil))
}

func CompareStrcut(fieldName []string, structA, structB interface{}) (diffrences []string) {
	structAValue := reflect.ValueOf(structA)
	structBValue := reflect.ValueOf(structB)
	for _, elem := range fieldName {
		switch structAValue.FieldByName(elem).Kind() {
		case reflect.Slice:

		}
		if !reflect.DeepEqual(structAValue.FieldByName(elem).String(), structBValue.FieldByName(elem).String()) {
			diffrences = append(diffrences, elem)
		}
	}
	return
}
