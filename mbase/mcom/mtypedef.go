/*
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 14:40:44
 * @LastEditTime: 2020-12-16 14:40:44
 * @LastEditors: Chen Long
 * @Reference: 
 */


 package mcom

/*  帧头定义  */

type MFrame struct {
	Len     uint32 			//	包长度。包括自身大小
	Flag    uint16	   		//	标志位
	Mark    uint16  	    //	记号
}
const MFrameSize int = 8
const MFrameMark uint16 = 0xabcd

/*  帧头标志定义  */
const (
	MFRAME_FLAG_TRACELOG uint16 = 0x1     //  打开调试日志标志
	MFRAME_FLAG_COMPRESS uint16 = 0x2	  //  压缩标志
)

/* 路由头 */
type MHeadPref struct {
	Plen     uint32
}
const MHeadPrefSize int = 4
/*  路由标志定义  */
const (
	MHEAD_FLAG_REQUEST			uint64 = (0x1)			//请求
	MHEAD_FLAG_RESPONSE			uint64 = (0x2)			//响应
	MHEAD_FLAG_COMPRESS			uint64 = (0x20)		   //压缩标记，指示业务负载数据是经过压缩的
)

