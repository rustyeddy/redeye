package redeye

import "strconv"

// CamStr maps a human-readable platform name or device shorthand to the
// device string that OpenCV understands.
//
// Known names:
//
//	"jetson", "nano"      Jetson Nano CSI camera via GStreamer nvarguscamerasrc (1280×720 @ 60 fps)
//	"rpi", "linux"        Raspberry Pi or generic Linux V4L2 camera (/dev/video0)
//	"mac", "default", ""  Default device index ("0")
//
// Any other string is returned unchanged, so callers can pass raw device
// paths ("/dev/video2"), numeric indexes ("1"), or full GStreamer pipelines.
func CamStr(name string) string {
	switch name {
	case "jetson", "nano":
		return jetsonCamStr(0, 1280, 720, 60, 0)
	case "rpi", "linux":
		return "/dev/video0"
	case "mac", "default", "0", "":
		return "0"
	default:
		return name
	}
}

// jetsonCamStr builds the nvarguscamerasrc GStreamer pipeline for the
// Jetson Nano CSI camera. Parameters mirror the nvarguscamerasrc caps:
// sensorID selects the camera port; width/height/fps set the capture
// resolution and frame rate; flipMethod rotates the image (0 = none).
func jetsonCamStr(sensorID, width, height, fps, flipMethod int) string {
	w := strconv.Itoa(width)
	h := strconv.Itoa(height)
	f := strconv.Itoa(fps)
	fl := strconv.Itoa(flipMethod)
	sid := strconv.Itoa(sensorID)

	return "nvarguscamerasrc sensor_id=" + sid +
		" ! video/x-raw(memory:NVMM)" +
		", width=(int)" + w +
		", height=(int)" + h +
		", format=(string)NV12" +
		", framerate=(fraction)" + f + "/1" +
		" ! nvvidconv flip-method=" + fl +
		" ! video/x-raw" +
		", width=(int)" + w +
		", height=(int)" + h +
		", format=(string)BGRx" +
		" ! videoconvert" +
		" ! video/x-raw, format=(string)BGR" +
		" ! appsink"
}
