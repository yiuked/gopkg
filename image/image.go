package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

func DrawTextToImg(path, out, text string, x int, y int, call func(w io.Writer)) error {
	// 打开图片文件
	file, err := os.Open(path)
	if err != nil {
		log.Printf("failed to open image file: %v\n", err)
		return err
	}
	defer file.Close()

	// 解码图片
	srcImage, format, err := image.Decode(file)
	if err != nil {
		log.Printf("failed to decode image file: %v\n", err)
		return err
	}

	// 创建一个新的图片，并将原始图片绘制到其中
	dstImage := image.NewRGBA(srcImage.Bounds())
	draw.Draw(dstImage, dstImage.Bounds(), srcImage, image.Point{}, draw.Src)

	// 创建一个绘制器
	drawer := freetype.NewContext()
	drawer.SetDst(dstImage)
	drawer.SetSrc(image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: 255})) // 设置文字颜色为白色

	// 设置字体大小？answer: 100.0
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatalf("failed to parse font: %v", err)
		return err
	}

	fontSize := 100.0 // 设置字体大小
	drawer.SetFont(font)
	drawer.SetFontSize(fontSize)
	drawer.SetClip(dstImage.Bounds())

	pt := freetype.Pt(x, y+int(drawer.PointToFixed(fontSize)>>6)) // 文字左下角的位置
	drawer.DrawString(text, pt)

	// 保存结果图片
	outputFile, err := os.Create(out)
	if err != nil {
		log.Printf("failed to create output image: %v\n", err)
		return err
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
		log.Printf("failed to save image: %v\n", err)
		return err
	}

	return nil
}
