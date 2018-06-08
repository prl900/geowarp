package geowarp

import (
	"fmt"
	"github.com/terrascope/gocog"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/scimage"
	"image"
	//"net/url"
	"os"
)

type Raster interface {
	Read() error
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
	defer f.Close()

	info, err := gocog.GeoTIFFInfo(f)
	if err != nil {
		return nil, err
	}

	return &GrayGeoRasterS16{src: path, Proj4: info.Proj4, GeoTrans: info.Geotransform[:]}, nil

}

type GrayGeoRasterS16 struct {
	src string
	*scimage.GrayS16
	Proj4    string
	GeoTrans []float64
	NoData   float64
}

func (gr *GrayGeoRasterS16) getIndex(x, y float64) (int, int) {
	return int(((x - gr.GeoTrans[0]) / gr.GeoTrans[1]) + .5), int(((y - gr.GeoTrans[3]) / gr.GeoTrans[5]) + .5)
}

func (gr *GrayGeoRasterS16) getLocation(i, j int) (float64, float64) {
	return gr.GeoTrans[0] + (gr.GeoTrans[1] * float64(i)), gr.GeoTrans[3] + (gr.GeoTrans[5] * float64(j))
}

func (gr *GrayGeoRasterS16) Read() error {
	f, err := os.Open(gr.src)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := gocog.Decode(f)
	if err != nil {
		return err
	}

	gr.GrayS16 = img.(*scimage.GrayS16)
	return nil

}

func (gr *GrayGeoRasterS16) Warp(dst Raster) error {
	dstS16, ok := dst.(*GrayGeoRasterS16)
	if !ok {
		return fmt.Errorf("dst has to be a GrayGeoRasterS16 type")
	}

	pixPoints := make([]proj4go.Point, dstS16.Bounds().Dx()*dstS16.Bounds().Dy())
	for i := 0; i < dstS16.Bounds().Dx(); i++ {
		for j := 0; j < dstS16.Bounds().Dy(); j++ {
			pixPoints[i+j*dstS16.Bounds().Dx()].X, pixPoints[i+j*dstS16.Bounds().Dx()].Y = dstS16.getLocation(i, j)
		}
	}

	err := proj4go.Transform(gr.Proj4, dstS16.Proj4, pixPoints)
	if err != nil {
		return err
	}

	pixLocations := make([]image.Point, dstS16.Bounds().Dx()*dstS16.Bounds().Dy())
	for i, pt := range pixPoints {
		pixLocations[i].X, pixLocations[i].Y = gr.getIndex(pt.X, pt.Y)
	}

	for i, loc := range pixLocations {
		if image.Rect(loc.X, loc.Y, loc.X, loc.Y).In(gr.Bounds()) {
			dstS16.Pix[2*i] = gr.Pix[loc.X*2+loc.Y*2*gr.Bounds().Dx()]
			dstS16.Pix[2*i+1] = gr.Pix[loc.X*2+loc.Y*2*gr.Bounds().Dx()+1]
		}
	}

	return nil
}
