package main

import (
	"flag"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/jrivets/gorivets"
	"github.com/jrivets/log4g"
	"gopkg.in/gin-gonic/gin.v1"
)

type server struct {
	logger      log4g.Logger
	storageDir  string
	storMaxSize int64
	port        int
	lock        sync.Mutex
}

func main() {
	defer log4g.Shutdown()

	s := new(server)
	var help, debug bool
	var storMaxSize string
	flag.StringVar(&s.storageDir, "storage-dir", "./", "Where we are going to store data")
	flag.StringVar(&storMaxSize, "max-size", "10Gb", "Where we are going to store data")
	flag.IntVar(&s.port, "port", 8080, "port to listen")
	flag.BoolVar(&help, "help", false, "Prints the usage")
	flag.BoolVar(&debug, "debug", false, "Debug mode")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	s.logger = log4g.GetLogger("streamer")
	if debug {
		log4g.SetLogLevel("streamer", log4g.DEBUG)
	}

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

	s.sweep()
	c.Status(http.StatusCreated)
}

func (s *server) sweep() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.logger.Info("Sweeping...")
	defer s.logger.Info("Sweeping Done")

	files, err := ioutil.ReadDir(s.storageDir)
	if err != nil {
		s.logger.Error("Could not read files from ", s.storageDir, ", cancel sweeping...")
		return
	}

	ss, _ := gorivets.NewSortedSliceByComp(cmpStr, len(files))
	var size int64 = 0
	for _, file := range files {
		_, fn := filepath.Split(file.Name())
		ext := strings.ToLower(filepath.Ext(fn))

		s.logger.Debug("Found file ", fn, " with ext=", ext)

		if strings.HasPrefix(fn, "2017-") && ext == ".mp4" {
			ss.Add(fn)
			size += file.Size()
		}
	}

	s.logger.Info(ss.Len(), " files found, total size is ", gorivets.FormatInt64(size, 1000))

	if size > s.storMaxSize {
		s.logger.Info("Folder size=", gorivets.FormatInt64(size, 1000), " is bigger than maximum value ", gorivets.FormatInt64(s.storMaxSize, 1000), " sweeping files...")
		removed := 0
		for size > s.storMaxSize && ss.Len() > 0 {
			fn := ss.DeleteAt(0).(string)
			fullFn := path.Join(s.storageDir, fn)

			fi, err := os.Stat(fullFn)
			if err != nil {
				s.logger.Error("Could not get stat for ", fullFn, ", err=", err)
				continue
			}

			s.logger.Info("Removing file ", fullFn, " ...")
			err = os.Remove(fullFn)
			if err != nil {
				s.logger.Error("Could not remove file ", fullFn, ", err=", err)
			} else {
				removed++
				size -= fi.Size()
			}
		}
		s.logger.Info(removed, " files were removed, new folder size is ", gorivets.FormatInt64(size, 1000), "(", size, ")")
	}
}

func cmpStr(a, b interface{}) int {
	return strings.Compare(a.(string), b.(string))
}
