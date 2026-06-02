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
  -device string     Camera device: index (0,1,…), name, or path (default "0")
  -rtsp   string     RTSP stream URL  (e.g. rtsp://192.168.1.10/stream)
  -video  string     Local video file path
  -image  string     Single static image file
  -pipeline string   Colon-separated filter pipeline (e.g. resize:0.5:face-detect)
  -addr   string     HTTP listen address (default "0.0.0.0:8080")
  -cascade-file string  Haar cascade XML for face detection
  -filters           List available filters and exit
```

Exactly one video source should be specified. When none is given, device `"0"` is used.

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
5. REST API scaffolding with `GET /health` liveness endpoint
6. Config file support (`redeye.json` or `~/.redeye.json`, overridden by flags)

## Near Term Roadmap

- Snapshot capture via `POST /api/camera/snap`
- Platform camera string helpers (Jetson Nano CSI, Raspberry Pi)
- MQTT messaging for distributed multi-camera control
- WebSocket endpoint for real-time event push to browser clients

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
