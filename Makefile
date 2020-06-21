NAME=self-update
GIT_COMMIT=`git rev-parse --short HEAD`
VERSION=1.1+${GIT_COMMIT}
GOPATH=$(shell go env GOPATH)

run: build
	./dist/self-update -dev

build:
	-mkdir dist
	go build -ldflags "-X main.Version=${VERSION}" -o dist/${NAME}

build-windows:
	-mkdir dist
	GOOS=windows go build -ldflags "-X main.Version=${VERSION}" -o dist/${NAME}.exe

lint:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOPATH)/bin v1.27.0
	${GOPATH}/bin/golangci-lint run