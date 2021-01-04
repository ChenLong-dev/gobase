package mbase

import (
	"encoding/xml"
	"io/ioutil"
)

var MBaseConfPath = "/usr/local/SaaSBG/etc/mbase.conf" //环境配置

type MBaseConf struct {
	SendChanMax int    `json:"sendChanMax" xml:"sendChanMax"`
	StreamBuff  int32  `json:"streamBuff" xml:"streamBuff"`
	SessionBuff int    `json:"sessionBuff" xml:"sessionBuff"`
	LogDir      string `json:"logDir" xml:"logDir"`
	CanTrace    int    `json:"canTrace" xml:"canTrace"`
}

var gMBaseConf = MBaseConf{LogDir:"/var/log", SendChanMax:30, StreamBuff:16536, SessionBuff:10*1024*1024}

func GetMBaseConf() *MBaseConf {
	return &gMBaseConf
}

func initMBaseConf() error {
	bs, err := ioutil.ReadFile(MBaseConfPath)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(bs, &gMBaseConf)
	return err

	return err
}


