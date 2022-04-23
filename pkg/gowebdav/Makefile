BIN := gowebdav
SRC := $(wildcard *.go) cmd/gowebdav/main.go

all: test cmd

cmd: ${BIN}

${BIN}: ${SRC}
	go build -o $@ ./cmd/gowebdav

test:
	go test -v --short ./...

api:
	@sed '/^## API$$/,$$d' -i README.md
	@echo '## API' >> README.md
	@godoc2md github.com/studio-b12/gowebdav | sed '/^$$/N;/^\n$$/D' |\
	sed '2d' |\
	sed 's/\/src\/github.com\/studio-b12\/gowebdav\//https:\/\/github.com\/studio-b12\/gowebdav\/blob\/master\//g' |\
	sed 's/\/src\/target\//https:\/\/github.com\/studio-b12\/gowebdav\/blob\/master\//g' |\
	sed 's/^#/##/g' >> README.md

check:
	gofmt -w -s $(SRC)
	@echo
	gocyclo -over 15 .
	@echo
	golint ./...

clean:
	@rm -f ${BIN}

.PHONY: all cmd clean test api check
