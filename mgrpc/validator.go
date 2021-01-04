/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:17:57
 * @LastEditTime: 2020-12-17 09:17:58
 * @LastEditors: Chen Long
 * @Reference:
 */

package mgrpc

import (
	"context"
	"google.golang.org/grpc"
)

type Validator interface {
	Validate() interface{}
}

func ValidateInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(Validator); ok {
			if ret := v.Validate(); ret != nil {
				return ret, nil
			}
		}
		return handler(ctx, req)
	}
}
