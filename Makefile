generate:
	GOOS=windows GOARCH=amd64 go build -o ./build/virus-amd64.exe ./cmd && \
	GOOS=linux GOARCH=amd64 go build -o ./build/virus-amd64-linux ./cmd && \
	GOOS=darwin GOARCH=arm64 go build -o ./build/virus-arm64-darwin ./cmd && \
	GOOS=darwin GOARCH=amd64 go build -o ./build/virus-amd64-darwin ./cmd
run:
	go run ./cmd
	