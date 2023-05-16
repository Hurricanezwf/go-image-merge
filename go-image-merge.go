package goimagemerge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Hurricanezwf/go-image-merge/utils"
	"golang.org/x/sync/errgroup"
)

// Specifies how the grid pixel size should be calculated
type gridSizeMode int

const (
	// The size in pixels is fixed for all the grids
	fixedGridSize gridSizeMode = iota
	// The size in pixels is set to the nth image size
	gridSizeFromImage
)

// Grid holds the data for each grid
type Grid struct {
	Image           image.Image
	ImageFilePath   string
	BackgroundColor color.Color
	OffsetX         int
	OffsetY         int
	Grids           []*Grid
}

// MergeImage is the struct that is responsible for merging the given images
type MergeImage struct {
	Grids           []*Grid
	ImageCountDX    int
	ImageCountDY    int
	BaseDir         string
	FixedGridSizeX  int
	FixedGridSizeY  int
	GridSizeMode    gridSizeMode
	GridSizeFromNth int
}

func NewWithRemoteImages(imageURLs []string, imageCountDX, imageCountDY int, opts ...func(*MergeImage)) *MergeImage {
	grids := []*Grid{}
	for _, imageURL := range imageURLs {
		grids = append(grids, &Grid{ImageFilePath: imageURL})
	}
	return New(grids, imageCountDX, imageCountDY, opts...)
}

// New returns a new *MergeImage instance
func New(grids []*Grid, imageCountDX, imageCountDY int, opts ...func(*MergeImage)) *MergeImage {
	mi := &MergeImage{
		Grids:        grids,
		ImageCountDX: imageCountDX,
		ImageCountDY: imageCountDY,
	}

	for _, option := range opts {
		option(mi)
	}

	return mi
}

// OptBaseDir is an functional option to set the BaseDir field
func OptBaseDir(dir string) func(*MergeImage) {
	return func(mi *MergeImage) {
		mi.BaseDir = dir
	}
}

// OptGridSize is an functional option to set the GridSize X & Y
func OptGridSize(sizeX, sizeY int) func(*MergeImage) {
	return func(mi *MergeImage) {
		mi.GridSizeMode = fixedGridSize
		mi.FixedGridSizeX = sizeX
		mi.FixedGridSizeY = sizeY
	}
}

// OptGridSizeFromNthImageSize is an functional option to set the GridSize from the nth image
func OptGridSizeFromNthImageSize(n int) func(*MergeImage) {
	return func(mi *MergeImage) {
		mi.GridSizeMode = gridSizeFromImage
		mi.GridSizeFromNth = n
	}
}

func (m *MergeImage) readGridImage(grid *Grid) (image.Image, error) {
	if grid.Image != nil {
		return grid.Image, nil
	}

	imgPath := grid.ImageFilePath

	if m.BaseDir != "" {
		imgPath = path.Join(m.BaseDir, grid.ImageFilePath)
	}

	return m.ReadImageFile(imgPath)
}

func (m *MergeImage) readGridsImagesFromRemote() ([]image.Image, error) {
	var imagesMutex sync.RWMutex
	var images = make([]image.Image, len(m.Grids))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	for idx, grid := range m.Grids {
		imageIdx := idx
		imageURL := grid.ImageFilePath
		g.Go(func() error {
			rsp, err := http.Get(imageURL)
			if err != nil {
				return fmt.Errorf("failed to download image %s, %w", imageURL, err)
			}
			defer rsp.Body.Close()

			body, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body, %w", err)
			}

			if rsp.StatusCode/100 != 2 {
				return fmt.Errorf("status code %d != 2xx, %s", rsp.StatusCode, string(body))
			}

			if utils.IsPNGImage(body) {
				img, err := png.Decode(bytes.NewReader(body))
				if err != nil {
					return fmt.Errorf("failed to decode png image, %w", err)
				}
				imagesMutex.Lock()
				images[imageIdx] = img
				imagesMutex.Unlock()
			} else if utils.IsJPEGImage(body) {
				img, err := jpeg.Decode(bytes.NewReader(body))
				if err != nil {
					return fmt.Errorf("failed to decode jpeg image, %w", err)
				}
				imagesMutex.Lock()
				images[imageIdx] = img
				imagesMutex.Unlock()
			} else {
				return fmt.Errorf("unsupported format of image %s, expected .png or .jpeg", imageURL)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return images, nil
}

func (m *MergeImage) readGridsImages() ([]image.Image, error) {
	var images []image.Image

	for _, grid := range m.Grids {
		img, err := m.readGridImage(grid)
		if err != nil {
			return nil, err
		}

		images = append(images, img)
	}

	return images, nil
}

func (m *MergeImage) ReadImageFile(path string) (image.Image, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	imgFile, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	var img image.Image
	splittedPath := strings.Split(path, ".")
	ext := splittedPath[len(splittedPath)-1]

	if ext == "jpg" || ext == "jpeg" {
		img, err = jpeg.Decode(imgFile)
	} else {
		img, err = png.Decode(imgFile)
	}

	if err != nil {
		return nil, err
	}

	return img, nil
}

func (m *MergeImage) mergeGrids(images []image.Image) (*image.RGBA, error) {
	var canvas *image.RGBA
	imageBoundX := 0
	imageBoundY := 0

	if m.GridSizeMode == fixedGridSize && m.FixedGridSizeX != 0 && m.FixedGridSizeY != 0 {
		imageBoundX = m.FixedGridSizeX
		imageBoundY = m.FixedGridSizeY
	} else if m.GridSizeMode == gridSizeFromImage {
		imageBoundX = images[m.GridSizeFromNth].Bounds().Dx()
		imageBoundY = images[m.GridSizeFromNth].Bounds().Dy()
	} else {
		imageBoundX = images[0].Bounds().Dx()
		imageBoundY = images[0].Bounds().Dy()
	}

	canvasBoundX := m.ImageCountDX * imageBoundX
	canvasBoundY := m.ImageCountDY * imageBoundY

	canvasMaxPoint := image.Point{canvasBoundX, canvasBoundY}
	canvasRect := image.Rectangle{image.Point{0, 0}, canvasMaxPoint}
	canvas = image.NewRGBA(canvasRect)

	// draw grids one by one
	for i, grid := range m.Grids {
		img := images[i]
		x := i % m.ImageCountDX
		y := i / m.ImageCountDX
		minPoint := image.Point{x * imageBoundX, y * imageBoundY}
		maxPoint := minPoint.Add(image.Point{imageBoundX, imageBoundY})
		nextGridRect := image.Rectangle{minPoint, maxPoint}

		if grid.BackgroundColor != nil {
			draw.Draw(canvas, nextGridRect, &image.Uniform{grid.BackgroundColor}, image.Point{}, draw.Src)
			draw.Draw(canvas, nextGridRect, img, image.Point{}, draw.Over)
		} else {
			draw.Draw(canvas, nextGridRect, img, image.Point{}, draw.Src)
		}

		if grid.Grids == nil {
			continue
		}

		// draw top layer grids
		for _, grid := range grid.Grids {
			img, err := m.readGridImage(grid)
			if err != nil {
				return nil, err
			}

			gridRect := nextGridRect.Bounds().Add(image.Point{grid.OffsetX, grid.OffsetY})
			draw.Draw(canvas, gridRect, img, image.Point{}, draw.Over)
		}
	}

	return canvas, nil
}

// Merge reads the contents of the given file paths, merges them according to given configuration
func (m *MergeImage) Merge() (*image.RGBA, error) {
	useRemote := false
	for _, grid := range m.Grids {
		if strings.HasPrefix(grid.ImageFilePath, "https") || strings.HasPrefix(grid.ImageFilePath, "http") {
			useRemote = true
			break
		}
	}

	var err error
	var images []image.Image

	if useRemote {
		images, err = m.readGridsImagesFromRemote()
	} else {
		images, err = m.readGridsImages()
	}
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, errors.New("There is no image to merge")
	}

	return m.mergeGrids(images)
}
