package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func main() {
	infile, err := os.Open(os.Args[1])
	if err != nil {
		// replace this with real error handling
		// panic(err.String())
		log.Println(err)
	}
	defer infile.Close()

	src, err := png.Decode(infile)
	log.Println(err)
	if err != nil {
		// panic(err.String())
	}

	// log.Println(src.ColorModel().Convert())
	log.Print(src.Bounds())
	log.Print(src.At(0, 0))
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	r_channel := image.NewGray(bounds)
	g_channel := image.NewGray(bounds)
	b_channel := image.NewGray(bounds)
	y_channel := image.NewGray(bounds)
	u_channel := image.NewGray(bounds)
	v_channel := image.NewGray(bounds)
	cb_channel := image.NewGray(bounds)
	cr_channel := image.NewGray(bounds)
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			r, g, b, _ := src.At(i, j).RGBA()
			y := uint32(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
			u := uint32(-0.169*float64(r) - 0.331*float64(g) + 0.5*float64(b))
			v := uint32(0.5*float64(r) - 0.419*float64(g) - 0.081*float64(b))
			cb := uint32(128 - 0.168736*float64(r) - 0.331264*float64(g) + 0.5*float64(b))
			cr := uint32(128 + 0.5*float64(r) - 0.418688*float64(g) - 0.081312*float64(b))
			r_channel.SetGray(i, j, color.Gray{Y: uint8(r >> 8)})
			g_channel.SetGray(i, j, color.Gray{Y: uint8(g >> 8)})
			b_channel.SetGray(i, j, color.Gray{Y: uint8(b >> 8)})
			y_channel.SetGray(i, j, color.Gray{Y: uint8(y >> 8)})
			u_channel.SetGray(i, j, color.Gray{Y: uint8(u >> 8)})
			v_channel.SetGray(i, j, color.Gray{Y: uint8(v >> 8)})
			cb_channel.SetGray(i, j, color.Gray{Y: uint8(cb >> 8)})
			cr_channel.SetGray(i, j, color.Gray{Y: uint8(cr >> 8)})
		}
	}

	os.Mkdir("output", 0777)

	ofile, err := os.Create("output/r_channel.png")
	png.Encode(ofile, r_channel)
	ofile.Close()

	ofile, err = os.Create("output/g_channel.png")
	png.Encode(ofile, g_channel)
	ofile.Close()

	ofile, err = os.Create("output/b_channel.png")
	png.Encode(ofile, b_channel)
	ofile.Close()

	ofile, err = os.Create("output/y_channel.png")
	png.Encode(ofile, y_channel)
	ofile.Close()

	ofile, err = os.Create("output/u_channel.png")
	png.Encode(ofile, u_channel)
	ofile.Close()

	ofile, err = os.Create("output/v_channel.png")
	png.Encode(ofile, v_channel)
	ofile.Close()

	ofile, err = os.Create("output/cb_channel.png")
	png.Encode(ofile, cb_channel)
	ofile.Close()

	ofile, err = os.Create("output/cr_channel.png")
	png.Encode(ofile, cr_channel)
	ofile.Close()

}
