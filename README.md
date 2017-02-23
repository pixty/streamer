# streamer
Streamer has 2 applications (client and server and does the following):

1. client is running in the same sub-network where web camera is. It runs ffmpeg locally to connect the camera and writes rtsp stream to a local file. The client will send the file to a relay server (server) which runs somewhere in Internet using pure HTTP
2. server runs in public (107.170.210.59) and accepts files received from the client

## Compilation 
To compile crossplatform use (for windows as an example)

```
GOOS=windows GOARCH=386 go build -o strmclient.exe streamer.go
```

To compile for linus set GOOS=linux etc...

## Installation on client side

### Installing camera
Get the YI camera and follow standard steps to set up it in the client network. As soon as it starts to work and be able to connected to WIFI.

1. Go to WIFI router and associate the camera MAC with a fixed IP (192.168.2.19) in the example. You will use that one that the router has.
1. Enable RTSP. Do the following steps to enable RTSP stream:

- Turn off the camera (unplug USB)
- Remove micro SD card, if it is installed
- Power on (inset USB) and long press reset button
- Turn off the camera again
- Put the unzipped context from bin/yi-cam-hack firmware file on SD card (with a PC / laptop). The file should be on root folder of the SD card
- Insert micro SD card an power on the camera again
- Wait around 5 min and try to connect with the smartphone app (probably Kalinka-mailinka will sound)
- Check your cam it’s IP address in the router DHCP list

1. You can run VLC and check that you have stream (srtp://192.168.2.19:554/ch0_0.h264)

### Install ffmpeg on local server (Windows)
The local server is a computer where strmclient will be running. You have to have ffmpeg installed there. For Windows machine go to https://ffmpeg.org/download.html#build-windows and download ffmpeg there. Remember the path where you unzip it (C:\tools\ffmpeg-20170221-a5c1c7a-win64-static\bin\ in the example)

Test that ffmpeg works perfect. Open a terminal and run the following command right out there:
```
"C:\tools\ffmpeg-20170221-a5c1c7a-win64-static\bin\ffmpeg.exe -i rtsp://192.168.2.19:554/ch0_0.h264 -t 10 -acodec copy -vcodec copy test.mp4
```
**DO NOT** forget to use correct web-cam IP address in the command above. The command should write 10 seconds of the wideo in local test.mp4 file!

### Running client on local server (Windows)

1. Download strmclient.exe from bin/win-client folder somewhere on the local server.
1. Run the strmclient.exe  by the following command:
``` 
strmclient.exe -command="C:\tools\ffmpeg-20170221-a5c1c7a-win64-static\bin\ffmpeg.exe -i rtsp://192.168.2.19:554/ch0_0.h264 -t <duration> -acodec copy -vcodec copy <outFile>" -duration=300 -target-url=http://107.170.210.59/video-stream
```
where:
- 192.168.2.19 - your camera address, 
- 107.170.210.59 - the relay server in the internet
- C:\tools\ffmpeg-20170221-a5c1c7a-win64-static\bin\ffmpeg.exe - path to ffmpeg.exe

***NOTE!!!*** You can test the command above with duration=10 (10 seconds) and watch for messages in the console. It should be able to write stream into the local files. The command above should be run in terminal and be not touched (run permanently). 

