#!/bin/bash

cd alist-web || exit
webCommit=$(git log --pretty=format:"%h" -1)
echo "web commit id: $webCommit"
yarn
if [ "$1" == "release" ]; then
  yarn build --base="https://cdn.jsdelivr.net/gh/Xhofe/alist-web@cdn/v2/$webCommit"
  mv dist/assets ..
else
  yarn build
fi
cd ..

cd alist
appName="alist"
builtAt="$(date +'%F %T %z')"
goVersion=$(go version | sed 's/go version //')
gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
gitCommit=$(git log --pretty=format:"%h" -1)

if [ "$1" == "release" ]; then
  gitTag=$(git describe --abbrev=0 --tags)
else
  gitTag=build-next
fi

echo "build version: $gitTag"

ldflags="\
-w -s \
-X 'github.com/Xhofe/alist/conf.BuiltAt=$builtAt' \
-X 'github.com/Xhofe/alist/conf.GoVersion=$goVersion' \
-X 'github.com/Xhofe/alist/conf.GitAuthor=$gitAuthor' \
-X 'github.com/Xhofe/alist/conf.GitCommit=$gitCommit' \
-X 'github.com/Xhofe/alist/conf.GitTag=$gitTag' \
"

cp -R ../alist-web/dist/* public

if [ "$1" == "release" ]; then
  xgo -out alist -ldflags="$ldflags" .
else
  xgo -targets=linux/amd64,windows/amd64,darwin/amd64 -out alist -ldflags="$ldflags" .
fi
mkdir "build"
mv alist-* build
cd build || exit
upx -9 ./*
find . -type f -print0 | xargs -0 md5sum > md5.txt
cat md5.txt
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
cd ../..

if [ "$1" == "release" ]; then
  cd alist-web
  git checkout cdn
  mkdir "v2/$webCommit"
  mv ../assets/ v2/$webCommit
  git add .
  git config --local user.email "i@nn.ci"
  git config --local user.name "Xhofe"
  git commit --allow-empty -m "upload $webCommit assets files" -a
  cd ..
fi