REM CDN_WEB
curl -L https://github.com/alist-org/alist-web/releases/latest/download/dist.tar.gz -o dist.tar.gz
tar -zxvf dist.tar.gz
xcopy /e /y .\dist\* .\public\
rmdir /s /q .\dist
del dist.tar.gz

REM BUILD_DOCKER
pwsh -NoProfile -Command "(Get-Content build.sh -Encoding UTF8).Replace(\"`n`r\",\"`n\") | Out-File -Encoding UTF8NoBOM build.sh"
docker build .
