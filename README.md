# GoNeoCam
GoNeoCam is a web server that exposes a REST api for the Neoden camera DLL.

## Building

```sh
set GOARCH=386
go build server.go
```

## API

Some function arguments arguments are still in the TBD category. .

```sh
/cameras/{cameraId}/imgSetWidthHeight?width={width}&height={height}
/cameras/{cameraId}/imgReadAsy?width={width}&height={height}&timeout){timeout}
/cameras/{cameraId}/imgLed?mode={mode}
/cameras/{cameraId}/imgReset
/cameras/{cameraId}/imgSetExposure?exposure={exposure}
/cameras/{cameraId}/imgSetGain?gain={gain}
/cameras/{cameraId}/imgSetLt?a2={a2}&a3={a3}
```

## Testing

* The Neoden camera DLL has to be located in the folder that will be searched during DLL loading. It wil reside quite happily in the same folder as the GoNeoCam server

* Keep in mind that some of these functions are still untested. Experiment at your own risk.
* img_init is executed a few seconds after the server has started.
* the native version of img_readAsy returns a 8 bit buffer of raw image bytes. The REST API version (imgReadAsy) returns a PNG greyscale image. Example:

![Alt text](/imgReadAsy.png?raw=true "Example output from imgReadAsy (1024x1024)")





