package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// This general type of listener can serve as a validator as well
type ImageURLProvider func(s Session) string

type ImageWidgetObj struct {
	BaseWidgetObj
	URLProvider ImageURLProvider
	fixedSize   IPoint // If not (0,0), size to display image as
}

type ImageWidget = *ImageWidgetObj

func NewImageWidget(id string) ImageWidget {
	t := &ImageWidgetObj{}
	t.InitBase(id)
	return t
}

// Set the size that the image will occupy.  This size will be scaled by the user's screen resolution.
func (w ImageWidget) SetSize(originalSize IPoint, scaleFactor float64) {
	w.fixedSize = originalSize.ScaledBy(scaleFactor)
}

func (w ImageWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf("", false)
	pr("rendering:", w.Id())

	// The outermost element must have the widget's id!  Or chaos happens during repainting.

	m.TgOpen(`div id=`).A(QUOTED, w.Id()).TgContent()
	m.Comment("image")

	{
		var imageSource string

		if w.URLProvider != nil {
			imageSource = w.URLProvider(s)
			pr("url provider returned image source:", imageSource)
			if imageSource == "" {
				imageSource = "https://upload.wikimedia.org/wikipedia/en/a/a9/Example.jpg"
			}
		} else {
			pr("no URLProvider!")
		}

		m.A(`<img src="`, imageSource, `" alt="uploaded image"`)

		PlotImageSizeMarkup(s, m, w.fixedSize)
		if w.fixedSize.IsPositive() {
			//
			//Todo("?Investigate relationship between pixel ratio, screen size")
			//screenWidth := s.BrowserInfo.ScreenSizeX()
			//scaleFactor := float64(screenWidth) / 2000
			//plotSize := w.fixedSize.ScaledBy(scaleFactor)

			m.A(` class="img-thumbnail"`)
		} else {
			m.A(` class="img-thumbnail image-fluid"`)
		}
		m.A(`>`)

	}
	m.TgClose()
	pr("done render")
}

func PlotImageSizeMarkup(s Session, m MarkupBuilder, normalizedTargetSize IPoint) {
	sz := normalizedTargetSize
	if sz.IsPositive() {
		Todo("?Investigate relationship between pixel ratio, screen size")
		screenWidth := s.BrowserInfo.ScreenSize().X
		scaleFactor := float64(screenWidth) / 2000
		plotSize := sz.ScaledBy(scaleFactor)
		m.A(` width="`, plotSize.X, `" height="`, plotSize.Y, `" `)
	}
}
