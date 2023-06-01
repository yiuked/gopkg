package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

func main() {
	// 打开图片文件
	file, err := os.Open("input.png")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}
	defer file.Close()

	// 解码图片
	srcImage, format, err := image.Decode(file)
	if err != nil {
		log.Fatalf("failed to decode image: %v", err)
	}

	// 创建一个新的图片，并将原始图片绘制到其中
	dstImage := image.NewRGBA(srcImage.Bounds())
	draw.Draw(dstImage, dstImage.Bounds(), srcImage, image.Point{}, draw.Src)

	// 创建一个绘制器
	drawer := freetype.NewContext()
	drawer.SetDst(dstImage)
	drawer.SetSrc(image.NewUniform(color.RGBA{R: 0, G: 0, B: 0, A: 255})) // 设置文字颜色为白色

	// 设置字体大小？answer: 100.0
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatalf("failed to parse font: %v", err)
	}

	fontSize := 90.0 // 设置字体大小
	drawer.SetFont(font)
	drawer.SetFontSize(fontSize)
	drawer.SetClip(dstImage.Bounds())

	text := "100789"
	numbers := strings.Split(text, "")
	for i, number := range numbers {
		pt := freetype.Pt(525+(i*76), 1965+int(drawer.PointToFixed(fontSize)>>6)) // 文字左下角的位置
		drawer.DrawString(number, pt)
	}

	// 保存结果图片
	outputFile, err := os.Create("output.jpg")
	if err != nil {
		log.Fatalf("failed to create output image: %v", err)
	}
	defer outputFile.Close()

	switch strings.ToLower(format) {
	case "jpeg":
		err = jpeg.Encode(outputFile, dstImage, nil)
	case "png":
		err = png.Encode(outputFile, dstImage)
	default:
		log.Fatalf("unsupported image format: %s", format)
	}

	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}

	fmt.Println("Image with text added successfully.")
}
