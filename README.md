# Redeye Smart Network Camera

RedEye is a framework to build _Intellient Video Network_
applications. Easily build _Streaming Video_, _Computer Vision_ and
_Machine Learning_ features to your _Environmental_ and
_Industrial_ _IoT_ applications. 

## Requirements

Not much really. Start with a single Raspberry pi or build a network
of 12 Jetson Nano's with a nice fit control station.

RedEye is *smart* IP camera software built to run on inexpensive
computers with connected cameras like the Raspberry Pi with a CSI or USB
camera's. 

The idea is to be able to *control* a *network of cameras* providing
various video stream(s) to _Computer Vision_ algorithms.

## Usage

```
redeye [flags]

Flags:
  -device string        Camera device: index (0,1,…), name, or path (default "0")
  -rtsp   string        RTSP stream URL  (e.g. rtsp://192.168.1.10/stream)
  -video  string        Local video file path
  -image  string        Single static image file
  -pipeline string      Colon-separated filter pipeline (e.g. resize:0.5:face-detect)
  -addr   string        HTTP listen address (default "0.0.0.0:8080")
  -broker string        MQTT broker URL (e.g. tcp://localhost:1883); empty disables MQTT
  -topic-prefix string  MQTT topic namespace prefix (default "/redeye")
  -cascade-file string  Haar cascade XML for face detection
  -filters              List available filters and exit
```

Exactly one video source should be specified. When none is given, device `"0"` is used.

## Building

```
make          # same as make build
make build    # build for the current host
make test     # run the full test suite with the race detector
make coverage # run tests and print per-function coverage
make clean    # remove binaries and coverage artifacts
```

### Compiling for arm64 (Raspberry Pi 4 / 5)

`make rpi` cross-compiles a **static** arm64 binary targeting Raspberry Pi 4 and 5
(Cortex-A72 / A76, 64-bit).  The output is written to `redeye/redeye-rpi` so it
does not overwrite the local development binary.

**Prerequisites**

1. Install the aarch64 cross-compiler:

   ```
   sudo apt install gcc-aarch64-linux-gnu
   ```

2. Provide arm64 static OpenCV libraries.  The linker needs `.a` archives for
   OpenCV and its dependencies (e.g. from a sysroot or a pre-built static
   package).  Without them the link step will fail with unresolved symbols.

**Build**

```
make rpi
# produces redeye/redeye-rpi — copy to the Pi and run directly
```

**Dynamically-linked fallback**

If fully static linking is impractical (OpenCV static libs are large and
platform-specific), a dynamically-linked arm64 binary works fine on any Pi
that has OpenCV installed via `apt`:

```
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc \
  go build -o redeye/redeye-rpi ./redeye
```

Install OpenCV on the Pi before running:

```
sudo apt install libopencv-dev
```

## Video Sources

| Flag | Description |
|---|---|
| *(none)* | USB/V4L camera at device index 0 |
| `-device 0` | USB/V4L camera at device index N |
| `-device jetson` or `-device nano` | Jetson Nano CSI camera via GStreamer (`nvarguscamerasrc`) |
| `-device rpi` or `-device linux` | Raspberry Pi / Linux V4L2 camera (`/dev/video0`) |
| `-device mac` | macOS built-in FaceTime camera |
| `-device /dev/video2` | Explicit device path (passed through unchanged) |
| `-rtsp rtsp://host/path` | RTSP network stream (IP cameras, NVRs; `rtsps://` for TLS) |
| `-video file.mp4` | Local video file (MP4, AVI, MKV, …) |
| `-image file.jpg` | Single static image (window stays open until a key is pressed) |

### Platform Device Names

`-device` accepts a short name that is resolved to the correct device string or GStreamer pipeline automatically:

| Name | Resolves to |
|---|---|
| `jetson`, `nano` | `nvarguscamerasrc` GStreamer pipeline (1280×720 @ 60 fps) |
| `rpi`, `linux` | `/dev/video0` |
| `mac`, `default`, `0`, *(empty)* | `0` (OpenCV default device) |
| anything else | passed through unchanged |

## Working Features

1. Multiple video sources: USB camera, RTSP network stream, local video file, static image
2. Platform camera name aliases: `jetson`/`nano` (GStreamer), `rpi`/`linux` (`/dev/video0`), `mac`
3. MJPEG streaming over HTTP at `/mjpeg`
4. Configurable computer vision filter pipeline (resize, face detection)
5. REST API with `GET /health` liveness endpoint and `POST /api/camera/snap` snapshot capture
6. MQTT pub/sub for distributed multi-camera control
7. WebSocket endpoint (`/ws`) for real-time event push; embedded web UI at `/`
8. Config file support (`redeye.json` or `~/.redeye.json`, overridden by flags)

## REST API & WebSocket

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Embedded web UI (MJPEG stream + live event feed) |
| `GET` | `/health` | Liveness check — returns `{"status":"ok"}` |
| `GET` | `/mjpeg` | Live MJPEG video stream |
| `GET` | `/ws` | WebSocket — server pushes JSON events to the browser |
| `POST` | `/api/camera/snap` | Save current frame to disk |

### WebSocket events

Connect to `ws://<host>/ws` to receive a stream of JSON events:

```json
{"type": "connected"}
{"type": "detection", "payload": {"count": 2, "faces": [{"x":100,"y":80,"w":60,"h":60}]}}
```

| Type | When emitted | Payload |
|---|---|---|
| `connected` | On every new connection | *(none)* |
| `detection` | After each frame processed by `face-detect` filter | `{count, faces:[{x,y,w,h}]}` |

The web UI at `/` connects automatically and displays the MJPEG stream alongside a live event log.

### Snapshot

```
POST /api/camera/snap?file=myshot.jpg
```

- `file` query param sets the output path (default: `snapshot-YYYYMMDD-HHMMSS.jpg` in the working directory).
- Returns `{"file":"myshot.jpg"}` on success, or a JSON error body with an appropriate HTTP status code.
- Works with all source types: USB camera, RTSP stream, video file, and static image.

## MQTT Distributed Control

When `-broker` is set, redeye connects to the broker on startup, announces itself, and listens for JSON commands.

**Topics** (using default prefix `/redeye`):

| Direction | Topic | Purpose |
|---|---|---|
| Publish | `/redeye/announce/camera` | Node announces itself on connect |
| Subscribe | `/redeye/camera/<hostname>` | Receives commands addressed to this node |

**Command format:**

```json
{ "command": "snap", "file": "optional-output.jpg" }
```

| Command | Effect |
|---|---|
| `snap` | Saves the current frame; `file` is optional (defaults to `snapshot-<timestamp>.jpg`) |

**Example** — trigger a snapshot on a remote camera from any MQTT client:

```
mosquitto_pub -h broker.local -t /redeye/camera/mycamera -m '{"command":"snap","file":"/data/event.jpg"}'
```

## Near Term Roadmap

- Additional event types (motion, resize stats, frame rate)
- WebSocket inbound command support (trigger snap, change pipeline from browser)

## Supported Platforms

+ Raspberry Pi 3/4 + CSI Camera
+ Jetson Nano + CSI Camera
+ Ubuntu 19 Desktop + USB Cam (V4L)
+ Macbook Pro and Air + Built-in Camera
+ Any IP camera exposing an RTSP stream

### TODO OpenCV Plugin and Stream Only

+ Raspberry Pi Zero (stream only)
+ esp32 cam (st)

## OpenCV Plugin and Performance

RedEye is built with _OpenCV_ and hence takes advantage of the
powerful and flexible _device support_ provided by _OpenCV_. With
that, we get an amazing amount of power and flexibility right out of
the box, and do not have to do too much hard work to get there.

However, it does come at quite a footprint regarding memory, and the
build time on smaller devices is _ridiculous_ by todays standards (I
feel like a spoiled brat).

The idea then is to simply have the camera stream video to the A/I
module on another system. That requires the following to that going, 
Computer Vision module to read streaming video from network. 

That way, the smart module, can just suck the video down from a player
that only knows how to stream the video.


## OpenCV and Pipeline Plugins

+ Built with OpenCV
+ Video Pipeline plugins
  + Face detection

## APIs

### Camera Control

- Play
- Pause 
- Snap

### Camera Config

- Resolution
- Frames Per Second
- Format

### Storage

- Location
- GetClips
- GetSnapshots
- SaveClip
- SaveSnapshot

## Otto Discovery

+ Otto Discovery with MQTT
  + requires an MQTT broker
  + optional if broker is MQTT broker is NOT present


See Todo.org
