VERSION 0.8

build:
    FROM golang:1.24.1

    RUN apt-get update && apt-get install -y --no-install-recommends musl-dev musl-tools

    ARG VERSION=v0.3.0

    RUN git config --global advice.detachedHead false
    RUN git clone https://github.com/ngyewch/gologin-test-app.git
    WORKDIR gologin-test-app
    RUN git checkout ${VERSION}
    RUN CGO_ENABLED=1 CC=musl-gcc go build --ldflags '-linkmode=external -extldflags=-static' -o gologin-test-app ./

    SAVE ARTIFACT gologin-test-app

docker:
    FROM scratch

    ARG VERSION=v0.3.0

    COPY (+build/gologin-test-app --VERSION=${VERSION}) /usr/local/bin/

    ENTRYPOINT ["/usr/local/bin/gologin-test-app"]

    SAVE IMAGE gologin-test-app:${VERSION}
