/*
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 18:04:53
 * @LastEditTime: 2020-12-16 18:04:53
 * @LastEditors: Chen Long
 * @Reference: 
 */


 package mutils

/*
const char* build_time(void)
{
	static const char* _build_time = "["__DATE__ " " __TIME__ "]";
    return _build_time;
}
*/
import "C"

var (
	buildTime = C.GoString(C.build_time())
)

func BuildTime() string {
	return buildTime
}
