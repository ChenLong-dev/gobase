/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:40:38
 * @LastEditTime: 2020-12-17 09:40:39
 * @LastEditors: Chen Long
 * @Reference:
 */

package tests

import (
	"fmt"
	"github.com/ChenLong-dev/gobase/etcd"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type EtcdSuite struct {
	suite.Suite
}

func (suite *EtcdSuite) TestEtcdDisCovery() {
	etcdDis := etcd.NewEtcdDis("services")
	etcdDis.Watch("v1.0.0")
	time.Sleep(10 * time.Second)
	node, ok := etcdDis.GetGrpcNodeByPath("v1.0.0", "/v1.0.0/saas_hello_demo")
	if !ok {
		fmt.Println("Not Found /saas_hello_demo")
	}
	fmt.Println(string(node.Info))
}

func TestEtcdSuite(t *testing.T) {
	suite.Run(t, new(EtcdSuite))
}
