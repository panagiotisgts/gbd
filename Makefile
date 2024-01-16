build:
	go build -ldflags="-X 'main.version=${VERSION}'" -o bin/gbd ./cmd/gbd