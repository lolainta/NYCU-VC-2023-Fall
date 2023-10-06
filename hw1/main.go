package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func write_png(filename string, img image.Image) {
	ofile, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	png.Encode(ofile, img)
	ofile.Close()
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 2 {
		log.Println("Usage: go run main.go <input.png>")
		return
	}
	infile, err := os.Open(os.Args[1])
	if err != nil {
		log.Println(err)
	}
	defer infile.Close()

	src, err := png.Decode(infile)
	if err != nil {
		log.Println(err)
	}

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
	write_png("output/r_channel.png", r_channel)
	write_png("output/g_channel.png", g_channel)
	write_png("output/b_channel.png", b_channel)
	write_png("output/y_channel.png", y_channel)
	write_png("output/u_channel.png", u_channel)
	write_png("output/v_channel.png", v_channel)
	write_png("output/cb_channel.png", cb_channel)
	write_png("output/cr_channel.png", cr_channel)

}
