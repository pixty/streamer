# streamer
## Running client on windows
1. Install ffmpeg. Go to https://ffmpeg.org/download.html and download the ffmpeg. Unzip it somewhere to C:\tools\ for example.
1. Download streamer.exe from bin folder somewhere.
1. Run the streamer.exe by the command:
``` 
streamer.exe -command="C:\tools\ffmpeg-20170221-a5c1c7a-win64-static\bin\ffmpeg.exe -i rtsp://192.168.2.19:554/ch0_0.h264 -t <duration> -acodec copy -vcodec copy <outFile>" -duration=60 -target-url=http://192.168.2.5:8081/video-stream
```
where 192.168.2.19 - your camera address, 192.168.2.5:8081 - the file server where all files will be stored to.
