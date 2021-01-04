/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 21:00:21
 * @LastEditTime: 2020-12-16 21:00:21
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"reflect"
)

func SliceSearch(slice interface{}, obj interface{}) int {
	if slice == nil {
		return -1
	}

	targetValue := reflect.ValueOf(slice)
	switch reflect.TypeOf(slice).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return i
			}
		}
		return -1
	default:
		return -1
	}
}
func SliceRemove(slice interface{}, obj interface{}) (newSlice interface{}) {
	if slice == nil {
		return slice
	}

	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		return slice
	}

	pos := SliceSearch(slice, obj)
	if pos == -1 {
		return slice
	}

	sliceValue := reflect.ValueOf(slice)
	sliceHead := sliceValue.Slice(0, pos)
	sliceTail := sliceValue.Slice(pos+1, sliceValue.Len())
	return reflect.AppendSlice(sliceHead, sliceTail).Interface()
}
