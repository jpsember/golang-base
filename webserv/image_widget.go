package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// This general type of listener can serve as a validator as well
type ImageURLProvider func(s Session) string

type ImageWidgetObj struct {
	BaseWidgetObj
	urlProvider     ImageURLProvider
	fixedSize       IPoint // If not (0,0), size to display image as
	escapedAltLabel string
	clickListener   ClickWidgetListener
}

type ImageWidget = *ImageWidgetObj

func NewImageWidget(id string, urlProvider ImageURLProvider, clickListener ClickWidgetListener) ImageWidget {
	t := &ImageWidgetObj{
		escapedAltLabel: Escaped("unknown image"),
		urlProvider:     urlProvider,
		clickListener:   clickListener,
	}
	if clickListener != nil {
		t.SetLowListener(func(sess Session, widget Widget, value string, args WidgetArgs) (any, error) {
			t.clickListener(sess, widget, args)
			return nil, nil
		})
	}
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

	prependedId := s.PrependId(w.Id())
	m.TgOpen(`div id='`).A(prependedId, `'`).TgContent()
	m.Comment("image")

	{
		var imageSource string

		if w.urlProvider != nil {
			imageSource = w.urlProvider(s)
			pr("url provider returned image source:", imageSource)
			if imageSource == "" {
				imageSource = "https://upload.wikimedia.org/wikipedia/en/a/a9/Example.jpg"
			}
		} else {
			pr("no URLProvider!")
		}

		m.A(`<img src="`, imageSource, `" alt="`, w.escapedAltLabel, `"`)
		if w.clickListener != nil {
			m.A(` onclick="jsButton('`, prependedId, `')"`)
		}

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
