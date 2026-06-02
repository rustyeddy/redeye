package redeye

import (
	"strings"
	"testing"
)

func TestCamStrKnownPlatforms(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// Numeric defaults — returned as the bare index string
		{"0", "0"},
		{"default", "0"},
		{"", "0"},

		// Mac built-in camera
		{"mac", "0"},

		// Linux V4L2
		{"linux", "/dev/video0"},
		{"rpi", "/dev/video0"},

		// Jetson / nano return a GStreamer pipeline
		{"jetson", "nvarguscamerasrc"},
		{"nano", "nvarguscamerasrc"},
	}

	for _, tt := range tests {
		got := CamStr(tt.name)
		if !strings.HasPrefix(got, tt.want) && got != tt.want {
			t.Errorf("CamStr(%q) = %q; want string starting with %q", tt.name, got, tt.want)
		}
	}
}

func TestCamStrJetsonContainsRequiredElements(t *testing.T) {
	pipeline := CamStr("jetson")
	for _, substr := range []string{
		"nvarguscamerasrc",
		"nvvidconv",
		"videoconvert",
		"appsink",
		"format=(string)BGR",
	} {
		if !strings.Contains(pipeline, substr) {
			t.Errorf("jetson pipeline missing %q\npipeline: %s", substr, pipeline)
		}
	}

	// nano must produce the same pipeline as jetson
	if CamStr("nano") != CamStr("jetson") {
		t.Error("CamStr(\"nano\") and CamStr(\"jetson\") should return the same pipeline")
	}
}

func TestCamStrPassThroughUnknownNames(t *testing.T) {
	passThrough := []string{
		"1",
		"2",
		"/dev/video1",
		"/dev/video2",
		"nvarguscamerasrc ! appsink",
		"rtsp://camera.local/stream",
		"some-unknown-platform",
	}
	for _, name := range passThrough {
		got := CamStr(name)
		if got != name {
			t.Errorf("CamStr(%q) = %q; want pass-through %q", name, got, name)
		}
	}
}

func TestJetsonCamStrParametersAppearInPipeline(t *testing.T) {
	pipeline := jetsonCamStr(1, 1920, 1080, 30, 2)
	for _, substr := range []string{
		"sensor_id=1",
		"width=(int)1920",
		"height=(int)1080",
		"framerate=(fraction)30/1",
		"flip-method=2",
	} {
		if !strings.Contains(pipeline, substr) {
			t.Errorf("jetsonCamStr: missing %q\npipeline: %s", substr, pipeline)
		}
	}
}
