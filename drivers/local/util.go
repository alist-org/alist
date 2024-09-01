package local

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func isSymlinkDir(f fs.FileInfo, path string) bool {
	if f.Mode()&os.ModeSymlink == os.ModeSymlink {
		dst, err := os.Readlink(filepath.Join(path, f.Name()))
		if err != nil {
			return false
		}
		if !filepath.IsAbs(dst) {
			dst = filepath.Join(path, dst)
		}
		stat, err := os.Stat(dst)
		if err != nil {
			return false
		}
		return stat.IsDir()
	}
	return false
}

func GetSnapshot(videoPath string, frameNum int) (imgData *bytes.Buffer, err error) {
	srcBuf := bytes.NewBuffer(nil)
	stream := ffmpeg.Input(videoPath).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		GlobalArgs("-loglevel", "error").Silent(true).
		WithOutput(srcBuf, os.Stdout)
	if err = stream.Run(); err != nil {
		return nil, err
	}
	return srcBuf, nil
}

func readDir(dirname string) ([]fs.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, nil
}

type thumbRequest struct {
	file      model.Obj
	result    chan *thumbResult
	fullPath  string
	thumbName string
}

type thumbResult struct {
	buf  *bytes.Buffer
	path *string
	err  error
}

type thumbGenerator struct {
	queue       chan *thumbRequest
	concurrency int
	cacheFolder string
}

func newThumbGenerator(concurrency int, cacheFolder string) *thumbGenerator {
	g := &thumbGenerator{
		queue:       make(chan *thumbRequest),
		concurrency: concurrency,
		cacheFolder: cacheFolder,
	}
	for i := 0; i < concurrency; i++ {
		go g.worker()
	}
	return g
}

func (g *thumbGenerator) worker() {
	for req := range g.queue {
		result := g.generateThumb(req.file, req.fullPath, req.thumbName)
		req.result <- result
	}
}

func (g *thumbGenerator) GenerateThumb(file model.Obj) (*bytes.Buffer, *string, error) {
	fullPath := file.GetPath()
	thumbPrefix := "alist_thumb_"
	thumbName := thumbPrefix + utils.GetMD5EncodeStr(fullPath) + ".png"

	if g.cacheFolder != "" {
		// skip if the file is a thumbnail
		if strings.HasPrefix(file.GetName(), thumbPrefix) {
			return nil, &fullPath, nil
		}
		thumbPath := filepath.Join(g.cacheFolder, thumbName)
		if utils.Exists(thumbPath) {
			return nil, &thumbPath, nil
		}
	}

	resultChan := make(chan *thumbResult)
	g.queue <- &thumbRequest{
		file:      file,
		result:    resultChan,
		fullPath:  fullPath,
		thumbName: thumbName,
	}
	result := <-resultChan
	return result.buf, result.path, result.err
}

func (g *thumbGenerator) generateThumb(file model.Obj, fullPath, thumbName string) *thumbResult {
	var srcBuf *bytes.Buffer
	var err error
	if utils.GetFileType(file.GetName()) == conf.VIDEO {
		videoBuf, err := GetSnapshot(fullPath, 10)
		if err != nil {
			return &thumbResult{err: err}
		}
		srcBuf = videoBuf
	} else {
		imgData, err := os.ReadFile(fullPath)
		if err != nil {
			return &thumbResult{err: err}
		}
		imgBuf := bytes.NewBuffer(imgData)
		srcBuf = imgBuf
	}

	image, err := imaging.Decode(srcBuf, imaging.AutoOrientation(true))
	if err != nil {
		return &thumbResult{err: err}
	}
	thumbImg := imaging.Resize(image, 144, 0, imaging.Lanczos)
	var buf bytes.Buffer
	err = imaging.Encode(&buf, thumbImg, imaging.PNG)
	if err != nil {
		return &thumbResult{err: err}
	}

	if g.cacheFolder != "" {
		err = os.WriteFile(filepath.Join(g.cacheFolder, thumbName), buf.Bytes(), 0666)
		if err != nil {
			return &thumbResult{err: errors.Wrap(err, "failed to write thumbnail to cache")}
		}
		// return &thumbResult{path: &thumbName}
		return &thumbResult{buf: &buf}
	}
	return &thumbResult{buf: &buf}
}

func (d *Local) getThumb(file model.Obj) (*bytes.Buffer, *string, error) {
	return d.thumbGenerator.GenerateThumb(file)
}
