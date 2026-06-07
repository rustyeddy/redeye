BINARY     := redeye/redeye
RPI_BINARY := redeye/redeye-rpi

# aarch64 cross-compiler: sudo apt install gcc-aarch64-linux-gnu
# Static OpenCV arm64 libs must be present for the linker.
RPI_CC := aarch64-linux-gnu-gcc

# Plugin .so files — add new entries here as new plugins are created.
PLUGIN_DIR  := plugins
PLUGIN_SOS  := $(PLUGIN_DIR)/grayscale.so

.PHONY: all build plugins test coverage clean rpi

all: build

build:
	go build -o $(BINARY) ./redeye

# Build all filter plugins as shared libraries.
# Plugins require a CGO-enabled, dynamically linked host (not compatible with
# the static RPi cross-build). Both the host and plugins must be compiled with
# the same Go toolchain and dependency versions.
plugins: $(PLUGIN_SOS)

$(PLUGIN_DIR)/grayscale.so:
	go build -buildmode=plugin -tags plugin -o $@ ./$(PLUGIN_DIR)/grayscale

test:
	go test -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm -f coverage.out

# Static arm64 binary for Raspberry Pi 4 / 5.
# Note: static binaries cannot load plugins at runtime.
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
	rm -f $(BINARY) $(RPI_BINARY) $(PLUGIN_SOS) coverage.out
	go clean
