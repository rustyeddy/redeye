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
  -plugins string       Directory of .so filter plugins to load at startup (default "plugins")
```

Exactly one video source should be specified. When none is given, device `"0"` is used.

## Filter Pipeline

Redeye processes each video frame through an ordered pipeline of filters.
Filters are applied left-to-right in the order they were added.

### Built-in filters

| Name | Description |
|---|---|
| `resize` | Scale the frame by X and Y factors (default 1.0) |
| `face-detect` | Haar cascade face detection; emits WebSocket detection events |
| `color-detect` | Colour detection stub (not yet implemented) |
| `textzoom` | Text zoom overlay |

### Specifying a pipeline

Pass a colon-separated list to `-pipeline`. Filter names and their config
values share the same colon delimiter; a token that matches a registered
filter name starts a new filter, anything else is treated as config for the
current one:

```
redeye -pipeline resize:0.5:face-detect
#  → resize.Init("0.5"), face-detect.Init("")

redeye -pipeline resize:0.5:0.75:face-detect
#  → resize.Init("0.5:0.75"), face-detect.Init("")
```

### Runtime pipeline control

The active pipeline can be changed at any time without restarting:

| Method | Path | Body | Effect |
|---|---|---|---|
| `GET` | `/api/camera/pipeline` | — | Return current pipeline descriptor |
| `POST` | `/api/camera/pipeline` | `filter=<name>` | Toggle a filter on or off |
| `POST` | `/api/camera/pipeline` | `pipeline=<str>` | Replace the entire pipeline |
| `POST` | `/api/camera/pipeline` | `pipeline=` | Clear the pipeline |

The web UI sends these automatically when you click filter toggle buttons.

### Filter parameters

Filters that implement the `Parametric` interface expose runtime-adjustable
parameters as inline sliders in the web UI.  Parameter changes are sent via:

```
POST /api/camera/filter/param
  filter=<name>    name of the filter
  <key>=<value>    one field per parameter (float values)
```

**`resize` parameters:**

| Key | Label | Range | Default |
|---|---|---|---|
| `x` | Scale X | 0.1 – 2.0 | 1.0 |
| `y` | Scale Y | 0.1 – 2.0 | 1.0 |

New filters gain automatic UI controls by implementing two methods:

```go
func (f *MyFilter) Params() []redeye.ParamDesc { ... }
func (f *MyFilter) SetParam(key string, value float64) error { ... }
```

## Building

```
make          # same as make build
make build    # build for the current host
make test     # run the full test suite with the race detector
make coverage # run tests and print per-function coverage
make clean    # remove binaries and coverage artifacts
make plugins  # build all filter plugins as .so shared libraries
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

### Filter plugins

Filters can be compiled as Go shared libraries (`.so`) and loaded at runtime
without recompiling the host binary.  This lets you ship or iterate on filters
independently.

**Constraint:** the plugin and the host binary must be compiled with the **same
Go toolchain version** and **identical dependency versions**.  Build both on the
same machine.  The static RPi cross-build (`make rpi`) is incompatible with
plugins — use the dynamically-linked fallback instead.

#### Building plugins

```bash
make plugins              # builds all plugins under plugins/
# or individually:
go build -buildmode=plugin -tags plugin -o plugins/grayscale.so ./plugins/grayscale
```

Drop the resulting `.so` files into the `plugins/` directory (or whichever path
is given to `--plugins`).  They are loaded automatically on the next startup.

#### Hot-loading at runtime

Load a plugin without restarting the server:

```bash
curl -X POST http://localhost:9382/api/plugins/load \
     -H 'Content-Type: application/json' \
     -d '{"path": "plugins/grayscale.so"}'
```

The newly registered filters appear immediately in the web UI filter list.

#### Writing a plugin

Create a new directory under `plugins/` with a `main.go`:

```go
//go:build plugin

package main

import "github.com/rustyeddy/redeye"

type MyFilter struct{ redeye.Flt }

func (f *MyFilter) Init(_ string)                      {}
func (f *MyFilter) Filter(fr *redeye.Frame) *redeye.Frame { /* process fr.Mat */ return fr }

func init() {
    redeye.Filters.Add("my-filter", &MyFilter{
        Flt: redeye.NewFlt("my-filter", "one-line description"),
    })
}
```

The `//go:build plugin` tag keeps the file out of regular `go build ./...` runs.
The `init()` function is the only required contract — it is called automatically
when the host opens the `.so` with `plugin.Open`.

To expose runtime-adjustable parameters, also implement `Parametric`:

```go
func (f *MyFilter) Params() []redeye.ParamDesc {
    return []redeye.ParamDesc{
        {Key: "strength", Label: "Strength", Type: "float", Min: 0, Max: 1, Step: 0.05, Value: f.strength},
    }
}
func (f *MyFilter) SetParam(key string, v float64) error {
    if key == "strength" { f.strength = v; return nil }
    return fmt.Errorf("unknown param %q", key)
}
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

## Web UI

Redeye ships a browser UI served at `/`. It is built with **[htmx](https://htmx.org)** for
server-driven controls and vanilla JS for the live event stream.

### Layout

```
┌─────────────────────────────┬──────────────────────┐
│  Live MJPEG stream          │  Snapshot            │
│                             │  [filename]  [Snap]  │
│                             │  ─────────────────── │
│                             │  Filters             │
│                             │  (loaded on demand)  │
├─────────────────────────────┴──────────────────────┤
│  ▶ Configuration  (click to expand)                │
│    Pipeline, MQTT, cascade file, log settings      │
├────────────────────────────────────────────────────┤
│  Live Events  (WebSocket, auto-reconnect)          │
└────────────────────────────────────────────────────┘
```

### Content negotiation

All UI routes are shared with the JSON API. The server detects htmx requests via
the `HX-Request: true` header and returns an HTML fragment; regular API clients
receive JSON as before.

| Route | API (JSON) | Browser (HTML fragment) |
|---|---|---|
| `GET /api/filters` | JSON array of filter names | Rendered filter list |
| `GET /api/camera/config` | JSON config object | Populated config form |
| `PUT /api/camera/config` | JSON body → JSON response | Form body → re-rendered form |
| `POST /api/camera/snap` | `?file=` query param | `name="file"` form field |

### Templates

HTML fragments live in `templates/*.html` and are embedded into the binary at
build time via `//go:embed`. The Go handler renders them with `html/template`
before sending the response.

| Template | Data type | Purpose |
|---|---|---|
| `templates/snap.html` | `snapData{File, Error}` | Snap success / error feedback |
| `templates/config.html` | `configFormData{Config, Message}` | Config form (GET + after PUT) |
| `templates/filters.html` | `[]filterEntry{Name, Desc}` | Available filter list |

### Offline deployments

The UI loads htmx from the unpkg CDN. On devices without internet access,
download the file and serve it locally:

```bash
curl -o static/htmx.min.js https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js
```

Then change the `<script src>` in `static/index.html` to `/static/htmx.min.js`
and add a route in `RegisterAPIRoutes`:

```go
mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
```

## Working Features

1. Multiple video sources: USB camera, RTSP network stream, local video file, static image
2. Platform camera name aliases: `jetson`/`nano` (GStreamer), `rpi`/`linux` (`/dev/video0`), `mac`
3. MJPEG streaming over HTTP at `/mjpeg`
4. Configurable computer vision filter pipeline (resize, face detection) with runtime toggle and reorder
5. Per-filter runtime parameter controls (sliders in the web UI, REST API, or MQTT)
6. Go plugin system — load `.so` filter plugins at startup or hot-load without restart
7. REST API with `GET /health` liveness endpoint and `POST /api/camera/snap` snapshot capture
8. MQTT pub/sub for distributed multi-camera control
9. WebSocket endpoint (`/ws`) for real-time event push; embedded web UI at `/`
10. Config file support (`redeye.json` or `~/.redeye.json`, overridden by flags)

## REST API & WebSocket

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Embedded web UI (MJPEG stream + live event feed) |
| `GET` | `/health` | Liveness check — returns `{"status":"ok"}` |
| `GET` | `/mjpeg` | Live MJPEG video stream |
| `GET` | `/ws` | WebSocket — server pushes JSON events to the browser |
| `POST` | `/api/camera/snap` | Save current frame to disk |
| `GET` | `/api/camera/pipeline` | Return current pipeline descriptor string |
| `POST` | `/api/camera/pipeline` | Toggle a filter or replace the entire pipeline |
| `POST` | `/api/camera/filter/param` | Update one or more parameters on a Parametric filter |
| `POST` | `/api/plugins/load` | Hot-load a `.so` filter plugin at runtime |

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
