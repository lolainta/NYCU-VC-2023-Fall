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
	defer ofile.Close()
	err = png.Encode(ofile, img)
	if err != nil {
		log.Println(err)
	}
	ofile.Close()
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
	result := block_matching(src1, src2, 8, 8, 8)
	write_png("result.png", result)
}

func block_matching(src1, src2 image.Image, w, h, diff int) image.Image {
	bounds := src1.Bounds()
	ret := image.NewGray(bounds)
	maxX := bounds.Max.X
	maxY := bounds.Max.Y
	for x := 0; x < maxX; x += w {
		for y := 0; y < maxY; y += h {
			dx, dy := match(src1, src2, x, y, w, diff)
			for i := 0; i < w; i++ {
				for j := 0; j < h; j++ {
					if x+i+dx < 0 || x+i+dx >= maxX || y+j+dy < 0 || y+j+dy >= maxY {
						log.Println("out of range")
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
	maxX := bounds.Max.X
	maxY := bounds.Max.Y
	var min int64 = 1 << 62
	for i := -diff; i <= diff; i++ {
		for j := -diff; j <= diff; j++ {
			if x+i < 0 || x+i+size >= maxX || y+j < 0 || y+j+size >= maxY {
				continue
			}
			sum := mse(src1, src2, x, y, x+i, y+j, size)
			if sum < min {
				min = sum
				retx = i
				rety = j
			}
		}
	}
	return
}

func mse(src1, src2 image.Image, x1, y1, x2, y2, size int) (sum int64) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			r1, g1, b1, _ := src1.At(x1+i, y1+j).RGBA()
			r2, g2, b2, _ := src2.At(x2+i, y2+j).RGBA()
			sum += int64((r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2))
		}
	}
	return
}
