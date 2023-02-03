package redeye

import (
	"fmt"
	"encoding/json"
	"net/http"
)

type Camera struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
	Port int	`json:"port"`
	URI	 string `json:"uri"`
}

var (
	cameras map[string]*Camera
)

func init() {
	cameras = make(map[string]*Camera)
}

func NewCamera(camstr string) *Camera {

	fmt.Println("Camstr: ", camstr)

	var cam Camera
	err := json.Unmarshal([]byte(camstr), &cam)
	if err != nil {
		fmt.Println("ERROR - unmarshal camera json", err)
		return nil
	}

	//cam := &Camera{Name: name, Addr: name, Port: 8080}
	cameras[cam.Name] = &cam
	return &cam;
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

