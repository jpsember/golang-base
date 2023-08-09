package jimg

import (
	"bufio"
	"bytes"
	. "github.com/jpsember/golang-base/base"
	"golang.org/x/image/draw"
	"image"
	_ "image/jpeg"
	"image/png"
	// Package image/jpeg is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand JPEG formatted images. Uncomment these
	// two lines to also understand GIF and PNG images:
	// _ "image/gif"
	_ "image/png"
)

func DecodeImage(imgbytes []byte) (JImage, error) {
	img, format, err := image.Decode(bytes.NewReader(imgbytes))
	var jmg JImage
	if err == nil {
		jmg = JImageOf(img)
	}
	if false {
		Pr("format:", format)
	}
	return jmg, err
}

type JImageStruct struct {
	image     image.Image
	imageType JImageType
	size      IPoint
}

type JImage = *JImageStruct

type JImageType int

const (
	typeUnitialized JImageType = iota
	TypeRGBA
	TypeNRGBA
	TypeCMYK
	TypeYCbCr
	TypeUnknown = -1
)

var itmap = map[JImageType]string{
	TypeNRGBA:   "NRGBA",
	TypeCMYK:    "CMYK",
	TypeYCbCr:   "YCbCr",
	TypeUnknown: "Unknown",
}

func ImageTypeStr(imgType JImageType) string {
	result := itmap[imgType]
	if result == "" {
		result = "???"
	}
	return result
}

func JImageOf(img image.Image) JImage {
	CheckNotNil(img)
	CheckArg(img.Bounds().Min == image.Point{}, "origin of image is not at (0,0)")
	t := &JImageStruct{
		image: img,
	}
	return t
}

func (ji JImage) Image() image.Image {
	return ji.image
}

func (ji JImage) Type() JImageType {

	if ji.imageType == typeUnitialized {
		ty := ji.imageType
		switch ji.image.(type) {
		case *image.RGBA:
			ty = TypeRGBA
		case *image.NRGBA:
			ty = TypeNRGBA
		case *image.CMYK:
			ty = TypeCMYK
		case *image.YCbCr:
			ty = TypeYCbCr
		default:
			Pr("Color model:", ji.image.ColorModel())
			ty = TypeUnknown
		}
		ji.imageType = ty
	}
	return ji.imageType
}

func (ji JImage) Size() IPoint {
	if ji.size == IPointZero {
		b := ji.image.Bounds()
		ji.size = IPointWith(b.Dx(), b.Dy())
	}
	return ji.size
}

func (ji JImage) ToJson() JSMap {
	m := NewJSMap()
	m.Put("", "JImage")
	m.Put("type", ImageTypeStr(ji.Type()))
	m.Put("size", ji.Size())
	return m
}

func GetImageInfo(image image.Image) JSMap {
	ji := JImageOf(image)
	return ji.ToJson()
}

func (ji JImage) AsType(desiredType JImageType) (JImage, error) {
	var result JImage
	errstring := "unsupported image type"
	if ji.Type() == desiredType {
		result = ji
	} else {
		var m draw.Image
		switch desiredType {
		case TypeNRGBA:
			m = image.NewNRGBA(image.Rect(0, 0, ji.Size().X, ji.Size().Y))
		}
		if m != nil {
			draw.Draw(m, m.Bounds(), ji.Image(), image.Point{}, draw.Src)
			result = JImageOf(m)
		}
	}
	if result == nil {
		return nil, Error(errstring)
	} else {
		return result, nil
	}
}

func (ji JImage) ToPNG() ([]byte, error) {
	if ji.Type() != TypeNRGBA {
		return nil, Error("Cannot convert to PNG", ji.ToJson())
	}
	var bb bytes.Buffer
	err := png.Encode(&bb, ji.Image())
	if err != nil {
		Pr("Failed to encode image as PNG")
	}
	return bb.Bytes(), err
}

func (ji JImage) ScaledTo(size IPoint) JImage {

	var targetX, targetY int

	origSize := ji.Size()
	if size.X == 0 {
		if size.Y > 0 {
			targetY = size.Y
			targetX = MaxInt(1, (origSize.X*targetY)/origSize.Y)
		}
	} else {
		if size.X > 0 {
			targetX = size.X
			targetY = MaxInt(1, (origSize.Y*targetX)/origSize.X)
		}
	}
	CheckArg(targetX > 0 && targetY > 0, "Cannot scale image of size", ji.Size(), "to", size)
	scaledImage := image.NewNRGBA(image.Rect(0, 0, targetX, targetY))
	inputImage := ji.Image()
	draw.ApproxBiLinear.Scale(scaledImage, scaledImage.Bounds(), inputImage, inputImage.Bounds(), draw.Over, nil)
	return JImageOf(scaledImage)
}

func (ji JImage) EncodePNG() ([]byte, error) {
	w := bytes.Buffer{}
	err := png.Encode(bufio.NewWriter(&w), ji.Image())
	if err == nil {
		return w.Bytes(), nil
	}
	return nil, err
}

//func (ji JImage) FitTo(size IPoint, strategy int) JImage {
//
//
//	w := float64(ji.Size().X)
//      h := float64(ji.Size().Y)
//     u := float64(size.X)
//      v := float64(size.Y)
//
//      lambdaCrop := float64(1)
//      lambdaLbox := float64(1)
//
//	  switch strategy {
//    default:
//		BadArg("strategy:",strategy)
//    case CROP:
//      lambdaLbox = 0
//    case LETTERBOX:
//      lambdaCrop = 0
//    case HYBRID:
//    }
//
//     sourceAspect := h / w;
//      targetAspect := v / u;
//    if sourceAspect < targetAspect {
//       temp := lambdaCrop;
//      lambdaCrop = lambdaLbox;
//      lambdaLbox = temp;
//    }
//
//    // I apply a cost function c as a function of the scale factor s:
//    //
//    //  c(s)   L_c(u - sw)^2 + L_l(v - sh)^2
//    //
//    // and take the derivative to find when c(s) is minimized, to yield optimal scale s*:
//    //
//    //  s* = L_c(wu) + L_l(hv)
//    //       -------------------
//    //       L_c(w^2) + L_l(h^2)
//    //
//      s := (lambdaCrop * w * u + lambdaLbox * h * v) / (lambdaCrop * w * w + lambdaLbox * h * h);
//
//      resultWidth := s * w;
//     resultHeight := s * h;
//
//    mTargetRectangle  := RectWithFloat( (u - resultWidth) * .5 , (v - resultHeight) * .5 , resultWidth,
//        resultHeight)
//	 mTargetRectangle.AssertValid()
//
//}
//
///**
// * MIT License
// *
// * Copyright (c) 2021 Jeff Sember
// *
// * Permission is hereby granted, free of charge, to any person obtaining a copy
// * of this software and associated documentation files (the "Software"), to deal
// * in the Software without restriction, including without limitation the rights
// * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// * copies of the Software, and to permit persons to whom the Software is
// * furnished to do so, subject to the following conditions:
// *
// * The above copyright notice and this permission notice shall be included in all
// * copies or substantial portions of the Software.
// *
// * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// * SOFTWARE.
// *
// **/
//package js.graphics;
//
//import static js.base.Tools.*;
//
//import java.awt.Graphics;
//import java.awt.image.BufferedImage;
//
//import js.base.BaseObject;
//import js.geometry.FRect;
//import js.geometry.IPoint;
//import js.geometry.IRect;
//import js.geometry.Matrix;
//import js.graphics.gen.ImageFitOptions;
//
///**
// * Apply heuristics to choose a 'best fit' of a source image to a target
// * rectangle. Doesn't actually involve images or pixels, just mathematical
// * rectangles
// */
//public final class ImageFit extends BaseObject {
//
//  public ImageFit(ImageFitOptions opt, IPoint sourceSize) {
//    mOptions = opt.build();
//    mSourceRectangleSize = sourceSize;
//  }
//
//  /**
//   * Construct an appropriate ImageFit instance, recycling existing one if
//   * exists and is appropriate
//   */
//  public static ImageFit constructForSize(ImageFit existing, ImageFitOptions options,
//      IPoint sourceImageSize) {
//    if (existing == null || !existing.sourceSize().equals(sourceImageSize))
//      existing = new ImageFit(options, sourceImageSize);
//    return existing;
//  }
//
//  public ImageFitOptions options() {
//    return mOptions;
//  }
//
//  public IPoint sourceSize() {
//    return mSourceRectangleSize;
//  }
//
//  public IRect transformedSourceRect() {
//    if (mTargetRectangle != null)
//      return mTargetRectangle;
//
//    if (!mOptions.scaleUp() || !mOptions.scaleDown())
//      throw notSupported("Scale up, scale down must be true");
//
//    float w = mSourceRectangleSize.x;
//    float h = mSourceRectangleSize.y;
//    float u = mOptions.targetSize().x;
//    float v = mOptions.targetSize().y;
//
//    float lambdaCrop = 1f;
//    float lambdaLbox = 1f;
//
//    switch (mOptions.fitType()) {
//    default:
//      throw notSupported(mOptions.fitType());
//    case CROP:
//      lambdaLbox = 0f;
//      break;
//    case LETTERBOX:
//      lambdaCrop = 0f;
//      break;
//    case HYBRID:
//      break;
//    }
//
//    float sourceAspect = h / w;
//    float targetAspect = v / u;
//    if (sourceAspect < targetAspect) {
//      float temp = lambdaCrop;
//      lambdaCrop = lambdaLbox;
//      lambdaLbox = temp;
//    }
//
//    // I apply a cost function c as a function of the scale factor s:
//    //
//    //  c(s)   L_c(u - sw)^2 + L_l(v - sh)^2
//    //
//    // and take the derivative to find when c(s) is minimized, to yield optimal scale s*:
//    //
//    //  s* = L_c(wu) + L_l(hv)
//    //       -------------------
//    //       L_c(w^2) + L_l(h^2)
//    //
//    float s = (lambdaCrop * w * u + lambdaLbox * h * v) / (lambdaCrop * w * w + lambdaLbox * h * h);
//
//    float resultWidth = s * w;
//    float resultHeight = s * h;
//
//    mTargetRectangleF = new FRect((u - resultWidth) * .5f, (v - resultHeight) * .5f, resultWidth,
//        resultHeight);
//    mTargetRectangle = mTargetRectangleF.toIRect();
//    return mTargetRectangle;
//  }
//
//  public FRect transformedSourceRectF() {
//    transformedSourceRect();
//    return mTargetRectangleF;
//  }
//
//  public boolean cropNecessary() {
//    return !transformedSourceRect().equals(new IRect(sourceSize()));
//  }
//
//  /**
//   * Get matrix that transforms source image points to target image
//   */
//  public Matrix matrix() {
//    if (mMatrix == null) {
//      IRect r = transformedSourceRect();
//      mMatrix = new Matrix(r.width / (float) mSourceRectangleSize.x, 0, 0,
//          r.height / (float) mSourceRectangleSize.y, r.x, r.y);
//    }
//    return mMatrix;
//  }
//
//  public Matrix inverse() {
//    if (mMatrixInv == null)
//      mMatrixInv = matrix().invert();
//    return mMatrixInv;
//  }
//
//  /**
//   * Apply ImageFit to a BufferedImage
//   *
//   * @param sourceImage
//   * @param targetImageType
//   *          type of BufferedImage to return
//   */
//  public BufferedImage apply(BufferedImage sourceImage, int targetImageType) {
//    IRect destRect = transformedSourceRect();
//    IPoint targetSize = options().targetSize();
//    BufferedImage resultImage = ImgUtil.build(targetSize, targetImageType);
//    Graphics g = resultImage.getGraphics();
//    IPoint sourceSize = ImgUtil.size(sourceImage);
//    g.drawImage(sourceImage, destRect.x, destRect.y, destRect.endX(), destRect.endY(), 0, 0, sourceSize.x,
//        sourceSize.y, null);
//    g.dispose();
//    return resultImage;
//  }
//
//  private final ImageFitOptions mOptions;
//  private final IPoint mSourceRectangleSize;
//  private IRect mTargetRectangle;
//  private FRect mTargetRectangleF;
//  private Matrix mMatrix;
//  private Matrix mMatrixInv;
//
//}
