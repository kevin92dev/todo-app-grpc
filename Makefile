clean:
	@rm -rf dist
	@mkdir -p dist

build: clean
	go build -o dist/server cmd/server/main.go
	go build -o dist/client cmd/client/main.go