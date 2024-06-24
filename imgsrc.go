package redeye

import (
	"fmt"
	"time"

	"gocv.io/x/gocv"
)

// ImgSrc is an interface for all redeye sources including cameras,
// video and image files.
type ImgSrc interface {
	Play() chan *Frame
	IsRunning() bool
	Close() error
}

// Cam is a concrete datatype for a camera, the Cam struct will obtain
// a FrameBuffer of size 10 by default and open a channel to deliver
// incoming frames
type Cam struct {
	BufferSize int

	devID   int
	cap     *gocv.VideoCapture
	frameQ  chan *Frame
	running bool
}

// GetCam will open the Camera device of the given deviceID and create
// the FrameQ channel to start sending Frames on
func GetCam(deviceID int) (cam *Cam, err error) {
	cam = &Cam{
		devID:      deviceID,
		BufferSize: 10,
	}
	cam.cap, err = gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, err
	}

	return cam, nil
}

// IsRunning will tell us if the camera is active delivering images
func (cam *Cam) IsRunning() bool {
	return cam.running
}

// Play will start reading images from the OpenCV frame device and
// start queing them up on the frame channel after doing a quick
// sanity check to ensure there are infact an images to be read.
func (cam *Cam) Play() chan *Frame {
	cam.running = true
	cam.frameQ = make(chan *Frame)

	frames := GetFrameBuffers(cam.BufferSize)
	go func() {
		for cam.running {
			time.Sleep(5 * time.Millisecond)

			frame := frames.Next()
			cam.cap.Read(frame.Mat)
			if frame.Mat.Empty() {
				continue
			}
			size := frame.Mat.Size()
			if size[0] <= 0 || size[1] <= 0 {
				continue
			}
			cam.frameQ <- &frame
		}
		close(cam.frameQ)
	}()

	return cam.frameQ
}

// Close stops reading images from the capture device and closes down
// the FrameQ channel
func (cam *Cam) Close() error {
	cam.running = false
	cam.Close()
	close(cam.frameQ)
	return nil
}

// Img will read a single file image and queue it up on the
// FrameQ. The file and subsequently the FrameQ will be closed
type Img struct {
	Filename string
	gocv.IMReadFlag

	frame  *Frame
	frameQ chan *Frame

	running bool
}

// GetImg opens the image file fname and returns the image
// processing and/or viewing.
func GetImg(fname string) (img *Img, err error) {
	//m := gocv.IMRead(fname, gocv.IMReadUnchanged)
	m := gocv.IMRead(fname, gocv.IMReadColor)
	if m.Empty() {
		return nil, fmt.Errorf("ERROR reading %s", fname)
	}

	f := NewFrame()
	img = &Img{Filename: fname, frame: &f}
	img.frame.Mat = &m
	return img, nil
}

// IsRunning tells us if the image is being displayed, it
// also fulfills the ImgSrc interface
func (i *Img) IsRunning() bool {
	return i.running
}

// Play will copy the opened frame into the pipeline
func (i *Img) Play() chan *Frame {
	i.frameQ = make(chan *Frame)
	i.running = true

	go func() {
		time.Sleep(10 * time.Millisecond)
		i.frameQ <- i.frame
		// i.running = false
	}()
	return i.frameQ
}

// Close the image
func (i *Img) Close() error {
	i.running = false
	i.frame.Mat.Close()
	close(i.frameQ)

	return nil
}

// VideoFile does as it's name indicates, reads a video from
// a file and runs it through the pipeline
type VideoFile struct {
	*gocv.VideoCapture

	fname      string
	frameQ     chan *Frame
	running    bool
	bufferSize int
}

// GetVideo reads the file specified by the fname and prepares
// to run it through
func GetVideo(fname string) (vid *VideoFile, err error) {
	vid = &VideoFile{}
	vid.VideoCapture, err = gocv.VideoCaptureFile(fname)
	vid.bufferSize = 5

	return vid, err
}

func (v *VideoFile) IsRunning() bool {
	return v.running
}

func (v *VideoFile) Play() chan *Frame {
	v.frameQ = make(chan *Frame)
	v.running = true

	frames := GetFrameBuffers(v.bufferSize)
	go func() {
		for v.running {
			time.Sleep(10 * time.Millisecond)

			frame := frames.Next()
			v.VideoCapture.Read(frame.Mat)
			if frame.Mat.Empty() {
				continue
			}
			size := frame.Mat.Size()
			if size[0] <= 0 || size[1] <= 0 {
				continue
			}
			v.frameQ <- &frame
		}
		close(v.frameQ)
	}()

	return v.frameQ
}

func (v *VideoFile) Close() error {
	return v.VideoCapture.Close()
}
