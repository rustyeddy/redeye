package redeye

type VideoStreamer interface{
	Stream(deviceID interface{}, vidQ chan []byte)
}

