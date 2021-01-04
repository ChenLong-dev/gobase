/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:11:15
 * @LastEditTime: 2020-12-16 14:11:15
 * @LastEditors: Chen Long
 * @Reference:
 */

package ac

import (
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/msys"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"go.etcd.io/etcd/pkg/fileutil"
	"os"
	"sync/atomic"
)

func DirutilsPathTransform(dir string) string {
	if len(dir) > 0 && dir[0] == '@' {
		return msys.ProcessDir() + dir[1:]
	} else {
		return dir
	}
}

type EpsLocType int32

const (
	EPS_LOC_SELF   = 0  /**< 未初始化，或通讯时本机通讯 */
	EPS_LOC_AGNT   = 1  /**< 位于agent */
	EPS_LOC_MGR    = 2  /**< 位于manager */
	EPS_LOC_MGRDC  = 3  /**< 位于manager_analysis */
	EPS_LOC_CLD    = 4  /**< 位于cloud */
	EPS_LOC_CLDDC  = 5  /**< 位于cloud_analysis */
	EPS_LOC_CLDSP1 = 6  /**< 接入cloud的特殊节点1 */
	EPS_LOC_CLDSP2 = 7  /**< 接入cloud的特殊节点2 */
	EPS_LOC_CLDSP3 = 8  /**< 接入cloud的特殊节点3 */
	EPS_LOC_CLDSP4 = 9  /**< 接入cloud的特殊节点4 */
	EPS_LOC_CLDSP5 = 10 /**< 接入cloud的特殊节点5 */
	EPS_LOC_CLDSP6 = 11 /**< 接入cloud的特殊节点6 */
	EPS_LOC_MGRSP1 = 12 /**< 接入manager的特殊节点1 */
	EPS_LOC_MGRSP2 = 13 /**< 接入manager的特殊节点2 */
	EPS_LOC_MGRSP3 = 14 /**< 接入manager的特殊节点3 */
	EPS_LOC_MGRSP4 = 15 /**< 接入manager的特殊节点4 */
	_EPS_LOC_MAX   = 16 /* 这个不能超过16，在ipc_proxy中有一个rflags为16位，如需要超过该值，需要增大rflags */
)

func EpsReadMode() EpsLocType {
	f, ferr := os.Open("/ac/etc/epsmode")
	if ferr != nil {
		return EPS_LOC_SELF
	}
	defer f.Close()

	buff := make([]byte, 32)
	if n, rerr := f.Read(buff); rerr != nil {
		return EPS_LOC_SELF
	} else {
		buff = buff[:n]
	}

	ret := mutils.ParseInt64(string(buff), 0)
	if ret > 0 && ret < int64(_EPS_LOC_MAX) {
		return EpsLocType(ret)
	}
	return EPS_LOC_SELF
}

var s_ret int32

func EpsGetLocation() EpsLocType {
	if ret := atomic.LoadInt32(&s_ret); ret != 0 {
		return EpsLocType(s_ret)
	}

	var serv string
	if env := os.Getenv("EPS_INSTALL_ROOT"); env != "" {
		serv = fmt.Sprintf("%s/services", env)
	} else {
		serv = DirutilsPathTransform("@/../services")
	}

	type epsLocation struct {
		name string
		ret  EpsLocType
	}
	epsLocations := []epsLocation{
		{"pm_manager", EPS_LOC_MGR},
		{"pm_cloud", EPS_LOC_CLD},
		{"pm_clouddc", EPS_LOC_CLDDC},
		{"pm_managerdc", EPS_LOC_MGRDC},
		{"pm_cloudsp1", EPS_LOC_CLDSP1},
		{"pm_cloudsp2", EPS_LOC_CLDSP2},
		{"pm_cloudsp3", EPS_LOC_CLDSP3},
		{"pm_cloudsp4", EPS_LOC_CLDSP4},
		{"pm_cloudsp5", EPS_LOC_CLDSP5},
		{"pm_cloudsp6", EPS_LOC_CLDSP6},
		{"pm_managersp1", EPS_LOC_MGRSP1},
		{"pm_managersp2", EPS_LOC_MGRSP2},
		{"pm_managersp3", EPS_LOC_MGRSP3},
		{"pm_managersp4", EPS_LOC_MGRSP4},
		{"pm_agent", EPS_LOC_AGNT},
	}
	for _, el := range epsLocations {
		if fileutil.Exist(serv + "/" + el.name) {
			atomic.StoreInt32(&s_ret, int32(el.ret))
			return el.ret
		}
	}

	if ret := EpsReadMode(); ret != EPS_LOC_SELF {
		atomic.StoreInt32(&s_ret, int32(ret))
		return ret
	}

	atomic.StoreInt32(&s_ret, int32(EPS_LOC_AGNT))
	return EPS_LOC_AGNT
}
