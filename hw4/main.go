package main

import (
	"image"
	"image/color"
	"image/png"
	"log"

	"math"
	"os"
)

type block struct {
	x, y int
	data []int
}

type encoded struct {
	x, y int
	data []int
	cnt  []int
}

func write_png(filename string, img image.Image) {
	ofile, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	png.Encode(ofile, img)
	ofile.Close()
}

func zigzag_coefs(n int) ([]int, []int) {
	retx := make([]int, n*n)
	rety := make([]int, n*n)
	cnt := 0
	for i := 0; i < 2*n-1; i++ {
		for j := max(0, i-n+1); j < min(i+1, n); j++ {
			x, y := i-j, j
			if i&1 == 1 {
				x, y = y, x
			}
			retx[cnt] = x
			rety[cnt] = y
			cnt++
		}
	}
	return retx, rety
}

func zigzag_flattern(src [][]int, n int) block {
	ret := block{
		x:    -1,
		y:    -1,
		data: make([]int, n*n),
	}
	cxs, cys := zigzag_coefs(n)
	for i := 0; i < n*n; i++ {
		ret.data[i] = src[cxs[i]][cys[i]]
	}
	return ret
}

func zigzag_deflattern(src block, n int) [][]int {
	ret := make([][]int, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]int, n)
	}
	cxs, cys := zigzag_coefs(n)
	for i := 0; i < n*n; i++ {
		ret[cxs[i]][cys[i]] = src.data[i]
	}
	return ret
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

func crop(img image.Image, x, y, n int) [][]float64 {
	ret := make([][]float64, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			ret[i][j] = float64(img.At(x+i, y+j).(color.Gray).Y)
		}
	}
	return ret
}

func C(i int) float64 {
	if i == 0 {
		return 1 / math.Sqrt(2)
	}
	return 1
}

func dct_2d(img [][]float64, n int) [][]float64 {
	c := make([]chan float64, n*n)
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			c[u*n+v] = make(chan float64)
			go func(u, v int, c chan float64) {
				ret := 0.0
				for x := 0; x < n; x++ {
					for y := 0; y < n; y++ {
						ret += img[x][y] *
							math.Cos((float64(2*x+1)*float64(u)*math.Pi)/
								(2*float64(n))) *
							math.Cos((float64(2*y+1)*float64(v)*math.Pi)/
								(2*float64(n)))
					}
				}
				c <- ret * C(u) * C(v) * 2 / float64(n)
			}(u, v, c[u*n+v])
		}
	}
	ret := make([][]float64, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]float64, n)
	}
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			ret[u][v] = <-c[u*n+v]
		}
	}
	return ret
}

func idct_2d(dct [][]float64, n int) [][]float64 {
	c := make([]chan float64, n*n)
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			c[x*n+y] = make(chan float64)
			go func(x, y int, c chan float64) {
				ret := 0.0
				for u := 0; u < n; u++ {
					for v := 0; v < n; v++ {
						ret += dct[u][v] *
							math.Cos((float64(2*x+1)*float64(u)*math.Pi)/
								(2*float64(n))) *
							math.Cos((float64(2*y+1)*float64(v)*math.Pi)/
								(2*float64(n))) *
							C(u) * C(v)
					}
				}
				c <- ret * 2 / float64(n)
			}(x, y, c[x*n+y])
		}
	}

	ret := make([][]float64, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]float64, n)
	}
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			ret[u][v] = <-c[u*n+v]
		}
	}
	return ret
}

func quantize(dct [][]float64) [][]int {
	w, h := len(dct), len(dct[0])
	ret := make([][]int, w)
	for i := 0; i < w; i++ {
		ret[i] = make([]int, h)
	}
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			if dct[i][j] < math.MinInt8 {
				ret[i][j] = math.MinInt8
			} else if dct[i][j] > math.MaxInt8 {
				ret[i][j] = math.MaxInt8
			} else {
				ret[i][j] = int(int8(dct[i][j]))
			}
		}
	}
	ret[0][0] = int(int16(dct[0][0]))
	return ret
}

func dequantize(dct [][]int) [][]float64 {
	w, h := len(dct), len(dct[0])
	ret := make([][]float64, w)
	for i := 0; i < w; i++ {
		ret[i] = make([]float64, h)
	}
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			ret[i][j] = float64(dct[i][j])
		}
	}
	return ret
}

func rl_encode(blk block) encoded {
	ret := encoded{
		x:    blk.x,
		y:    blk.y,
		data: make([]int, 0),
		cnt:  make([]int, 0),
	}
	cur := blk.data[0]
	cnt := 0
	for i := 0; i < len(blk.data); i++ {
		if blk.data[i] == cur {
			cnt++
		} else {
			ret.data = append(ret.data, cur)
			ret.cnt = append(ret.cnt, cnt)
			cur = blk.data[i]
			cnt = 1
		}
	}
	ret.data = append(ret.data, cur)
	ret.cnt = append(ret.cnt, cnt)
	return ret
}

func rl_decode(data encoded, n int) block {
	ret := block{
		x:    data.x,
		y:    data.y,
		data: make([]int, n*n),
	}
	cur := 0
	for i := 0; i < len(data.data); i++ {
		for j := 0; j < data.cnt[i]; j++ {
			ret.data[cur] = data.data[i]
			cur++
		}
	}
	return ret
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 2 {
		log.Println("Usage: go run main.go <image>")
		return
	}
	img, err := os.Open(os.Args[1])
	if err != nil {
		log.Println(err)
	}
	defer img.Close()

	src, err := png.Decode(img)
	if err != nil {
		log.Println(err)
	}

	gray := gray_scale(src)
	os.Mkdir("output", 0777)
	write_png("output/gray.png", gray)

	bounds := gray.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	n := 8

	data := make([]encoded, 0)
	for i := 0; i < w; i += n {
		for j := 0; j < h; j += n {
			cur := crop(gray, i, j, n)
			dct_coefs := dct_2d(cur, n)
			dct_qcoef := quantize(dct_coefs)
			fimg := zigzag_flattern(dct_qcoef, n)
			fimg.x = i
			fimg.y = j
			data = append(data, rl_encode(fimg))
		}
	}

	restore := image.NewGray(bounds)

	for i := 0; i < len(data); i++ {
		d := data[i]
		fimg := rl_decode(d, n)
		dct_qcoef := zigzag_deflattern(fimg, n)
		dct_coefs := dequantize(dct_qcoef)
		idct_coefs := idct_2d(dct_coefs, n)
		for x := 0; x < n; x++ {
			for y := 0; y < n; y++ {
				restore.SetGray(x+fimg.x, y+fimg.y, color.Gray{Y: uint8(idct_coefs[x][y])})
			}
		}
	}

	write_png("output/restore.png", restore)
}
