/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 18:05:34
 * @LastEditTime: 2020-12-16 20:34:28
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type DnsResolver struct {
	m            sync.Map
	updatePeriod int64 //second
	once         sync.Once
}

func resolveIPV4(dname string) (ip string, err error) {
	ips, err := net.LookupIP(dname)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no ips resolve get")
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("no ipv4 match")
}
func (dr *DnsResolver) Get(dname string) (ip string, err error) {
	if v, ok := dr.m.Load(dname); ok {
		return v.(string), nil
	}

	ip, err = resolveIPV4(dname)
	if err != nil {
		return ip, err
	}
	dr.m.Store(dname, ip)

	dr.once.Do(func() {
		if dr.updatePeriod > 0 {
			go dr.cycleUpdate()
		}
	})

	return ip, err
}
func (dr *DnsResolver) cycleUpdate() {
	type drPair struct {
		dname string
		ip    string
	}
	for {
		time.Sleep(time.Duration(dr.updatePeriod) * time.Second)

		var amended []*drPair
		dr.m.Range(func(key, val interface{}) bool {
			dname := key.(string)
			ip := val.(string)
			if rip, err := resolveIPV4(dname); err == nil {
				if rip != "" && rip != ip {
					amended = append(amended, &drPair{dname: dname, ip: rip})
				}
			}
			return true
		})

		for _, drp := range amended {
			dr.m.Store(drp.dname, drp.ip)
		}
	}
}

func NewDnsResolver(updatePeriod int64) *DnsResolver {
	return &DnsResolver{updatePeriod: updatePeriod}
}

var stdDnsResolver = NewDnsResolver(60)

func ResolveDns(dname string) (ip string, err error) {
	return stdDnsResolver.Get(dname)
}
