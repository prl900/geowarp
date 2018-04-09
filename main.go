package main

import (
	//"fmt"
	"./geowarp"
	"image"
	"image/png"
	"os"
	//"time"
	"fmt"
	"github.com/prl900/intersekt"
	"github.com/prl900/proj4go"
)

type Position struct {
	i, j int
}

const (
	sinuProj = "+proj=sinu +lon_0=0 +x_0=0 +y_0=0 +a=6371007.181 +b=6371007.181 +units=m +no_defs "
	tileSize = 240
)

var modisGeot = []float64{-20015109.3539999984204769, 463.3127165279165069, 0.0, 10007554.677005993, 0.0, -463.3127165279165069}

func GetChunkLoop(i, j int) (*intersekt.Loop, error) {

	id := float64(i)
	jd := float64(j)

	tl := proj4go.Point{modisGeot[0] + id*tileSize*modisGeot[1], modisGeot[3] + jd*tileSize*modisGeot[5]}
	tr := proj4go.Point{modisGeot[0] + (id+1)*tileSize*modisGeot[1], modisGeot[3] + jd*tileSize*modisGeot[5]}
	br := proj4go.Point{modisGeot[0] + (id+1)*tileSize*modisGeot[1], modisGeot[3] + (jd+1)*tileSize*modisGeot[5]}
	bl := proj4go.Point{modisGeot[0] + id*tileSize*modisGeot[1], modisGeot[3] + (jd+1)*tileSize*modisGeot[5]}

	return intersekt.NewLoop([]proj4go.Point{tl, tr, br, bl})
}

func GetLoops() ([]intersekt.Loop, error) {
	loops := make([]intersekt.Loop, 360*180)
	for i := 0; i < 360; i++ {
		for j := 0; j < 180; j++ {
			l, err := GetChunkLoop(i, j)
			if err != nil {
				return loops, err
			}
			loops[360*j+i] = *l
		}
	}

	return loops, nil
}

func main() {

	imgPath := "/Users/pablo/Downloads/FC.v302.MCD43A4.A2017001.h30v12.006.png"

	geotModis := []float64{13343406.236, 463.312716528, 0.0, -3335851.559, 0.0, -463.312716528}
	projModis := "+proj=sinu +lon_0=0 +x_0=0 +y_0=0 +a=6371007.181 +b=6371007.181 +units=m +no_defs "

	geotCanvas := []float64{148.7109375, 0.01171875, 0, -33.578014746143985, 0, -0.01171875}

	imgFile, err := os.Open(imgPath)
	if err != nil {
		panic(err)
	}
	defer imgFile.Close()

	modis, err := png.Decode(imgFile)
	if err != nil {
		panic(err)
	}

	imgRGBA := modis.(*image.RGBA)
	imgGray := image.NewGray(imgRGBA.Rect)
	for i := 0; i < imgGray.Rect.Dx()*imgGray.Rect.Dy(); i++ {
		imgGray.Pix[i] = imgRGBA.Pix[i*4]
	}

	modisRaster := geowarp.ByteRaster{imgGray.Pix, geotModis, imgGray.Rect.Dx(),
		imgGray.Rect.Dy(), projModis}
	canvasRaster := geowarp.ByteRaster{make([]byte, 256*256), geotCanvas,
		256, 256, ""}

	modisRaster.Warp(canvasRaster)

	imgCanvas := image.NewGray(image.Rect(0, 0, 256, 256))
	imgCanvas.Pix = canvasRaster.Pix

	outFile, err := os.Create("/Users/pablo/Downloads/outGray.png")
	if err != nil {
		panic(err)
	}
	defer imgFile.Close()

	err = png.Encode(outFile, imgCanvas)
	if err != nil {
		panic(err)
	}

	loops, _ := GetLoops()

	tl := intersekt.Point{0, 40}
	tr := intersekt.Point{5, 40}
	br := intersekt.Point{5, 30}
	bl := intersekt.Point{0, 30}

	a, _ := intersekt.NewLoop([]intersekt.Point{tl, tr, br, bl})
	fmt.Println(a)
	sa := a.Segmentize(2)

	b := a.Transform(sinuProj)
	sb := sa.Transform(sinuProj)

	fmt.Println(b)
	fmt.Println(sb)

	i := 0
	for _, l := range loops {
		if b.Intersects(&l) {
			i++
		}
	}
	fmt.Println(i)
}
