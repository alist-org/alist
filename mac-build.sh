appName="alist"
builtAt="$(date +'%F %T %z')"
goVersion=1.20
gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
gitCommit=$(git log --pretty=format:"%h" -1)
version=$(git describe --long --tags --dirty --always)
webVersion=$(wget -qO- -t1 -T2 "https://api.github.com/repos/alist-org/alist-web/releases/latest" | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g')
ldflags="\
-w -s \
-X 'github.com/alist-org/alist/v3/internal/conf.BuiltAt=$builtAt' \
-X 'github.com/alist-org/alist/v3/internal/conf.GoVersion=$goVersion' \
-X 'github.com/alist-org/alist/v3/internal/conf.GitAuthor=$gitAuthor' \
-X 'github.com/alist-org/alist/v3/internal/conf.GitCommit=$gitCommit' \
-X 'github.com/alist-org/alist/v3/internal/conf.Version=$version' \
-X 'github.com/alist-org/alist/v3/internal/conf.WebVersion=$webVersion' \
"
echo -e $ldflags
export CGO_LDFLAGS="-static" && export GOROOT=/Users/lowking/go-sdk/go$goVersion && export GOPATH=/Users/lowking/Desktop && GOOS=linux && export GOARCH=amd64 && export CC=x86_64-linux-gnu-gcc && export CGO_ENABLED=1 && ~/go-sdk/go$goVersion/bin/go build -o ~/alist/build/alist -tags=jsoniter -ldflags="$ldflags" .
ssh root@10.0.0.205 "systemctl stop alist" && scp /Users/lowking/alist/build/alist root@10.0.0.205:/root/ && ssh root@10.0.0.205 "systemctl start alist"
