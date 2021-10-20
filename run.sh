GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -ldflags "-X main.version=test" -0 bin/linux/pwc-controller
docker build --no-cache -f Dockerfile -t kmc-pwc-controller-test.latest .