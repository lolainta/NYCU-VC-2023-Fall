package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"time"
)

func write_png(filename string, img image.Image) {
	ofile, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	defer ofile.Close()
	err = png.Encode(ofile, img)
	if err != nil {
		log.Println(err)
	}
	ofile.Close()
}

func psnr(img1, img2 image.Image) float64 {
	bounds := img1.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	if w != h {
		log.Println("Image is not square")
		return 0
	}
	ret := 0.0
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			f1, _, _, _ := img1.At(i, j).RGBA()
			f2, _, _, _ := img2.At(i, j).RGBA()
			r1 := float64(f1) / 256
			r2 := float64(f2) / 256
			ret += math.Pow(r1-r2, 2)
		}
	}
	ret /= float64(w * h)
	ret = 10 * math.Log10(255*255/ret)
	return ret
}

func eval(img1, img2 image.Image, diff int) {
	start := time.Now()
	result := block_matching(img1, img2, 8, 8, diff)
	end := time.Now()
	rate := psnr(img1, result)
	ofile := fmt.Sprintf("result_%d.png", diff)
	write_png(ofile, result)
	log.Printf("%s, time: %v, psnr: %v\n", ofile, end.Sub(start), rate)
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 3 {
		log.Println("Usage: go run main.go <img1.png> <img2.pnng>")
		return
	}
	img1, err := os.Open(os.Args[1])
	if err != nil {
		log.Println(err)
	}
	defer img1.Close()

	src1, err := png.Decode(img1)
	if err != nil {
		log.Println(err)
	}

	img2, err := os.Open(os.Args[2])
	if err != nil {
		log.Println(err)
	}
	defer img2.Close()

	src2, err := png.Decode(img2)
	if err != nil {
		log.Println(err)
	}

	eval(src1, src2, 4)
	eval(src1, src2, 8)
	eval(src1, src2, 16)
	eval(src1, src2, 32)
}

func block_matching(src1, src2 image.Image, w, h, diff int) image.Image {
	bounds := src1.Bounds()
	ret := image.NewGray(bounds)
	maxX, maxY := bounds.Max.X, bounds.Max.Y
	for x := 0; x < maxX; x += w {
		for y := 0; y < maxY; y += h {
			dx, dy := match(src1, src2, x, y, w, diff)
			for i := 0; i < w; i++ {
				for j := 0; j < h; j++ {
					if x+i+dx < 0 || x+i+dx >= maxX || y+j+dy < 0 || y+j+dy >= maxY {
						continue
					}
					tar, _, _, _ := src2.At(x+i+dx, y+j+dy).RGBA()
					ret.SetGray(x+i, y+j, color.Gray{Y: uint8(tar >> 8)})
				}
			}
		}
	}
	return ret
}

func match(src1, src2 image.Image, x, y, size, diff int) (retx, rety int) {
	bounds := src1.Bounds()
	maxX, maxY := bounds.Max.X, bounds.Max.Y
	var min int64 = 1 << 62
	for i := -diff; i <= diff; i++ {
		for j := -diff; j <= diff; j++ {
			if x+i < 0 || x+i+size >= maxX || y+j < 0 || y+j+size >= maxY {
				continue
			}
			cur := mse(src1, src2, x, y, x+i, y+j, size)
			if cur < min {
				min, retx, rety = cur, i, j
			}
		}
	}
	return
}

func mse(src1, src2 image.Image, x1, y1, x2, y2, size int) (sum int64) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			v1, _, _, _ := src1.At(x1+i, y1+j).RGBA()
			v2, _, _, _ := src2.At(x2+i, y2+j).RGBA()
			sum += int64((v1 - v2) * (v1 - v2))
		}
	}
	return
}
