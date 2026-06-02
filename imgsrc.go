package redeye

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
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

// snapReq is sent on the snap channel to request that the next valid
// frame be written to disk. The goroutine sends the result on reply.
type snapReq struct {
	file  string
	reply chan error
}

// Cam is a concrete datatype for a camera, the Cam struct will obtain
// a FrameBuffer of size 10 by default and open a channel to deliver
// incoming frames
type Cam struct {
	BufferSize int

	device string
	cap    *gocv.VideoCapture
	frames *FrameBuffers
	frameQ chan *Frame
	quit   chan struct{}
	snapCh chan snapReq
	wg     sync.WaitGroup
}

// GetCam opens a camera by device descriptor and returns a Cam ready to Play.
// The descriptor is resolved through CamStr first, so platform names
// ("jetson", "nano", "rpi", "linux", "mac") and raw device paths
// ("/dev/video2") are all accepted in addition to plain integer indexes.
func GetCam(device string) (cam *Cam, err error) {
	cam = &Cam{
		device:     device,
		BufferSize: 10,
		quit:       make(chan struct{}),
		snapCh:     make(chan snapReq, 1),
	}

	camstr := CamStr(device)
	if devID, convErr := strconv.Atoi(camstr); convErr == nil {
		cam.cap, err = gocv.VideoCaptureDevice(devID)
	} else {
		// GStreamer pipeline or device path (e.g. /dev/video0)
		cam.cap, err = gocv.VideoCaptureFile(camstr)
	}
	if err != nil {
		return nil, err
	}

	return cam, nil
}

// IsRunning will tell us if the camera is active delivering images
func (cam *Cam) IsRunning() bool {
	select {
	case <-cam.quit:
		return false
	default:
		return true
	}
}

// Play will start reading images from the OpenCV frame device and
// start queing them up on the frame channel after doing a quick
// sanity check to ensure there are infact an images to be read.
func (cam *Cam) Play() chan *Frame {
	cam.frameQ = make(chan *Frame)
	cam.frames = GetFrameBuffers(cam.BufferSize)
	cam.wg.Add(1)
	go func() {
		defer cam.wg.Done()
		defer close(cam.frameQ)
		for {
			select {
			case <-cam.quit:
				return
			default:
			}
			time.Sleep(5 * time.Millisecond)

			frame := cam.frames.Next()
			cam.cap.Read(frame.Mat)
			if frame.Mat.Empty() {
				continue
			}
			size := frame.Mat.Size()
			if size[0] <= 0 || size[1] <= 0 {
				continue
			}
			// Drain a pending snap request before forwarding the frame.
			select {
			case req := <-cam.snapCh:
				if gocv.IMWrite(req.file, *frame.Mat) {
					req.reply <- nil
				} else {
					req.reply <- fmt.Errorf("IMWrite failed for %s", req.file)
				}
			default:
			}
			select {
			case cam.frameQ <- &frame:
			case <-cam.quit:
				return
			}
		}
	}()

	return cam.frameQ
}

// Snap saves the next valid camera frame to file and returns any error.
// It blocks until a frame is available (up to 5 s) or the camera stops.
func (cam *Cam) Snap(file string) error {
	req := snapReq{file: file, reply: make(chan error, 1)}
	select {
	case cam.snapCh <- req:
	default:
		return fmt.Errorf("snap already in progress")
	}
	select {
	case err := <-req.reply:
		return err
	case <-cam.quit:
		return fmt.Errorf("camera stopped before snap completed")
	case <-time.After(5 * time.Second):
		return fmt.Errorf("snap timed out waiting for next frame")
	}
}

// Close stops reading images from the capture device, waits for the
// goroutine to exit, then releases resources.
func (cam *Cam) Close() error {
	close(cam.quit)
	cam.wg.Wait() // wait for goroutine before freeing Mats
	cam.cap.Close()
	cam.frames.Close()
	return nil
}

// Img will read a single file image and queue it up on the
// FrameQ. The file and subsequently the FrameQ will be closed
type Img struct {
	Filename string
	gocv.IMReadFlag

	frame  *Frame
	frameQ chan *Frame
}

// GetImg opens the image file fname and returns the image
// processing and/or viewing.
func GetImg(fname string) (img *Img, err error) {
	m := gocv.IMRead(fname, gocv.IMReadColor)
	if m.Empty() {
		return nil, fmt.Errorf("ERROR reading %s", fname)
	}

	f := Frame{Mat: &m}
	img = &Img{Filename: fname, frame: &f}
	return img, nil
}

// IsRunning reports true until the single buffered frame has been consumed.
func (i *Img) IsRunning() bool {
	return len(i.frameQ) > 0
}

// Play pre-loads the frame into a buffered channel and closes it immediately,
// so no goroutine is needed and Close() has nothing to race against.
func (i *Img) Play() chan *Frame {
	i.frameQ = make(chan *Frame, 1)
	i.frameQ <- i.frame
	close(i.frameQ)
	return i.frameQ
}

// Close releases the image Mat.
func (i *Img) Close() error {
	i.frame.Mat.Close()
	return nil
}

// Snap writes the loaded image to file.
func (i *Img) Snap(file string) error {
	if !gocv.IMWrite(file, *i.frame.Mat) {
		return fmt.Errorf("IMWrite failed for %s", file)
	}
	return nil
}

// VideoFile does as it's name indicates, reads a video from
// a file and runs it through the pipeline
type VideoFile struct {
	*gocv.VideoCapture

	fname      string
	frames     *FrameBuffers
	frameQ     chan *Frame
	quit       chan struct{}
	snapCh     chan snapReq
	wg         sync.WaitGroup
	bufferSize int
}

// GetVideo reads the file specified by the fname and prepares
// to run it through
func GetVideo(fname string) (vid *VideoFile, err error) {
	vid = &VideoFile{
		bufferSize: 5,
		quit:       make(chan struct{}),
		snapCh:     make(chan snapReq, 1),
	}
	vid.VideoCapture, err = gocv.VideoCaptureFile(fname)
	if err != nil {
		return nil, err
	}

	return vid, nil
}

// GetRTSP opens an RTSP network stream and returns it as a VideoFile.
// The url must begin with rtsp:// or rtsps://.
func GetRTSP(url string) (*VideoFile, error) {
	if !strings.HasPrefix(url, "rtsp://") && !strings.HasPrefix(url, "rtsps://") {
		return nil, fmt.Errorf("GetRTSP: URL must begin with rtsp:// or rtsps://, got %q", url)
	}
	vid := &VideoFile{
		bufferSize: 5,
		quit:       make(chan struct{}),
		snapCh:     make(chan snapReq, 1),
	}
	var err error
	vid.VideoCapture, err = gocv.VideoCaptureFile(url)
	if err != nil {
		return nil, fmt.Errorf("GetRTSP: failed to open %s: %w", url, err)
	}
	return vid, nil
}

func (v *VideoFile) IsRunning() bool {
	select {
	case <-v.quit:
		return false
	default:
		return true
	}
}

func (v *VideoFile) Play() chan *Frame {
	v.frameQ = make(chan *Frame)
	v.frames = GetFrameBuffers(v.bufferSize)
	v.wg.Add(1)
	go func() {
		defer v.wg.Done()
		defer close(v.frameQ)
		for {
			select {
			case <-v.quit:
				return
			default:
			}
			time.Sleep(10 * time.Millisecond)

			frame := v.frames.Next()
			v.VideoCapture.Read(frame.Mat)
			if frame.Mat.Empty() {
				continue
			}
			size := frame.Mat.Size()
			if size[0] <= 0 || size[1] <= 0 {
				continue
			}
			select {
			case req := <-v.snapCh:
				if gocv.IMWrite(req.file, *frame.Mat) {
					req.reply <- nil
				} else {
					req.reply <- fmt.Errorf("IMWrite failed for %s", req.file)
				}
			default:
			}
			select {
			case v.frameQ <- &frame:
			case <-v.quit:
				return
			}
		}
	}()

	return v.frameQ
}

// Snap saves the next valid video frame to file and returns any error.
// It blocks until a frame is available (up to 5 s) or the source stops.
func (v *VideoFile) Snap(file string) error {
	req := snapReq{file: file, reply: make(chan error, 1)}
	select {
	case v.snapCh <- req:
	default:
		return fmt.Errorf("snap already in progress")
	}
	select {
	case err := <-req.reply:
		return err
	case <-v.quit:
		return fmt.Errorf("video source stopped before snap completed")
	case <-time.After(5 * time.Second):
		return fmt.Errorf("snap timed out waiting for next frame")
	}
}

func (v *VideoFile) Close() error {
	close(v.quit)
	v.wg.Wait() // wait for goroutine before freeing Mats
	v.frames.Close()
	return v.VideoCapture.Close()
}
