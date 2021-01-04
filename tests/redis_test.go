/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:41:20
 * @LastEditTime: 2020-12-17 09:41:20
 * @LastEditors: Chen Long
 * @Reference:
 */

package tests

import (
	"context"
	"testing"

	"github.com/ChenLong-dev/gobase/mredis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedisSuite struct {
	suite.Suite
}

func (suite *RedisSuite) TestRedisPing() {
	cmd := mredis.RedisCluster().Ping(context.TODO())
	assert.Equal(suite.T(), "PONG", cmd.Val())
}

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}
