/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:12:19
 * @LastEditTime: 2020-12-17 09:12:19
 * @LastEditors: Chen Long
 * @Reference:
 */

package mbase

import (
	"fmt"

	"github.com/ChenLong-dev/gobase/mbase/msys"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
)

const mdeploy_conf_path = "/usr/local/SaaSBG/etc/deploy.conf"

type Deploy struct {
	Workif string `json:"workif" xml:"workif"`
	Workip string `json:"workip" xml:"workip"`
}

var gDeploy Deploy = Deploy{Workif: "eth0"}

func GetDeployConf() *Deploy {
	return &gDeploy
}

func initDeployConf() error {
	err := mutils.XmlConfLoad(mdeploy_conf_path, &gDeploy)
	if gDeploy.Workip == "" {
		workip, werr := msys.InterfaceIpString(gDeploy.Workif)
		if werr != nil {
			fmt.Printf("get workif[%s] ip error:%v\n", gDeploy.Workif, werr)
			return werr
		}
		gDeploy.Workip = workip
	}

	return err
}
