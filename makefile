build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o cisco-socks5 cmd/*
	chmod +x cisco-socks5

build:
	go build -o cisco-socks5 cmd/*
	chmod +x cisco-socks5