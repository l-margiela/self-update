NAME=self-update
GIT_COMMIT=${git rev-list -1 HEAD}
VERSION=1.0-${GIT_COMMIT}

run: build
	./self-update

build:
	go build -ldflags "-X main.Version=${VERSION}" -o ${NAME}
