/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:50:13
 * @LastEditTime: 2020-12-16 14:50:13
 * @LastEditors: Chen Long
 * @Reference:
 */

package msys

import (
	"encoding/binary"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func InterfaceIpString(ifname string) (string, error) {

	inter, err := net.InterfaceByName(ifname)
	if err == nil {
		addrs, err := inter.Addrs()
		if err == nil && len(addrs) > 0 {
			workip := addrs[0].String()
			ind := strings.Index(workip, "/")
			if ind >= 0 {
				workip = workip[:ind]
			}
			return workip, nil
		}
	}

	return "", err
}
func InterfaceIp(ifname string) uint32 {
	ipstr, _ := InterfaceIpString(ifname)
	if len(ipstr) == 0 {
		return 0
	}

	ip := net.ParseIP(ipstr)
	if ip == nil {
		return 0
	}

	if len(ip) == 16 {
		return binary.LittleEndian.Uint32(ip[12:16])
		//return mutils.BytesReadUint32(ip[12:16])
	}
	return binary.LittleEndian.Uint32(ip)
	//return mutils.BytesReadUint32(ip)
}
func InterfaceMac(ifname string) (string, error) {
	inter, err := net.InterfaceByName(ifname)
	if err == nil {
		return inter.HardwareAddr.String(), nil
	}

	return "", err
}
func LookupIp(host string) uint32 {
	ips, err := net.LookupIP(host)
	if err != nil {
		return 0
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return binary.LittleEndian.Uint32(ipv4)
			//return mutils.BytesReadUint32(ipv4)
		}
	}
	return 0
}
func ProcessId() int {
	return os.Getpid()
}
func ProcessName() string {
	return filepath.Base(os.Args[0])
}
func ProcessPath() string {
	filePath, _ := exec.LookPath(os.Args[0])

	return filePath
}

//返回的路径后面不带/
func ProcessDir() string {
	execPath := ProcessPath()
	absPath, _ := filepath.Abs(filepath.Dir(execPath))
	return absPath
}

//返回的路径后面不带/
func ProcessWd() string {
	pwd, _ := os.Getwd()
	return pwd
}
func FileSize(fpath string) (int64, error) {
	statinfo, err := os.Stat(fpath)
	if err != nil {
		return 0, err
	}
	return statinfo.Size(), nil
}
