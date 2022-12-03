# !/bin/bash
VERSION=`cat ./conf/version.go|grep 'VERSION'|awk -F = {'print $2'}|awk -F \" {'print $2'}`
echo build $VERSION
NAME=runnerGo
go build -o $NAME
mv $NAME $NAME-mac-v$VERSION
tar -zcvf $NAME-mac-v$VERSION.tgz $NAME-mac-v$VERSION
rm -rf $NAME-mac-v$VERSION

env GOOS=linux go build -o $NAME
mv $NAME $NAME-linux-v$VERSION
tar -zcvf $NAME-linux-v$VERSION.tgz $NAME-linux-v$VERSION
rm -rf $NAME-linux-v$VERSION
