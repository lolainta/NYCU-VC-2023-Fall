package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/schollz/progressbar/v3"
)

func write_png(filename string, img image.Image) {
	ofile, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	png.Encode(ofile, img)
	ofile.Close()
}

func gray_scale(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	gray := image.NewGray(bounds)
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			tmp := int32(0.299*float64(r)+0.587*float64(g)+0.114*float64(b)) >> 8
			gray.SetGray(i, j, color.Gray{Y: uint8(tmp)})
		}
	}
	return gray
}

func C(i int) float64 {
	if i == 0 {
		return 1 / math.Sqrt(2)
	}
	return 1
}

func dct_2d(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	if w != h {
		log.Println("Image is not square")
		return nil
	}
	c := make([]chan float64, w*h)
	bar := progressbar.Default(int64(w * h))
	for u := 0; u < w; u++ {
		for v := 0; v < h; v++ {
			c[u*h+v] = make(chan float64)
			go func(u, v int, c chan float64) {
				ret := 0.0
				for x := 0; x < w; x++ {
					for y := 0; y < h; y++ {
						f, _, _, _ := img.At(x, y).RGBA()
						r := float64(f) / 256
						ret += float64(r) *
							math.Cos((float64(2*x+1)*float64(u)*math.Pi)/
								(2*float64(w))) *
							math.Cos((float64(2*y+1)*float64(v)*math.Pi)/
								(2*float64(h)))
					}
				}
				c <- ret * C(u) * C(v) * 2 /
					math.Sqrt(float64(w)*float64(h))
			}(u, v, c[u*h+v])
			bar.Add(1)
		}
	}
	ret := image.NewGray(bounds)
	for u := 0; u < w; u++ {
		for v := 0; v < h; v++ {
			cur := 20*math.Log10(math.Abs(<-c[u*h+v])) + 128
			ret.SetGray(u, v, color.Gray{Y: uint8(cur)})
		}
	}
	return ret
}

func dct_1d(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	if w != h {
		log.Println("Image is not square")
		return nil
	}
	bar := progressbar.Default(int64(w + h))
	reg := make([]float64, w*h)
	for u := 0; u < w; u++ {
		cw := make([]chan float64, h)
		for v := 0; v < h; v++ {
			cw[v] = make(chan float64)
			go func(v int, c chan float64) {
				ret := 0.0
				for x := 0; x < w; x++ {
					f, _, _, _ := img.At(x, v).RGBA()
					r := float64(f) / 256
					ret += float64(r) * math.Cos(
						(float64(2*x+1)*float64(u)*math.Pi)/(2*float64(w)))
				}
				c <- ret * C(u) * math.Sqrt(2/float64(w))
			}(v, cw[v])
		}
		bar.Add(1)
		for v := 0; v < h; v++ {
			reg[u*h+v] = <-cw[v]
		}
	}
	for v := 0; v < h; v++ {
		ch := make([]chan float64, w)
		for u := 0; u < w; u++ {
			ch[u] = make(chan float64)
			go func(u int, c chan float64) {
				ret := 0.0
				for y := 0; y < h; y++ {
					ret += reg[u*h+y] * math.Cos(
						(float64(2*y+1)*float64(v)*math.Pi)/(2*float64(h)))
				}
				c <- ret * C(v) * math.Sqrt(2/float64(h))
			}(u, ch[u])
		}
		bar.Add(1)
		for u := 0; u < w; u++ {
			reg[u*w+v] = <-ch[u]
		}
	}
	ret := image.NewGray(bounds)
	for u := 0; u < w; u++ {
		for v := 0; v < h; v++ {
			cur := 20*math.Log10(math.Abs(reg[u*h+v])) + 128
			ret.SetGray(u, v, color.Gray{Y: uint8(cur)})
		}
	}
	return ret
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
	os.Mkdir("output", 0755)
	log.Println("Start processing...")
	gray := gray_scale(src)
	write_png("output/gray.png", gray)
	log.Println("Gray scale done")
	log.Println("Start 1d dct...")
	dct1 := dct_1d(gray)
	write_png("output/dct_1d.png", dct1)
	log.Println("1d dct done")
	log.Println("Start 2d dct...")
	dct2 := dct_2d(gray)
	write_png("output/dct_2d.png", dct2)
	log.Println("2d dct done")
}
