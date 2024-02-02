build:
	go build -ldflags="-X 'main.version=${VERSION}'" -o bin/gbd ./cmd/gbd

test-unit:
	go test -v -tags unit ./...

test-docker:
	go test -v -tags docker ./...