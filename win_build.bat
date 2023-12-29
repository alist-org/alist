@echo off

REM ============= build script for windows ================
REM how to use
REM win_build.bat v0.0.1
REM =======================================================

REM ============= variable definitions ================
set currentDir=%CD%
set output=build
set name=alist
set version=%1

REM ============= build action ================
@REM call :build_task %name%-%version%-windows-x64 windows amd64
@REM call :build_task %name%-%version%-windows-arm windows arm
call :build_task %name%-%version%-linux-amd64 linux amd64
@REM call :build_task %name%-%version%-darwin-macos-amd64 darwin amd64
@REM call :build_task %name%-%version%-darwin-macos-arm64 darwin arm64

goto:EOF

REM ============= build function ================
:build_task
setlocal

set targetName=%1
set GOOS=%2
set GOARCH=%3
set goarm=%4
set GO386=sse2
set CGO_ENABLED=0
set GOARM=%goarm%

echo "Building %targetName% ..."
if %GOOS% == windows (
  go build -ldflags "-s -w" -o "%output%/%1/%name%.exe"
) ^
else (
  go build -ldflags "-s -w" -o "%output%/%1/%name%"
)

endlocal

