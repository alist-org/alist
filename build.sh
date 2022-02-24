#!/bin/bash

# 构建前端,在当前目录产生一个dist文件夹
BUILD_WEB() {
  git clone https://github.com/alist-org/alist-web.git
  cd alist-web
  yarn
  yarn build
  sed -i -e "s/\/CDN_URL\//\//g" dist/index.html
  sed -i -e "s/assets/\/assets/g" dist/index.html
  rm -f dist/index.html-e
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
  webTag=$(wget -qO- -t1 -T2 "https://api.github.com/repos/alist-org/alist-web/releases/latest" | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g')
  ldflags="\
-w -s \
-X 'github.com/Xhofe/alist/conf.BuiltAt=$builtAt' \
-X 'github.com/Xhofe/alist/conf.GoVersion=$goVersion' \
-X 'github.com/Xhofe/alist/conf.GitAuthor=$gitAuthor' \
-X 'github.com/Xhofe/alist/conf.GitCommit=$gitCommit' \
-X 'github.com/Xhofe/alist/conf.GitTag=$gitTag' \
-X 'github.com/Xhofe/alist/conf.WebTag=$webTag' \
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
  webTag=$(wget -qO- -t1 -T2 "https://api.github.com/repos/alist-org/alist-web/releases/latest" | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g')
  ldflags="\
-w -s \
-X 'github.com/Xhofe/alist/conf.BuiltAt=$builtAt' \
-X 'github.com/Xhofe/alist/conf.GoVersion=$goVersion' \
-X 'github.com/Xhofe/alist/conf.GitAuthor=$gitAuthor' \
-X 'github.com/Xhofe/alist/conf.GitCommit=$gitCommit' \
-X 'github.com/Xhofe/alist/conf.GitTag=$gitTag' \
-X 'github.com/Xhofe/alist/conf.WebTag=$webTag' \
  "
  if [ "$1" == "release" ]; then
    OS_ARCHES=("aix/ppc64" "android/386" "android/amd64" "android/arm" "android/arm64" "darwin/amd64" "darwin/arm64" "dragonfly/amd64" "freebsd/386" "freebsd/amd64" "freebsd/arm" "freebsd/arm64" "illumos/amd64" "ios/amd64" "ios/arm64" "js/wasm" "linux/386" "linux/amd64" "linux/arm" "linux/arm64" "linux/mips" "linux/mips64" "linux/mips64le" "linux/mipsle" "linux/ppc64" "linux/ppc64le" "linux/riscv64" "linux/s390x" "netbsd/386" "netbsd/amd64" "netbsd/arm" "netbsd/arm64" "openbsd/386" "openbsd/amd64" "openbsd/arm" "openbsd/arm64" "openbsd/mips64" "plan9/386" "plan9/amd64" "plan9/arm" "solaris/amd64" "windows/386" "windows/amd64" "windows/arm" "windows/arm64")
  else
    OS_ARCHES=("darwin/amd64" "linux/amd64" "windows/amd64")
  fi
  for i in "${!OS_ARCHES[@]}"; do
      os_arch=${OS_ARCHES[$i]}
      echo building for ${os_arch}
      export GOOS=${os_arch%%/*}
      export GOARCH=${os_arch##*/}
      export CGO_ENABLED=0
      if [ $GOOS == "windows" ]; then
        go build -o ./build/$appName-$GOOS-$GOARCH.exe -ldflags="$ldflags" -tags=jsoniter alist.go
      else
        go build -o ./build/$appName-$GOOS-$GOARCH -ldflags="$ldflags" -tags=jsoniter alist.go
      fi
  done
  cd build
  upx -9 ./*
  find . -type f -print0 | xargs -0 md5sum >md5.txt
  cat md5.txt
  cd ..
  cd ..
}

RELEASE() {
  cd alist/build
  mkdir compress
  mv md5.txt compress
#  win
  mkdir windows
  mv $appName-windows-* windows/
  cd windows
  for i in $(find . -type f -name "$appName-windows-*"); do
    zip ../compress/$(echo $i | sed 's/\.[^.]*$//').zip "$i"
  done
  cd ..
#  end win
  for i in $(find . -type f -name "$appName-*"); do
    tar -czvf compress/"$i".tar.gz "$i"
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
