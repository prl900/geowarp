package geowarp

import (
	"errors"
	"github.com/prl900/proj4go"
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

type ByteRaster struct {
	Pix          []byte
	Geotransform []float64
	SizeX, SizeY int
	Projection   string
}

func (br *ByteRaster) IsInBounds(i, j int) error {
	if i < 0 || j < 0 {
		return OutOfBoundsError
	}
	if i > br.SizeX || j > br.SizeY {
		return OutOfBoundsError
	}

	return nil
}

func (br *ByteRaster) GetIndex(x, y float64) (int, int) {
	return int(((x - br.Geotransform[0]) / br.Geotransform[1]) + .5), int(((y - br.Geotransform[3]) / br.Geotransform[5]) + .5)
}

func (br *ByteRaster) GetLocation(i, j int) (float64, float64) {
	return br.Geotransform[0] + (br.Geotransform[1] * float64(i)), br.Geotransform[3] + (br.Geotransform[5] * float64(j))
}

func (br *ByteRaster) Warp(dst ByteRaster) error {
	pixPoints := make([]proj4go.Point, dst.SizeX*dst.SizeY)
	for i := 0; i < dst.SizeX; i++ {
		for j := 0; j < dst.SizeY; j++ {
			pixPoints[i+j*dst.SizeX].X, pixPoints[i+j*dst.SizeX].Y = dst.GetLocation(i, j)
		}
	}

	err := proj4go.Forwards(br.Projection, pixPoints)
	if err != nil {
		return err
	}

	pixLocations := make([]Position, dst.SizeX*dst.SizeY)
	for i, pt := range pixPoints {
		pixLocations[i].i, pixLocations[i].j = br.GetIndex(pt.X, pt.Y)
	}

	for i, loc := range pixLocations {
		if err := br.IsInBounds(loc.i, loc.j); err == nil {
			dst.Pix[i] = br.Pix[loc.i+loc.j*br.SizeX]
		}
	}

	return nil
}
