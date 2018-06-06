package geowarp

import (
	"errors"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/scimage"
)

type Position struct {
	i, j int
}

type Raster interface {
	IsInBounds(int, int) error
	GetIndex(float64, float64, []float64) (int, int)
	GetLocation(int, int, []float64) (float64, float64)
	Warp(Raster) error
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

func (gr *GrayGeoRasterS16) Warp(dst GrayGeoRasterS16) error {
	pixPoints := make([]proj4go.Point, dst.Bounds().Dx()*dst.Bounds().Dy())
	for i := 0; i < dst.Bounds().Dx(); i++ {
		for j := 0; j < dst.Bounds().Dy(); j++ {
			pixPoints[i+j*dst.Bounds().Dx()].X, pixPoints[i+j*dst.Bounds().Dx()].Y = dst.GetLocation(i, j)
		}
	}

	err := proj4go.Transform(gr.Proj4, dst.Proj4, pixPoints)
	if err != nil {
		return err
	}

	pixLocations := make([]Position, dst.Bounds().Dx()*dst.Bounds().Dy())
	for i, pt := range pixPoints {
		pixLocations[i].i, pixLocations[i].j = gr.GetIndex(pt.X, pt.Y)
	}

	for i, loc := range pixLocations {
		if err := gr.IsInBounds(loc.i, loc.j); err == nil {
			dst.Pix[2*i] = gr.Pix[loc.i*2+loc.j*2*gr.Bounds().Dx()]
			dst.Pix[2*i+1] = gr.Pix[loc.i*2+loc.j*2*gr.Bounds().Dx()+1]
		}
	}

	return nil
}
