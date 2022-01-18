#!/bin/bash

# 构建前端,在当前目录产生一个dist文件夹
BUILD_WEB() {
  git clone https://github.com/alist-org/alist-web.git
  cd alist-web
  yarn
  yarn build
  mv dist ..
  cd ..
  rm -rf alist-web
}

CDN_WEB() {
  curl -L https://github.com/alist-org/alist-web/releases/latest/download/dist.tar.gz -o dist.tar.gz
  tar -zxvf dist.tar.gz
  rm -f dist.tar.gz
}

# 在DOCKER中构建
BUILD_DOCKER() {
  appName="alist"
  builtAt="$(date +'%F %T %z')"
  goVersion=$(go version | sed 's/go version //')
  gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
  gitCommit=$(git log --pretty=format:"%h" -1)
  gitTag=$(git describe --long --tags --dirty --always)
  ldflags="\
-w -s \
-X 'github.com/Xhofe/alist/conf.BuiltAt=$builtAt' \
-X 'github.com/Xhofe/alist/conf.GoVersion=$goVersion' \
-X 'github.com/Xhofe/alist/conf.GitAuthor=$gitAuthor' \
-X 'github.com/Xhofe/alist/conf.GitCommit=$gitCommit' \
-X 'github.com/Xhofe/alist/conf.GitTag=$gitTag' \
  "
  go build -o ./bin/alist -ldflags="$ldflags" -tags=jsoniter alist.go
}

BUILD() {
  cd alist
  appName="alist"
  builtAt="$(date +'%F %T %z')"
  goVersion=$(go version | sed 's/go version //')
  gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
  gitCommit=$(git log --pretty=format:"%h" -1)
  gitTag=$(git describe --long --tags --dirty --always)
  echo "build version: $gitTag"
  ldflags="\
-w -s \
-X 'github.com/Xhofe/alist/conf.BuiltAt=$builtAt' \
-X 'github.com/Xhofe/alist/conf.GoVersion=$goVersion' \
-X 'github.com/Xhofe/alist/conf.GitAuthor=$gitAuthor' \
-X 'github.com/Xhofe/alist/conf.GitCommit=$gitCommit' \
-X 'github.com/Xhofe/alist/conf.GitTag=$gitTag' \
"

  if [ "$1" == "release" ]; then
    xgo -out "$appName" -ldflags="$ldflags" -tags=jsoniter .
  else
    xgo -targets=linux/amd64,windows/amd64 -out alist -ldflags="$ldflags" -tags=jsoniter .
  fi
  mkdir "build"
  mv alist-* build
  cd build
  upx -9 ./*
  find . -type f -print0 | xargs -0 md5sum >md5.txt
  cat md5.txt
  cd ../..
}

RELEASE() {
  cd alist/build
  mkdir compress
  mv md5.txt compress
  for i in $(find . -type f -name "$appName-linux-*"); do
    tar -czvf compress/"$i".tar.gz "$i"
  done
  for i in $(find . -type f -name "$appName-darwin-*"); do
    tar -czvf compress/"$i".tar.gz "$i"
  done
  for i in $(find . -type f -name "$appName-windows-*"); do
    zip compress/$(echo $i | sed 's/\.[^.]*$//').zip "$i"
  done
  cd ../..
}

if [ "$1" = "web" ]; then
  BUILD_WEB
elif [ "$1" = "cdn" ]; then
  CDN_WEB
elif [ "$1" = "docker" ]; then
  BUILD_DOCKER
elif [ "$1" = "build" ]; then
  BUILD build
elif [ "$1" = "release" ]; then
  BUILD release
  RELEASE
else
  echo -e "${RED_COLOR} Parameter error ${RES}"
fi
