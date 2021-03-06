package redeye

import (
	"encoding/json"
	"net/http"
)

var (
	cameras map[string]*Camera
)

func init() {
	cameras = make(map[string]*Camera)
}

func GetCameras(w http.ResponseWriter, r *http.Request) {
	clist := GetCameraList()
	json.NewEncoder(w).Encode(clist)
}

func GetCameraList() (clist []*Camera) {
	for _, cam := range cameras {
		clist = append(clist, cam)
	}
	return clist
}


func (cam *Camera) Handler(w http.ResponseWriter, req *http.Request) {
	GetCameras(w, req)
}

