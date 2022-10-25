package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//compress a jpg or png format image, the new images will be named autoly
func compressImg(source string, forh uint) error {
	var width int
	var err error
	sb, err := ioutil.ReadFile(source)
	if err != nil {
		panic(err)
	}
	var file *os.File
	reg, _ := regexp.Compile(`^.*\.((png)|(jpg))$`)
	if !reg.MatchString(source) {
		err = errors.New("%s is not a .png or .jpg file")
		panic(err)
	}
	if file, err = os.Open(source); err != nil {
		panic(err)
	}
	defer file.Close()
	name := file.Name()
	var img image.Image
	switch {
	case strings.HasSuffix(name, ".png"):
		width, _ = getWidthHeightForPng(sb)
		if img, err = png.Decode(file); err != nil {
			panic(err)
		}
	case strings.HasSuffix(name, ".jpg"):
		width, _ = getWidthHeightForJpg(sb)
		if img, err = jpeg.Decode(file); err != nil {
			panic(err)
		}
	default:
		err = fmt.Errorf("Images %s name not right!", name)
		panic(err)
	}
	if forh > 0 {
		width = int(forh)
	}
	resizeImg := resize.Resize(uint(width), uint(0), img, resize.Lanczos3)
	newName := newName(source)
	if outFile, err := os.Create(newName); err != nil {
		panic(err)
	} else {
		defer outFile.Close()
		err = jpeg.Encode(outFile, resizeImg, nil)
		if err != nil {
			panic(err)
		}
	}
	abspath, _ := filepath.Abs(newName)
	log.Printf("New imgs successfully save at: %s", abspath)
	return nil
}

//create a file name for the iamges that after resize
func newName(name string) string {
	dir, file := filepath.Split(name)
	return fmt.Sprintf("%s/%s/%s", dir, "data", file)
}

/**
* 入参： JPG 图片文件的二进制数据
* 出参：JPG 图片的宽和高
**/
func getWidthHeightForJpg(imgBytes []byte) (int, int) {
	var offset int
	imgByteLen := len(imgBytes)
	for i := 0; i < imgByteLen-1; i++ {
		if imgBytes[i] != 0xff {
			continue
		}
		if imgBytes[i+1] == 0xC0 || imgBytes[i+1] == 0xC1 || imgBytes[i+1] == 0xC2 {
			offset = i
			break
		}
	}
	offset += 5
	if offset >= imgByteLen {
		return 0, 0
	}
	height := int(imgBytes[offset])<<8 + int(imgBytes[offset+1])
	width := int(imgBytes[offset+2])<<8 + int(imgBytes[offset+3])
	return width, height
}

// 获取 PNG 图片的宽高
func getWidthHeightForPng(imgBytes []byte) (int, int) {
	pngHeader := "\x89PNG\r\n\x1a\n"
	if string(imgBytes[:len(pngHeader)]) != pngHeader {
		return 0, 0
	}
	offset := 12
	if "IHDR" != string(imgBytes[offset:offset+4]) {
		return 0, 0
	}
	offset += 4
	width := int(binary.BigEndian.Uint32(imgBytes[offset : offset+4]))
	height := int(binary.BigEndian.Uint32(imgBytes[offset+4 : offset+8]))
	return width, height
}

func getFilelist(path string) []string {
	files := make([]string, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func main() {
	fs := getFilelist("test")
	for _, f := range fs {
		if err := compressImg(f, 500); err != nil {
			panic(err)
		}
	}

}
