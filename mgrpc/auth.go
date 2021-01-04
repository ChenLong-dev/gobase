/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:15:13
 * @LastEditTime: 2020-12-17 09:15:13
 * @LastEditors: Chen Long
 * @Reference:
 */

package mgrpc

import (
	"context"
	"github.com/ChenLong-dev/gobase/mlog"
	"github.com/ChenLong-dev/gobase/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"sync"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)
var GlobalAuthData *AuthData

type AuthData struct {
	sync.Mutex
	AuthData map[string]string
}

func NewGlobalAuth(data map[string]string) {
	GlobalAuthData.Lock()
	defer GlobalAuthData.Unlock()
	GlobalAuthData.AuthData = make(map[string]string)
	for k, v := range data {
		GlobalAuthData.AuthData[k] = v
	}
}

func (gad *AuthData) Add(key, value string) {
	GlobalAuthData.Lock()
	defer GlobalAuthData.Unlock()

	GlobalAuthData.AuthData[key] = value
}

func (gad *AuthData) IsAuth(timestamp, clientKey, clientSign string) bool {
	// 判断clientKey是否存在
	value, ok := GlobalAuthData.AuthData[clientKey]
	if ok {
		return false
	}
	// 计算本地sign
	localSign := utils.StringMd5(strings.Join([]string{value, timestamp}, ""))
	// 使用本地sign与传入sign进行对比
	if ret := strings.Compare(localSign, clientSign); ret != 0 {
		mlog.Errorf("[UnMatch ERROR] clientSign: %s, localSign: %s \n", clientSign, localSign)
		return false
	}
	return true
}

func authValid(md map[string][]string) bool {
	// 从metaHeader获取数据, 每个字段的数据格式都为 map[string][]string
	timestamp, ok := md["timestamp"]
	if !ok {
		mlog.Error("not found header timestamp")
		return false
	}
	clientKey, ok := md["clientKey"]
	if !ok {
		mlog.Errorf("not found header clientKey")
		return false
	}
	sign, ok := md["sign"]
	if !ok {
		mlog.Errorf("not found header sign")
		return false
	}
	return GlobalAuthData.IsAuth(timestamp[0], clientKey[0], sign[0])
}

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 获取metadata中的数据
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errMissingMetadata
		}
		if !authValid(md) {
			return nil, errInvalidToken
		}
		return handler(ctx, req)
	}
}
