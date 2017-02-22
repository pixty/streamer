package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jrivets/gorivets"
	"github.com/jrivets/log4g"
	"golang.org/x/net/context"
)

type streamer struct {
	logger log4g.Logger
	ctx    context.Context

	// sets the duration of file to be written
	durationSec int

	// where to write files
	outputDir string

	// run command
	// ffmpeg -i pixty2.avi -t <duration> -acodec copy -vcodec copy <outFile>
	//
	// <duration> - the record duration is seconds
	// <outFile> - the outputFileName
	cmd string

	outFileExt string
	targetUrl  string
}

func main() {
	logger := log4g.GetLogger("streamer")
	defer log4g.Shutdown()

	s := new(streamer)
	s.logger = logger

	var help bool
	flag.StringVar(&s.outputDir, "output-dir", "./", "Output directory")
	flag.StringVar(&s.cmd, "command", "ffmpeg -i pixty2.avi -t <duration> -acodec copy -vcodec copy <outFile>", "Command")
	flag.StringVar(&s.outFileExt, "file-ext", "mp4", "Output file extension (no dot)")
	flag.StringVar(&s.targetUrl, "target-url", "http://localhost:8080/video-stream", "Where to send the file")
	flag.IntVar(&s.durationSec, "duration", 60, "A file chunk duration")
	flag.BoolVar(&help, "help", false, "Prints the usage")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	s.ctx = ctx
	go s.writeRTSP()

	<-signalChan
	close(signalChan)
	cancel()

	logger.Info("Stopping ...")
}

func (s *streamer) writeRTSP() {
	ch := make(chan string)
	go s.sendFileRoutine(ch)
	for s.ctx.Err() == nil {
		ctx, cancel := context.WithTimeout(s.ctx, time.Duration(s.durationSec+10)*time.Second)

		filename := s.getFileName()
		fullFN := path.Join(s.outputDir, filename)

		cmd, args := s.parseCmd(fullFN)
		s.logger.Info("Executing ", cmd, " with args=", args)
		err := exec.CommandContext(ctx, cmd, args...).Run()
		if err != nil {
			s.logger.Error("Could not execulte, err=", err)
			cancel()
			os.Remove(fullFN)
			time.Sleep(time.Second)
		} else {
			s.logger.Info("Successfully executed, notify sender ...")
			ch <- fullFN
		}
	}
	close(ch)
	s.logger.Info("Context is closed. Exiting writeRTSP")
}

func (s *streamer) getFileName() string {
	now := time.Now()
	return fmt.Sprintf("%4d-%02d-%02d_%02d_%02d_%02d.%s", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), s.outFileExt)
}

func (s *streamer) parseCmd(filename string) (string, []string) {
	dur := strconv.Itoa(s.durationSec)
	cmd := strings.Replace(s.cmd, "<duration>", dur, -1)
	cmd = strings.Replace(cmd, "<outFile>", filename, -1)
	res := strings.Split(cmd, " ")
	return res[0], res[1:]
}

func (s *streamer) sendFileRoutine(ch chan string) {
	for {
		select {
		case fn, ok := <-ch:
			if !ok {
				s.logger.Info("sendFileRoutine: Get channel closed error ")
				return
			}

			err := s.sendFile(fn)
			if err != nil {
				s.logger.Warn("Could not send file=", fn, " to ", s.targetUrl, ", err=", err)
			}

			os.Remove(fn)
		case <-s.ctx.Done():
			s.logger.Info("sendFileRoutine: closed context.")
			return
		}
	}
}

func (s *streamer) sendFile(filename string) error {
	targetUrl := s.targetUrl

	toSec := gorivets.Max(15, s.durationSec/2)
	s.logger.Info("Sending ", filename, " to targetUrl=", targetUrl, ", timeout=", toSec, " sec")

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	_, fn := path.Split(filename)
	fileWriter, err := bodyWriter.CreateFormFile("file", fn)
	if err != nil {
		s.logger.Error("Could not create fileWriter err=", err)
		return err
	}

	fh, err := os.Open(filename)
	if err != nil {
		s.logger.Error("Could not open file=", filename, ", err=", err)
		return err
	}

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		s.logger.Error("Could not copy file, err=", err)
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	timeout := time.Duration(toSec) * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	s.logger.Info("status=", resp.Status, " body=", resp_body)
	return nil
}
