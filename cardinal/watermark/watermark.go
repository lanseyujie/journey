package watermark

import (
    "errors"
    "image"
    "image/draw"
    "image/gif"
    "image/jpeg"
    "image/png"
    "os"
    "path/filepath"
    "strings"
)

type Position uint

type Watermark struct {
    image    image.Image
    padding  int
    position Position
}

const (
    TopLeft      Position = iota
    TopCenter    Position = iota
    TopRight     Position = iota
    CenterLeft   Position = iota
    CenterCenter Position = iota
    CenterRight  Position = iota
    BottomLeft   Position = iota
    BottomCenter Position = iota
    BottomRight  Position = iota
)

var ErrUnsupportedImageType = errors.New("image: unsupported image type")

func NewWatermark() *Watermark {
    return &Watermark{}
}

// SetMarkImage
func (wm *Watermark) SetMarkImage(file string, pos Position) error {
    f, err := os.Open(file)
    if err != nil {
        return err
    }

    var img image.Image
    switch strings.ToLower(filepath.Ext(file)) {
    case ".jpg", ".jpeg":
        img, err = jpeg.Decode(f)
    case ".png":
        img, err = png.Decode(f)
    case ".gif":
        gifImg, err := gif.DecodeAll(f)
        if err != nil {
            return err
        }
        img = gifImg.Image[0]
    default:
        err = ErrUnsupportedImageType
    }

    if err != nil {
        return err
    }

    wm.image = img
    wm.position = pos

    return nil
}

// getPoint
func (wm *Watermark) getPoint(width, height int) (point image.Point) {
    watermarkWidth := wm.image.Bounds().Dx()
    watermarkHeight := wm.image.Bounds().Dy()
    switch wm.position {
    case TopLeft:
        point = image.Point{
            X: -wm.padding,
            Y: -wm.padding,
        }
    case TopCenter:
        point = image.Point{
            X: -(width - wm.padding - watermarkWidth) / 2,
            Y: -wm.padding,
        }
    case TopRight:
        point = image.Point{
            X: -(width - wm.padding - watermarkWidth),
            Y: -wm.padding,
        }
    case CenterLeft:
        point = image.Point{
            X: -wm.padding,
            Y: -(height - wm.padding - watermarkWidth) / 2,
        }
    case CenterCenter:
        point = image.Point{
            X: -(width - wm.padding - watermarkWidth) / 2,
            Y: -(height - wm.padding - watermarkHeight) / 2,
        }
    case CenterRight:
        point = image.Point{
            X: -(width - wm.padding - watermarkWidth),
            Y: -(height - wm.padding - watermarkHeight) / 2,
        }
    case BottomLeft:
        point = image.Point{
            X: -wm.padding,
            Y: -(height - wm.padding - watermarkHeight),
        }
    case BottomCenter:
        point = image.Point{
            X: -(width - wm.padding - watermarkWidth) / 2,
            Y: -(height - wm.padding - watermarkHeight),
        }
    case BottomRight:
        point = image.Point{
            X: -(width - wm.padding - watermarkWidth),
            Y: -(height - wm.padding - watermarkHeight),
        }
    default:
        // BottomRight
        point = image.Point{
            X: -(width - wm.padding - watermarkWidth),
            Y: -(height - wm.padding - watermarkHeight),
        }
    }

    return point
}

// Do
func (wm *Watermark) Do(imageFile string) error {
    ext := strings.ToLower(filepath.Ext(imageFile))
    f, err := os.OpenFile(imageFile, os.O_RDWR, 0644)
    if err != nil {
        return err
    }

    var (
        img    image.Image
        gifImg *gif.GIF
    )

    switch ext {
    case ".jpg", ".jpeg":
        img, err = jpeg.Decode(f)
    case ".png":
        img, err = png.Decode(f)
    case ".gif":
        gifImg, err = gif.DecodeAll(f)
        if err != nil {
            return err
        }
        img = gifImg.Image[0]
    default:
        err = ErrUnsupportedImageType
    }

    if err != nil {
        return err
    }

    point := wm.getPoint(img.Bounds().Dx(), img.Bounds().Dy())
    if _, err = f.Seek(0, 0); err != nil {
        return err
    }

    switch ext {
    case ".jpg", ".jpeg":
        dstImg := image.NewNRGBA64(img.Bounds())
        draw.Draw(dstImg, dstImg.Bounds(), img, image.Point{}, draw.Src)
        draw.Draw(dstImg, dstImg.Bounds(), wm.image, point, draw.Over)
        return jpeg.Encode(f, dstImg, nil)
    case ".png":
        dstImg := image.NewNRGBA64(img.Bounds())
        draw.Draw(dstImg, dstImg.Bounds(), img, image.Point{}, draw.Src)
        draw.Draw(dstImg, dstImg.Bounds(), wm.image, point, draw.Over)
        return png.Encode(f, dstImg)
    case ".gif":
        for index, img := range gifImg.Image {
            dstImg := image.NewPaletted(img.Bounds(), img.Palette)
            draw.Draw(dstImg, dstImg.Bounds(), img, image.Point{}, draw.Src)
            draw.Draw(dstImg, dstImg.Bounds(), wm.image, point, draw.Over)
            gifImg.Image[index] = dstImg
        }
        return gif.EncodeAll(f, gifImg)
    default:
        return ErrUnsupportedImageType
    }
}
