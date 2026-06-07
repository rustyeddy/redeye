BINARY     := redeye/redeye
RPI_BINARY := redeye/redeye-rpi

# aarch64 cross-compiler: sudo apt install gcc-aarch64-linux-gnu
# Static OpenCV arm64 libs must be present for the linker.
RPI_CC := aarch64-linux-gnu-gcc

.PHONY: all build test coverage clean rpi

all: build

build:
	go build -o $(BINARY) ./redeye

test:
	go test -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm -f coverage.out

# Static arm64 binary for Raspberry Pi 4 / 5.
rpi:
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=arm64 \
	CC=$(RPI_CC) \
	go build \
	  -ldflags="-linkmode=external -extldflags='-static'" \
	  -o $(RPI_BINARY) \
	  ./redeye

clean:
	rm -f $(BINARY) $(RPI_BINARY) coverage.out
	go clean
