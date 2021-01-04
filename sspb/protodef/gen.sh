workpath=$(cd `dirname $0`; pwd)
###
 # @Description: 
 # @Author: Chen Long
 # @Date: 2020-12-17 09:23:33
 # @LastEditTime: 2020-12-17 09:23:52
 # @LastEditors: Chen Long
 # @Reference: 
### 
cd $workpath

ROOTDIR=`while true; do if [ -d protos ]; then pwd;exit; else cd ..;fi;done;`
PROTODIR=${ROOTDIR}/protos/
#PROTODIR7=${ROOTDIR}/env/source/libs/comproto-cpp/proto_def/

for proto in `ls ${PROTODIR}*.proto`; do
	cp -f $proto ./
done

to_utf8()
{
    fname=$1
    encode=`file -i $fname | awk -F '=' '{print $2}'`
    if [ "$encode" == "" ]; then
        exit 1
    fi
    
    if [ "$encode" == "utf-8" ]; then
        return 0
    fi
    
    iconv -f $encode -t "utf-8" $fname > ${fname}.tmp
    mv -f ${fname}.tmp ${fname}
}

cpwd=`pwd`

for f in `ls *.proto` 
do
    fname=$cpwd/$f
        
    to_utf8 ${fname}
done

#protoc --go_out=.  *.proto 
protoc *.proto --go_out=plugins=grpc:..


