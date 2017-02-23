package main

import (
	"flag"
	"io"
	//"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	//"path/filepath"
	"strconv"
	"strings"

	"github.com/jrivets/gorivets"
	"github.com/jrivets/log4g"
	"gopkg.in/gin-gonic/gin.v1"
)

type server struct {
	logger      log4g.Logger
	storageDir  string
	storMaxSize int64
	port        int
}

func main() {
	defer log4g.Shutdown()

	s := new(server)
	var help bool
	var storMaxSize string
	flag.StringVar(&s.storageDir, "storage-dir", "./", "Where we are going to store data")
	flag.StringVar(&storMaxSize, "max-size", "10Gb", "Where we are going to store data")
	flag.IntVar(&s.port, "port", 8080, "port to listen")
	flag.BoolVar(&help, "help", false, "Prints the usage")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	s.logger = log4g.GetLogger("streamer")

	maxSize, err := gorivets.ParseInt64(storMaxSize, 10000000, math.MaxInt64, 10000000)
	if err != nil {
		s.logger.Error("Could not parse -max-size=", storMaxSize, ", err=", err)
		return
	}
	s.storMaxSize = maxSize
	s.logger.Info("Starting for storage-dir=", s.storageDir, ", -max-size=", storMaxSize,
		"(", maxSize, ")")

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

func (s *server) sweep() {
	//	s.logger.Info("Sweeping...")
	//	files, err := ioutil.ReadDir(s.storageDir)
	//	if err != nil {
	//		s.logger.Error("Could not read files from ", s.storageDir, ", cancel sweeping...")
	//		return
	//	}

	//	ss := gorivets.NewSortedSliceByComp(cmpStr, len(files))
	//	for file := range files {
	//		_, fn := filepath.Split(file.Name())
	//		ext := strings.ToLower(filepath.Ext(fn))

	//		if
	//	}
}

func cmpStr(a, b interface{}) int {
	return strings.Compare(a.(string), b.(string))
}
