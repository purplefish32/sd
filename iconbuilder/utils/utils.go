package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"

	"sd/iconbuilder"
)

func CreateIconBuffer(req iconbuilder.IconRequest) ([]byte, error) {
	width, height := 188, 188

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Layer 1: Background color
	bgColor := color.RGBA{R: 255, G: 0, B: 0, A: 255} // Default black

	if len(req.BackgroundColor) == 7 && req.BackgroundColor[0] == '#' {
		var r, g, b uint8
		_, err := fmt.Sscanf(req.BackgroundColor, "#%02x%02x%02x", &r, &g, &b)
		if err == nil {
			bgColor = color.RGBA{R: r, G: g, B: b, A: 255}
		}
	}

	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	// // Layer 2: Overlay the icon
	// iconFile, err := os.Open(req.IconPath)
	// if err != nil {
	// 	return nil, err
	// }
	// defer iconFile.Close()

	// icon, err := png.Decode(iconFile)
	// if err != nil {
	// 	return nil, err
	// }
	// iconBounds := icon.Bounds()
	// offset := image.Pt((width-iconBounds.Dx())/2, (height-iconBounds.Dy())/2)
	// draw.Draw(img, icon.Bounds().Add(offset), icon, image.Point{}, draw.Over)

	// // Layer 3: Render text
	// fontBytes, err := os.ReadFile("font.ttf")
	// if err != nil {
	// 	return nil, err
	// }
	// ft, err := freetype.ParseFont(fontBytes)
	// if err != nil {
	// 	return nil, err
	// }

	// ctx := freetype.NewContext()
	// ctx.SetDPI(72)
	// ctx.SetFont(ft)
	// ctx.SetFontSize(24)
	// ctx.SetClip(img.Bounds())
	// ctx.SetDst(img)
	// ctx.SetSrc(image.Black)
	// pt := freetype.Pt(10, 120)
	// _, err = ctx.DrawString(req.Text, pt)
	// if err != nil {
	// 	return nil, err
	// }

	// Encode to JPEG
	var buf bytes.Buffer
	
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}


