package geowarp

import (
	"errors"
	"fmt"
	"github.com/terrascope/gocog"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/scimage"
	"image"
	//"net/url"
	"os"
)

type Position struct {
	i, j int
}

type Raster interface {
	IsInBounds(int, int) error
	GetIndex(float64, float64) (int, int)
	GetLocation(int, int) (float64, float64)
	Warp(Raster) error
}

func New(rect image.Rectangle, min, max int16, proj4 string, geot []float64) (Raster, error) {
	return &GrayGeoRasterS16{GrayS16: scimage.NewGrayS16(rect, min, max), Proj4: proj4, GeoTrans: geot}, nil
}

func Open(path string) (Raster, error) {
	//if _, err := url.ParseRequestURI(toTest)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := gocog.GeoTIFFInfo(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	f, err = os.Open(path)
	if err != nil {
		return nil, err
	}

	img, err := gocog.Decode(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	return &GrayGeoRasterS16{GrayS16: img.(*scimage.GrayS16), Proj4: info.Proj4, GeoTrans: info.Geotransform[:]}, nil

}

var OutOfBoundsError = errors.New("out of bounds")

type GrayGeoRasterS16 struct {
	*scimage.GrayS16
	Proj4    string
	GeoTrans []float64
	NoData   float64
}

func (gr *GrayGeoRasterS16) IsInBounds(i, j int) error {
	if i < 0 || j < 0 {
		return OutOfBoundsError
	}
	if i >= gr.Bounds().Dx() || j >= gr.Bounds().Dy() {
		return OutOfBoundsError
	}

	return nil
}

func (gr *GrayGeoRasterS16) GetIndex(x, y float64) (int, int) {
	return int(((x - gr.GeoTrans[0]) / gr.GeoTrans[1]) + .5), int(((y - gr.GeoTrans[3]) / gr.GeoTrans[5]) + .5)
}

func (gr *GrayGeoRasterS16) GetLocation(i, j int) (float64, float64) {
	return gr.GeoTrans[0] + (gr.GeoTrans[1] * float64(i)), gr.GeoTrans[3] + (gr.GeoTrans[5] * float64(j))
}

func (gr *GrayGeoRasterS16) Warp(dst Raster) error {
	dstS16, ok := dst.(*GrayGeoRasterS16)
	if !ok {
		return fmt.Errorf("dst has to be a GrayGeoRasterS16 type")
	}

	pixPoints := make([]proj4go.Point, dstS16.Bounds().Dx()*dstS16.Bounds().Dy())
	for i := 0; i < dstS16.Bounds().Dx(); i++ {
		for j := 0; j < dstS16.Bounds().Dy(); j++ {
			pixPoints[i+j*dstS16.Bounds().Dx()].X, pixPoints[i+j*dstS16.Bounds().Dx()].Y = dstS16.GetLocation(i, j)
		}
	}

	err := proj4go.Transform(gr.Proj4, dstS16.Proj4, pixPoints)
	if err != nil {
		return err
	}

	pixLocations := make([]Position, dstS16.Bounds().Dx()*dstS16.Bounds().Dy())
	for i, pt := range pixPoints {
		pixLocations[i].i, pixLocations[i].j = gr.GetIndex(pt.X, pt.Y)
	}

	for i, loc := range pixLocations {
		if err := gr.IsInBounds(loc.i, loc.j); err == nil {
			dstS16.Pix[2*i] = gr.Pix[loc.i*2+loc.j*2*gr.Bounds().Dx()]
			dstS16.Pix[2*i+1] = gr.Pix[loc.i*2+loc.j*2*gr.Bounds().Dx()+1]
		}
	}

	return nil
}
