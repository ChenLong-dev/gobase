/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:16:34
 * @LastEditTime: 2020-12-17 09:16:35
 * @LastEditors: Chen Long
 * @Reference:
 */

package mgrpc

import (
	"encoding/json"
	"fmt"
	"testing"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestErrorJson(t *testing.T) {
	data := ErrorRet{Code: 0, Message: "aaaa", Data: Person{Name: "aaaa", Age: 2}}
	str, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", str)
}
