/*
Copyright [2019] [Telenor Digital AS]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/gorilla/mux"
)

var mod = syscall.NewLazyDLL("neodencamera.dll")
var cameras int

// initializeCameras snoozes for 5 seconds in order to allow the Neoden DLL to get its shit together before cameras can be detected via imgInit()
func initializeCameras() int {
	var delay time.Duration = 5
	log.Printf("Starting up in %d seconds...", delay)
	time.Sleep(delay * time.Second)
	cameras := imgInit()
	if cameras < 2 {
		log.Fatal(fmt.Printf("Bummer, detected %d cameras...", cameras))
	}
	log.Printf("Downward facing camera: ID 1")
	log.Printf("Upward facing camera: ID 5")
	log.Printf("ImgInit detected %d cameras", cameras)

	return cameras
}

// imgInit is a wrapper for img_init in the Neoden camera DLL.
// Returns the number of cameras detected
func imgInit() int {
	log.Printf("imgInit()")
	var f = mod.NewProc("img_init")
	ret, _, _ := f.Call()
	return int(ret)
}

// imgSetWidthHeight is a wrapper for img_set_whm in the Neoden camera DLL.
// Arguments:
// 		camera	: cameraID (0: first camera)
// 		width 	: image width
// 		height	: image height
// Returns:
//		camera ID
func imgSetWidthHeight(camera int, width int, height int) int {
	log.Printf("imgSetWidthHeight - camera:%d width:%d height:%d", camera, width, height)
	var f = mod.NewProc("img_set_wh")
	ret, _, _ := f.Call(uintptr(camera), uintptr(width), uintptr(height))
	return int(ret) // retval is cameraID
}

// imgReadAsy is a wrapper for img_readAsy in the Neoden camera DLL.
// Arguments:
// 		camera	: cameraID (0: first camera)
// 		width	: image width
// 		height	: image height
//		timeout	: timeout in milliseconds
// Returns:
//		camera ID
func imgReadAsy(camera int, width int, height int, timeout int) (int, []byte) {
	log.Printf("imgReadAsy camera:%d width:%d height:%d timeout:%d", camera, width, height, timeout)

	var buffer = make([]byte, width*height, width*height)
	var f = mod.NewProc("img_readAsy")
	ret, _, _ := f.Call(uintptr(camera), uintptr(unsafe.Pointer(&buffer[0])), uintptr(width*height), uintptr(timeout))

	img := image.NewGray(image.Rect(0, 0, width, height))
	img.Pix = buffer
	var pngBuffer = new(bytes.Buffer)

	errEncode := png.Encode(pngBuffer, img)
	if nil != errEncode {
		log.Printf("Error encoding PNG : %v\n", errEncode)
	}

	return int(ret), pngBuffer.Bytes()
}

// imgLed is a wrapper for img_led in the Neoden camera DLL
// Arguments:
// 		camera	: cameraID (0: first camera)
// 		mode 	: TBD
// Returns:
//		camera ID
func imgLed(camera int, mode int) int {
	log.Printf("imgLed camera:%d mode:%d", camera, mode)
	var f = mod.NewProc("img_led")
	ret, _, _ := f.Call(uintptr(camera), uintptr(mode))
	return int(ret) // retval is cameraID
}

// imgReset is a wrapper for img_reset in the Neoden camera DLL
// Arguments:
// 		camera	: cameraID (0: first camera)
// Returns:
//		camera ID
func imgReset(camera int) int {
	log.Printf("imgReset camera:%d", camera)
	var f = mod.NewProc("img_reset")
	ret, _, _ := f.Call(uintptr(camera))
	return int(ret) // retval is cameraID
}

// imgSetExposure is a wrapper for img_set_exp in the Neoden camera DLL
// Arguments:
// 		camera		: cameraID (0: first camera)
// 		exposure	: exposure (TBD valid range)
// Returns:
//		camera ID
func imgSetExposure(camera int, exposure int) int {
	log.Printf("imgSetExposure camera:%d exposure:%d", camera, exposure)
	var f = mod.NewProc("img_set_exp")
	ret, _, _ := f.Call(uintptr(camera), uintptr(exposure))
	return int(ret) // retval is cameraID
}

// imgSetGain is a wrapper for img_set_gain in the Neoden camera DLL
// Arguments:
// 		camera		: cameraID (0: first camera)
// 		gain		: gain (TBD valid range)
// Returns:
//		camera ID
func imgSetGain(camera int, gain int) int {
	log.Printf("imgSetGain camera:%d gain:%d", camera, gain)
	var f = mod.NewProc("img_set_gain")
	ret, _, _ := f.Call(uintptr(camera), uintptr(gain))
	return int(ret) // retval is cameraID
}

// imgSetLt is a wrapper for img_set_gain in the Neoden camera DLL
// Arguments:
// 		camera		: cameraID (0: first camera)
// 		a2			: TBD
// 		a3			: TBD
// Returns:
//		camera ID
func imgSetLt(camera int, a2 int, a3 int) int {
	log.Printf("imgReset camera:%d a2:%d a3:%d", camera, a2, a3)
	var f = mod.NewProc("img_set_lt")
	ret, _, _ := f.Call(uintptr(camera), uintptr(a2), uintptr(a3))
	return int(ret) // retval is cameraID
}

// isValidCamera checks for specified down camera (1) or up camera (5)
func isValidCamera(cameraID int) bool {
	if (cameraID == 1) || (cameraID == 5) {
		return true
	}
	return false
}

// imgReadAsyHandler is a handler for /cameras/{cameraId}/imgReadAsy
// Expected query params:
//		width: Image width
//		height: Image height
//		timeout: Timeout in milliseconds
// Returns:
//		Camera image as greyscale PNG
func imgReadAsyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cameraID, _ := strconv.Atoi(vars["cameraId"])
	if !isValidCamera(cameraID) {
	 	http.Error(w, "Invalid camera", 500)
	 	return
	 }

	width, _ := strconv.Atoi(vars["width"])
	height, _ := strconv.Atoi(vars["height"])
	timeout, _ := strconv.Atoi(vars["timeout"])

	_, png := imgReadAsy(cameraID, width, height, timeout)

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Write(png)
}

// imgSetWidthHeightHandler is a handler for /cameras/{cameraId}/imgSetWidthHeight
// Expected query params:
//		width: Image width
//		height: Image height
func imgSetWidthHeightHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cameraID, _ := strconv.Atoi(vars["cameraId"])
	if !isValidCamera(cameraID) {
	 	http.Error(w, "Invalid camera", 500)
	 	return
	 }

	width, _ := strconv.Atoi(vars["width"])
	height, _ := strconv.Atoi(vars["height"])

	imgSetWidthHeight(cameraID, width, height)
}

// imgLedHandler is a handler for /cameras/{cameraId}/imgLed
// Expected query params:
//		mode : TBD
func imgLedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cameraID, _ := strconv.Atoi(vars["cameraId"])
	if !isValidCamera(cameraID) {
	 	http.Error(w, "Invalid camera", 500)
	 	return
	 }

	mode, _ := strconv.Atoi(vars["mode"])

	imgLed(cameraID, mode)
}

// imgResetHandler is a handler for /cameras/{cameraId}/imgReset
func imgResetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cameraID, _ := strconv.Atoi(vars["cameraId"])
	if !isValidCamera(cameraID) {
		http.Error(w, "Invalid camera", 500)
		return
	}

	imgReset(cameraID)
}

// imgSetExposureHandler is a handler for /cameras/{cameraId}/imgSetExposure
func imgSetExposureHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cameraID, _ := strconv.Atoi(vars["cameraId"])
	if !isValidCamera(cameraID) {
		http.Error(w, "Invalid camera", 500)
		return
	}

	exposure, _ := strconv.Atoi(vars["exposure"])
	imgSetExposure(cameraID, exposure)
}

// imgSetGainHandler is a handler for /cameras/{cameraId}/imgSetGain
func imgSetGainHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cameraID, _ := strconv.Atoi(vars["cameraId"])
	if !isValidCamera(cameraID) {
		http.Error(w, "Invalid camera", 500)
		return
	}

	gain, _ := strconv.Atoi(vars["gain"])
	imgSetGain(cameraID, gain)
}

// imgSetLtHandler is a handler for /cameras/{cameraId}/imgSetLt
func imgSetLtHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cameraID, _ := strconv.Atoi(vars["cameraId"])
	if !isValidCamera(cameraID) {
		http.Error(w, "Invalid camera", 500)
		return
	}

	a2, _ := strconv.Atoi(vars["a2"])
	a3, _ := strconv.Atoi(vars["a3"])
	imgSetLt(cameraID, a2, a3)
}

func main() {
	cameras = initializeCameras()

	r := mux.NewRouter()
	r.HandleFunc("/cameras/{cameraId}/imgSetWidthHeight", imgSetWidthHeightHandler).Queries("width", "{width}").Queries("height", "{height}")
	r.HandleFunc("/cameras/{cameraId}/imgReadAsy", imgReadAsyHandler).Queries("width", "{width}").Queries("height", "{height}").Queries("timeout", "{timeout}")
	r.HandleFunc("/cameras/{cameraId}/imgLed", imgLedHandler).Queries("mode", "{mode}")
	r.HandleFunc("/cameras/{cameraId}/imgReset", imgResetHandler)
	r.HandleFunc("/cameras/{cameraId}/imgSetExposure", imgSetExposureHandler).Queries("exposure", "{exposure}")
	r.HandleFunc("/cameras/{cameraId}/imgSetGain", imgSetGainHandler).Queries("gain", "{gain}")
	r.HandleFunc("/cameras/{cameraId}/imgSetLt", imgSetLtHandler).Queries("a2", "{a2}").Queries("a3", "{a3}")

	port := 8080
	log.Printf("Listening on : %d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}