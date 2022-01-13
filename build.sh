#!/bin/bash

set -e

# -------------------------------------------------------------------------------------------
# 需要自动部署的服务列表
# 示例, 请严格按照示例书写, 只决定再测试/开发环境是否自动化部署，编译，build镜像不受此参数影响
# SERVICES="account"
SERVICES=""

# 服务目录
CMD_DIR=cmd

GROUP=crema
# -------------------------------------------------------------------------------------------

OP=compile
VERSION=local

# 编译
function Compile() {
    echo "Complie $2:$VERSION"
    DATE=`date +%FT%T%z`
    FLAGS="-X git.cplus.link/go/utils/version._DATE_=$DATE -X go.cplus.link/go/utils/version._VERSION_=$VERSION"
    GO111MODULE=on GOOS=linux go build -v -o ./$1/$2 -ldflags "$FLAGS" ./$1
    echo "Complie $2:$VERSION OK"
    echo "------------------------------------"
}


# build镜像(build要求一定传入版本号)
function Build() {
    echo "Build $2:$VERSION"
    IMAGE="harbor.cplus.link/backend/$GROUP/$2"
    /kaniko/executor --context ./$1/ --dockerfile ./$1/Dockerfile --destination $IMAGE:$VERSION --cleanup
    rm -fr $1/$2
    echo "Build $2:$VERSION OK"
    echo "------------------------------------"
}

# 自动部署(deploy要求一定传入版本号)
function Deploy() {
    echo "Deploy $1:$VERSION"
    IMAGE="harbor.cplus.link/backend/$GROUP/$1:$VERSION"
    servicename=$1
    echo "ansible cplus-qa -m shell -a \"/data/bin/deploy -g $GROUP -s $servicename -i $IMAGE\""
    ansible cplus-qa -m shell -a "/data/bin/deploy -g $GROUP -s $servicename -i $IMAGE"
    echo "Deploy $1:$VERSION OK"
    echo "------------------------------------"
}

# 设置默认值
if [ -n "$1" ]; then
    OP=$1
fi
if [ -n "$2" ]; then
    VERSION=$2
fi
if [[ $(expr match "$2"  'v[0-9].*.[0-9].*.[0-9].*$') == 0 ]];then
    if [ "$COMMIT" == "" ];then
        VERSION=$VERSION-`date +%Y%m%d%H%M%S`
    else
        VERSION=$VERSION-$COMMIT
    fi
fi

if [ "$OP" == "compile" ];then
    for path in `ls -d $CMD_DIR/*`; do
        servicename="${path##*/}"
        if [ -f ${path} ]; then
            continue
        fi
        Compile ${path} $servicename
    done
elif [ "$OP" == "build" ];then
    for path in `ls -d $CMD_DIR/*`; do
        servicename="${path##*/}"
        if [ -f ${path} ]; then
            continue
        fi
        Build ${path} $servicename
    done
elif [ "$OP" == "clean" ];then
    for path in `ls -d $CMD_DIR/*`; do
        servicename="${path##*/}"
        if [ -f ${path} ]; then
            continue
        fi
        rm -fr $path/$servicename
    done
elif [ "$OP" == "deploy" ];then
    set -- $SERVICES
    if [[ -z $SERVICES ]];then
        for path in `ls -d $CMD_DIR/*`; do
            servicename="${path##*/}"
            if [ -f ${path} ]; then
                continue
            fi
            Deploy $servicename
        done
    else
        for servicename in $@; do
            Deploy $servicename
        done
    fi
elif [ "$OP" == "upgomod" ];then
    cat go.mod|grep -v "=>"|grep -Eo "git.cplus.link.*\s"|xargs go get -u
fi
