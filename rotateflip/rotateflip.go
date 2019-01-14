package rotateflip

import (
	"image"
	"image/color"
)

// RotateFlipType specifies a clockwise rotation and flip to apply to an image.
type RotateFlipType int

const (
	RotateNoneFlipNone RotateFlipType = iota
	Rotate90FlipNone
	Rotate180FlipNone
	Rotate270FlipNone

	RotateNoneFlipX
	Rotate90FlipX
	Rotate180FlipX
	Rotate270FlipX

	RotateNoneFlipY  = Rotate180FlipX
	Rotate90FlipY    = Rotate270FlipX
	Rotate180FlipY   = RotateNoneFlipX
	Rotate270FlipY   = Rotate90FlipX
	RotateNoneFlipXY = Rotate180FlipNone
	Rotate90FlipXY   = Rotate270FlipNone
	Rotate180FlipXY  = RotateNoneFlipNone
	Rotate270FlipXY  = Rotate90FlipNone
)

// Orientation is an image orientation as specified by EXIF 2.2 and TIFF 6.0
type Orientation int

const (
	TopLeft Orientation = iota + 1
	TopRight
	BottomRight
	BottomLeft
	LeftTop
	RightTop
	RightBottom
	LeftBottom
)

// Type gets the RotateFlipType that restores an image with this Orientation to TopLeft Orientation
func (or Orientation) Type() RotateFlipType {
	switch or {
	default:
		return RotateNoneFlipNone
	case TopRight:
		return RotateNoneFlipX
	case BottomRight:
		return RotateNoneFlipXY
	case BottomLeft:
		return RotateNoneFlipY
	case LeftTop:
		return Rotate90FlipX
	case RightTop:
		return Rotate90FlipNone
	case RightBottom:
		return Rotate90FlipY
	case LeftBottom:
		return Rotate90FlipXY
	}
}

// Image applies an Operation to an image
func Image(src image.Image, t RotateFlipType) image.Image {
	t &= 7 // sanitize

	if t == 0 {
		return src // nop
	}

	rotate := t&1 != 0
	bounds := rotateBounds(src.Bounds(), rotate)

	// fast path, eager
	switch src := src.(type) {
	case *image.Alpha:
		dst := image.NewAlpha(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
		return dst

	case *image.Alpha16:
		dst := image.NewAlpha16(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 2)
		return dst

	case *image.CMYK:
		dst := image.NewCMYK(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 4)
		return dst

	case *image.Gray:
		dst := image.NewGray(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
		return dst

	case *image.Gray16:
		dst := image.NewGray16(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 2)
		return dst

	case *image.NRGBA:
		dst := image.NewNRGBA(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 4)
		return dst

	case *image.NRGBA64:
		dst := image.NewNRGBA64(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 8)
		return dst

	case *image.RGBA:
		dst := image.NewRGBA(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 4)
		return dst

	case *image.RGBA64:
		dst := image.NewRGBA64(bounds)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 8)
		return dst

	case *image.Paletted:
		dst := image.NewPaletted(bounds, src.Palette)
		rotateFlip(dst.Pix, dst.Stride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Pix, src.Stride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
		return dst

	case *image.YCbCr:
		if sr, ok := rotateYCbCrSubsampleRatio(src.SubsampleRatio, rotate); ok {
			dst := image.NewYCbCr(bounds, sr)
			rotateFlip(dst.Y, dst.YStride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Y, src.YStride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
			rotateFlip(dst.Cb, dst.CStride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Cb, src.CStride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
			rotateFlip(dst.Cr, dst.CStride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Cr, src.CStride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
			return dst
		}

	case *image.NYCbCrA:
		if sr, ok := rotateYCbCrSubsampleRatio(src.SubsampleRatio, rotate); ok {
			dst := image.NewNYCbCrA(bounds, sr)
			rotateFlip(dst.Y, dst.YStride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Y, src.YStride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
			rotateFlip(dst.A, dst.AStride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.A, src.AStride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
			rotateFlip(dst.Cb, dst.CStride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Cb, src.CStride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
			rotateFlip(dst.Cr, dst.CStride, dst.Bounds().Dx(), dst.Bounds().Dy(), src.Cr, src.CStride, src.Bounds().Dx(), src.Bounds().Dy(), t, 1)
			return dst
		}
	}

	// slow path, lazy
	return &rotateFlipImage{src, t}
}

type rotateFlipImage struct {
	src image.Image
	typ RotateFlipType
}

func (rft *rotateFlipImage) ColorModel() color.Model {
	return rft.src.ColorModel()
}

func (rft *rotateFlipImage) Bounds() image.Rectangle {
	return rotateBounds(rft.src.Bounds(), rft.typ&1 != 0)
}

func (rft *rotateFlipImage) At(x, y int) color.Color {
	bounds := rft.src.Bounds()
	switch rft.typ {
	default:
		return rft.src.At(bounds.Min.X+x, bounds.Min.Y+y)
	case RotateNoneFlipX:
		return rft.src.At(bounds.Max.X-x, bounds.Min.Y+y)
	case RotateNoneFlipY:
		return rft.src.At(bounds.Min.X+x, bounds.Max.Y-y)
	case RotateNoneFlipXY:
		return rft.src.At(bounds.Max.X-x, bounds.Max.Y-y)
	case Rotate90FlipX:
		return rft.src.At(bounds.Min.X+y, bounds.Min.Y+x)
	case Rotate90FlipNone:
		return rft.src.At(bounds.Min.X+y, bounds.Max.Y-x)
	case Rotate90FlipY:
		return rft.src.At(bounds.Max.X-y, bounds.Max.Y-x)
	case Rotate90FlipXY:
		return rft.src.At(bounds.Max.X-y, bounds.Min.Y+x)
	}
}

func rotateFlip(dst []uint8, dst_stride, dst_width, dst_height int, src []uint8, src_stride, src_width, src_height int, t RotateFlipType, bpp int) {
	rotate := t&1 != 0
	flip_y := t&2 != 0
	flip_x := parity(t)

	var dst_row, src_row int

	if flip_x {
		dst_row += bpp * (dst_width - 1)
	}
	if flip_y {
		dst_row += dst_stride * (dst_height - 1)
	}

	var dst_x_offset, dst_y_offset int

	if rotate {
		if flip_x {
			dst_y_offset = -bpp
		} else {
			dst_y_offset = +bpp
		}
		if flip_y {
			dst_x_offset = -dst_stride
		} else {
			dst_x_offset = +dst_stride
		}
	} else {
		if flip_x {
			dst_x_offset = -bpp
		} else {
			dst_x_offset = +bpp
		}
		if flip_y {
			dst_y_offset = -dst_stride
		} else {
			dst_y_offset = +dst_stride
		}
	}

	if dst_x_offset == bpp {
		for y := 0; y < src_height; y++ {
			copy(dst[dst_row:], src[src_row:src_row+src_width*bpp])
			dst_row += dst_y_offset
			src_row += src_stride
		}
	} else {
		for y := 0; y < src_height; y++ {
			dst_pixel := dst_row
			src_pixel := src_row

			for x := 0; x < src_width; x++ {
				copy(dst[dst_pixel:], src[src_pixel:src_pixel+bpp])
				dst_pixel += dst_x_offset
				src_pixel += bpp
			}

			dst_row += dst_y_offset
			src_row += src_stride
		}
	}
}

func rotateBounds(bounds image.Rectangle, rotate bool) image.Rectangle {
	var dx, dy int
	if rotate {
		dx = bounds.Dy()
		dy = bounds.Dx()
	} else {
		dx = bounds.Dx()
		dy = bounds.Dy()
	}
	return image.Rectangle{image.ZP, image.Point{dx, dy}}

}

func rotateYCbCrSubsampleRatio(subsampleRatio image.YCbCrSubsampleRatio, rotate bool) (image.YCbCrSubsampleRatio, bool) {
	if rotate {
		switch subsampleRatio {
		default:
			return 0, false
		case image.YCbCrSubsampleRatio422:
			return image.YCbCrSubsampleRatio440, true
		case image.YCbCrSubsampleRatio440:
			return image.YCbCrSubsampleRatio422, true
		case image.YCbCrSubsampleRatio444, image.YCbCrSubsampleRatio420:
		}
	}
	return subsampleRatio, true
}

func parity(t RotateFlipType) bool {
	t = 026 >> uint8(t)
	return t&1 != 0
}
