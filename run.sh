#!/bin/bash
arch=`uname -m`
case $arch in
x86_64)
     arch="amd64"
     ;;
aarch64)
     arch="arm64"
     ;;
*)
     arch="arm"
     ;;
esac
filename="xdd-linux-${arch}"
url="http://xdd.smxy.xyz/${filename}"
dirname="xdd"
curl -L $url -O dirname
