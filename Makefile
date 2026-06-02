BINARY := redeye/redeye
PKGS   := github.com/rustyeddy/redeye github.com/rustyeddy/redeye/filters

.PHONY: all build test coverage clean rpi nano

all: build

build:
	go build -o $(BINARY) ./redeye

test:
	go test -race $(PKGS)

coverage:
	go test -coverprofile=coverage.out $(PKGS)
	go tool cover -func=coverage.out
	@rm -f coverage.out

rpi:
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(BINARY) ./redeye

nano:
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(BINARY) ./redeye

clean:
	rm -f $(BINARY) coverage.out
	go clean
