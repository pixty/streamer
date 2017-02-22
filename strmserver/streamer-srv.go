package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/jrivets/log4g"
	"gopkg.in/gin-gonic/gin.v1"
)

type server struct {
	logger     log4g.Logger
	storageDir string
	port       int
}

func main() {
	defer log4g.Shutdown()

	s := new(server)
	var help bool
	flag.StringVar(&s.storageDir, "storage-dir", "./", "Where we are going to store data")
	flag.IntVar(&s.port, "port", 8080, "port to listen")
	flag.BoolVar(&help, "help", false, "Prints the usage")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	s.logger = log4g.GetLogger("streamer")
	ge := gin.New()
	ge.GET("/ping", func(c *gin.Context) { s.ping(c) })
	ge.POST("/video-stream", func(c *gin.Context) { s.newImage(c) })
	ge.Run("0.0.0.0:" + strconv.Itoa(s.port))
}

// GET /ping
func (s *server) ping(c *gin.Context) {
	s.logger.Info("GET /ping")
	c.String(http.StatusOK, "pong")
}

// POST /video-stream
func (s *server) newImage(c *gin.Context) {
	s.logger.Info("POST /video-stream")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		s.logger.Error("could not obtain file for upload err=", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	filename := header.Filename
	fullFN := path.Join(s.storageDir, filename)
	s.logger.Info("store data to ", fullFN)

	out, err := os.Create(fullFN)
	if err != nil {
		s.logger.Error("Could not create new file ", fullFN)
		c.Status(http.StatusBadRequest)
		return
	}

	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		s.logger.Error("Could not copy data to file ", fullFN)
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusCreated)
}
