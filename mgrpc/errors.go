/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:16:05
 * @LastEditTime: 2020-12-17 09:16:05
 * @LastEditors: Chen Long
 * @Reference:
 */

package mgrpc

import (
	"encoding/json"

	"github.com/ChenLong-dev/gobase/mlog"
)

type ErrorRet struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func RetJson(code int, message string, data interface{}) ([]byte, error) {
	source := ErrorRet{Code: code, Message: message, Data: data}
	str, err := json.Marshal(source)
	if err != nil {
		mlog.Errorf("[JSON MARSHAL] error: %s", err)
		return nil, err
	}
	return str, nil
}
