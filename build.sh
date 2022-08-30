appName="alist"
builtAt="$(date +'%F %T %z')"
goVersion=$(go version | sed 's/go version //')
gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
gitCommit=$(git log --pretty=format:"%h" -1)

if [ "$1" = "release" ]; then
  version=$(git describe --long --tags --dirty --always)
  webVersion=$(wget -qO- -t1 -T2 "https://api.github.com/repos/alist-org/alist-web/releases/latest" | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g')
else
  version="dev"
  webVersion="dev"
fi

echo "build version: $gitTag"

ldflags="\
-w -s \
-X 'github.com/alist-org/alist/v3/internal/conf.BuiltAt=$builtAt' \
-X 'github.com/alist-org/alist/v3/internal/conf.GoVersion=$goVersion' \
-X 'github.com/alist-org/alist/v3/internal/conf.GitAuthor=$gitAuthor' \
-X 'github.com/alist-org/alist/v3/internal/conf.GitCommit=$gitCommit' \
-X 'github.com/alist-org/alist/v3/internal/conf.Version=$version' \
-X 'github.com/alist-org/alist/v3/internal/conf.WebVersion=$webVersion' \
"

FetchWebBuild() {
  curl -L https://codeload.github.com/alist-org/web-dist/tar.gz/refs/heads/main -o web-dist-main.tar.gz
  tar -zxvf web-dist-main.tar.gz
  rm -rf public/dist
  mv -f web-dist-main/dist public
  rm -rf web-dist-main web-dist-main.tar.gz
}

BuildDev() {
  rm -rf .git/
  xgo -targets=linux/amd64,windows/amd64,darwin/amd64 -out "$appName" -ldflags="$ldflags" -tags=jsoniter .
  mkdir -p "dist"
  mv alist-* dist
  cd dist
  upx -9 ./alist-linux*
  upx -9 ./alist-windows*
  find . -type f -print0 | xargs -0 md5sum >md5.txt
  cat md5.txt
  cd .. || exit
}

if [ "$1" = "dev" ]; then
  FetchWebBuild
  BuildDev
elif [ "$1" = "release" ]; then
  echo -e "To be implement"
else
  echo -e "Parameter error"
fi