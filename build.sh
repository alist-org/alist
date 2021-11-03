#!/bin/bash

appName="alist"
builtAt="$(date +'%F %T %z')"
goVersion=$(go version | sed 's/go version //')
gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
gitCommit=$(git log --pretty=format:"%h" -1)

if [ "$1" == "release" ]; then
  gitTag=$(git describe --abbrev=0 --tags)
else
  gitTag="beta"
fi

ldflags="\
-w -s \
-X 'main.builtAt=$builtAt' \
-X 'main.goVersion=$goVersion' \
-X 'main.gitAuthor=$gitAuthor' \
-X 'main.gitCommit=$gitCommit' \
-X 'main.gitTag=$gitTag' \
"

cp -R ../alist-web/dist/* public

xgo -out alist -ldflags="$ldflags" .
mkdir "build"
mv alist-* build
cd build || exit
upx -9 ./*
find . -type f -print0 | xargs -0 md5sum > md5.txt

# compress file (release)
if [ "$1" == "release" ]; then
    mkdir compress
    mv md5.txt compress
    for i in `find . -type f -name "$appName-linux-*"`
    do
      tar -czvf compress/"$i".tar.gz "$i"
    done
    for i in `find . -type f -name "$appName-darwin-*"`
    do
      tar -czvf compress/"$i".tar.gz "$i"
    done
    for i in `find . -type f -name "$appName-windows-*"`
    do
      zip compress/$(echo $i | sed 's/\.[^.]*$//').zip "$i"
    done
fi