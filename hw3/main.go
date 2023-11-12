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

func write_png(filename string, img *image.Image) {
	ofile, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	defer ofile.Close()
	err = png.Encode(ofile, *img)
	if err != nil {
		log.Println(err)
	}
	ofile.Close()
}

func psnr(img1, img2 *image.Image) float64 {
	bounds := (*img1).Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	if w != h {
		log.Println("Image is not square")
		return 0
	}
	ret := 0.0
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			f1, _, _, _ := (*img1).At(i, j).RGBA()
			f2, _, _, _ := (*img2).At(i, j).RGBA()
			r1 := float64(f1) / 256
			r2 := float64(f2) / 256
			ret += math.Pow(r1-r2, 2)
		}
	}
	ret /= float64(w * h)
	ret = 10 * math.Log10(255*255/ret)
	return ret
}

func eval(img1, img2 *image.Image, fname string, diff int, method func(*image.Image, *image.Image, int, int, int) image.Image) {
	start := time.Now()
	result := method(img1, img2, 8, 8, diff)
	end := time.Now()
	rate := psnr(img1, &result)
	ofile := fmt.Sprintf("output/%s.png", fname)
	write_png(ofile, &result)
	log.Printf("%24s, time: %v, psnr: %v\n", ofile, end.Sub(start), rate)
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 3 {
		log.Println("Usage: go run main.go <img1.png> <img2.png>")
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
	os.Mkdir("output", 0777)
	eval(&src1, &src2, "block_4", 4, block_matching)
	eval(&src1, &src2, "block_8", 8, block_matching)
	eval(&src1, &src2, "block_16", 16, block_matching)
	eval(&src1, &src2, "block_32", 32, block_matching)
	eval(&src1, &src2, "three_step", 4, three_step)
	eval(&src1, &src2, "full_search", 320, block_matching)
}

func block_matching(src1, src2 *image.Image, w, h, diff int) image.Image {
	bounds := (*src1).Bounds()
	ret := image.NewGray(bounds)
	maxX, maxY := bounds.Max.X, bounds.Max.Y
	for x := 0; x < maxX; x += w {
		for y := 0; y < maxY; y += h {
			dx, dy := match(src1, src2, x, y, w, diff, 1)
			for i := 0; i < w; i++ {
				for j := 0; j < h; j++ {
					tar, _, _, _ := (*src2).At(x+i+dx, y+j+dy).RGBA()
					ret.SetGray(x+i, y+j, color.Gray{Y: uint8(tar >> 8)})
				}
			}
		}
	}
	return ret
}

func three_step(src1, src2 *image.Image, w, h, diff int) image.Image {
	bounds := (*src1).Bounds()
	ret := image.NewGray(bounds)
	maxX, maxY := bounds.Max.X, bounds.Max.Y
	for x := 0; x < maxX; x += w {
		for y := 0; y < maxY; y += h {
			dx, dy := match(src1, src2, x, y, w, diff, diff)
			dx1, dy1 := match(src1, src2, x+dx, y+dy, w, diff/2, diff/2)
			dx2, dy2 := match(src1, src2, x+dx+dx1, y+dy+dy1, w, diff/4, diff/4)
			for i := 0; i < w; i++ {
				for j := 0; j < h; j++ {
					tar, _, _, _ := (*src2).At(x+i+dx+dx1+dx2, y+j+dy+dy1+dy2).RGBA()
					ret.SetGray(x+i, y+j, color.Gray{Y: uint8(tar >> 8)})
				}
			}
		}
	}
	return ret
}

func match(src1, src2 *image.Image, x, y, size, bd, step int) (retx, rety int) {
	bounds := (*src1).Bounds()
	maxX, maxY := bounds.Max.X, bounds.Max.Y
	var min int64 = 1 << 62
	for i := -bd; i <= bd; i += step {
		for j := -bd; j <= bd; j += step {
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

func mse(src1, src2 *image.Image, x1, y1, x2, y2, size int) (sum int64) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			v1, _, _, _ := (*src1).At(x1+i, y1+j).RGBA()
			v2, _, _, _ := (*src2).At(x2+i, y2+j).RGBA()
			sum += int64((v1 - v2) * (v1 - v2))
		}
	}
	return
}
