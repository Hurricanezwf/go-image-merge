package utils

import (
	"errors"
)

type ImageFormat string

const (
	ImageFormatUnknown = ""
	ImageFormatBMP     = ".bmp"
	ImageFormatJPEG    = ".jpg"
	ImageFormatGIF     = ".gif"
	ImageFormatPNG     = ".png"
	ImageFormatWEBP    = ".webp"
)

// Ext 返回图片后缀名, 包含点;
func (f ImageFormat) Ext() string {
	return string(f)
}

// ResolveImageFormat 从图片数据中解析出图片格式.
func ResolveImageFormat(data []byte) ImageFormat {
	if IsBMPImage(data) {
		return ImageFormatBMP
	}
	if IsJPEGImage(data) {
		return ImageFormatJPEG
	}
	if IsGIFImage(data) {
		return ImageFormatGIF
	}
	if IsPNGImage(data) {
		return ImageFormatPNG
	}
	if IsWebpImage(data) {
		return ImageFormatWEBP
	}
	return ImageFormatUnknown
}

// IsBMPImage 判断图片数据是否是BMP格式.
func IsBMPImage(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return data[0] == 0x42 && data[1] == 0x4d
}

// IsJPEGImage 判断图片数据是否是JPEG格式.
func IsJPEGImage(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return (data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff)
}

// IsGIFImage 判断图片数据是否是GIF格式.
func IsGIFImage(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return (data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38)
}

// IsPNGImage 判断图片数据是否是PNG格式.
func IsPNGImage(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return (data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4e && data[3] == 0x47)
}

// IsWebpImage 判断图片数据是否是Webp格式.
func IsWebpImage(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return (data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50)
}

// EstimateRatio 估算图片的宽高比.
func EstimateRatio(w, h float64) (string, error) {
	if (w == h) || (w > h && (w-h) < 2) || (h > w && (h-w) < 2) {
		return "1:1", nil
	}

	//
	if w > h {
		delta := w/h - float64(4)/float64(3)
		if delta < 0 {
			delta = -1 * delta
		}
		if delta <= 0.1 {
			return "4:3", nil
		}
	}
	if h > w {
		delta := h/w - float64(4)/float64(3)
		if delta < 0 {
			delta = -1 * delta
		}
		if delta <= 0.1 {
			return "3:4", nil
		}
	}

	//
	if w > h {
		delta := w/h - float64(16)/float64(9)
		if delta < 0 {
			delta = -1 * delta
		}
		if delta <= 0.1 {
			return "16:9", nil
		}
	}
	if h > w {
		delta := h/w - float64(16)/float64(9)
		if delta < 0 {
			delta = -1 * delta
		}
		if delta <= 0.1 {
			return "9:16", nil
		}
	}
	return "", errors.New("unsupported aspect ratio")
}
