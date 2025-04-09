build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o warp-server cmd/app/*
	chmod +x warp-server

build:
	go build -o warp-server cmd/app/*
	chmod +x warp-server