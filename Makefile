all:
	@GOOS=linux GOARCH=amd64 go build -v -o githubssh-linux-amd64 -ldflags "-s -w"
	@GOOS=linux GOARCH=arm64 go build -v -o githubssh-linux-arm64 -ldflags "-s -w"
	@GOOS=darwin GOARCH=amd64 go build -v -o githubssh-macos-amd64 -ldflags "-s -w"
	@GOOS=darwin GOARCH=arm64 go build -v -o githubssh-macos-arm64 -ldflags "-s -w"