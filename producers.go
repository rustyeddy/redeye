package redeye

type VideoCapture interface{
	Stream(vidQ chan []byte)
}

