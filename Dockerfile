FROM golang:1.23-alpine AS build
WORKDIR /workspace
COPY . .
RUN go build -o build/gologin-test-app .

FROM scratch
COPY --from=build /workspace/build/gologin-test-app /bin/gologin-test-app
