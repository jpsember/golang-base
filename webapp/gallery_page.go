package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

type GalleryPageStruct struct {
	BasicPage
}

type GalleryPage = *GalleryPageStruct

func NewGalleryPage(sess Session, parentWidget Widget) GalleryPage {
	t := &GalleryPageStruct{
		NewBasicPage(sess, parentWidget),
	}
	t.devLabel = "gallery_page"
	return t
}

const sampleImageId = "sample_image"

func (p GalleryPage) Generate() {
	m := p.GenerateHeader()

	alertWidget = NewAlertWidget("sample_alert", AlertInfo)
	alertWidget.SetVisible(false)
	m.Add(alertWidget)

	m.Open()
	m.Col(6)
	m.Id("sample_upload").Label("Photo").Listener(p.uploadListener).AddFileUpload()
	imgWidget := m.Id("sample_image").AddImage()
	imgWidget.URLProvider = p.provideURL
	m.Close()

	m.Col(4)
	m.Label("uniform delta").AddText()
	m.Col(8)
	m.Id("x58").Label(`X58`).Listener(buttonListener).AddButton().SetEnabled(false)

	m.Col(2).AddSpace()
	m.Col(3).AddSpace()
	m.Col(3).AddSpace()
	m.Col(4).AddSpace()

	m.Col(6)
	m.Listener(birdListener)
	m.Label("Bird").Id("bird")
	m.AddInput()

	m.Col(6)
	m.Open()
	m.Id("x59").Label(`Label for X59`).Listener(checkboxListener).AddCheckbox()
	m.Id("x60").Label(`With fruit`).Listener(checkboxListener).AddSwitch()
	m.Close()

	m.Col(4)
	m.Id("launch").Label(`Launch`).Listener(buttonListener).AddButton()

	m.Col(8)
	m.Label(`Sample text; is 5 < 26? A line feed
"Quoted string"
Multiple line feeds:


   an indented final line`)
	m.AddText()

	m.Col(4)
	m.Listener(zebraListener)
	m.Label("Animal").Id("zebra").AddInput()
}

func birdListener(s Session, widget Widget) {
	Todo("?can we have sessions produce listener functions with appropriate handling of sess any?")
	newVal := s.GetValueString()
	s.SetWidgetProblem(widget, nil)
	s.State.Put(widget.Id(), newVal)
	Todo("!do validation as a global function somewhere")
	if newVal == "parrot" {
		s.SetWidgetProblem(widget, "No parrots, please!")
	}
	s.WidgetManager().Repaint(widget)
}

func zebraListener(s Session, widget Widget) {

	// Get the requested new value for the widget
	newVal := s.GetValueString()

	// Store this as the new value for this widget within the session state map
	s.State.Put(widget.Id(), newVal)

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal

	alertWidget.SetVisible(newVal != "")

	s.State.Put(alertWidget.BaseId,
		strings.TrimSpace(newVal+" "+
			RandomText(myRand, 55, false)))
	s.WidgetManager().Repaint(alertWidget)
}

func buttonListener(s Session, widget Widget) {
	wid := s.GetWidgetId()
	newVal := "Clicked: " + wid

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal
	alertWidget.SetVisible(true)

	s.State.Put(alertWidget.BaseId,
		strings.TrimSpace(newVal))
	s.WidgetManager().Repaint(alertWidget)
}

func checkboxListener(s Session, widget Widget) {
	wid := s.GetWidgetId()

	// Get the requested new value for the widget
	newVal := s.GetValueBoolean()

	Todo("It is safe to not check if there was a RequestProblem, as any state changes will still go through validation...")

	s.State.Put(wid, newVal)
	// Repainting isn't necessary, as the web page has already done this
}

func (p GalleryPage) uploadListener(s Session, widget Widget) {
	pr := PrIf(true)

	m := s.WidgetManager()

	fileUploadWidget := widget.(FileUpload)

	var jpeg []byte
	var imageId int
	var img jimg.JImage
	var err error

	problem := ""
	for {
		problem = "Decoding image"
		if img, err = jimg.DecodeImage(fileUploadWidget.ReceivedBytes()); err != nil {
			break
		}
		pr("decoded:", INDENT, img)

		problem = "Converting to default type"
		if img, err = img.AsDefaultType(); err != nil {
			break
		}
		pr("converted to default type")

		problem = "Problem with dimensions"
		if Clamp(img.Size().X, 50, 3000) != img.Size().X || //
			Clamp(img.Size().Y, 50, 3000) != img.Size().Y {
			break
		}

		pr("dimensions ok")

		img = img.ScaleToSize(IPointWith(400, 600))

		problem = "Converting image"
		if jpeg, err = img.ToJPEG(); err != nil {
			break
		}
		pr("encoded as jpeg")

		problem = "Storing image"

		b := NewBlob()
		b.SetData(jpeg)
		AssignBlobName(b)
		var created Blob
		if created, err = CreateBlob(b); err != nil {
			break
		}
		imageId = created.Id()
		pr("created blob, id:", BlobSummary(created))
		problem = ""
		break
	}
	if problem != "" {
		Pr("Problem with upload:", problem)
		if err != nil {
			Pr("...error was:", err)
		}
		s.SetWidgetProblem(widget, "Trouble uploading image: "+problem)
	} else {
		// Store the id of the blob in the image widget
		s.State.Put(sampleImageId, imageId)
	}
	m.RepaintIds(sampleImageId)
}

func (p GalleryPage) provideURL() string {
	pr := PrIf(true)
	url := ""
	s := p.session
	imageId := s.State.OptInt(sampleImageId, 0)

	pr("provideURL, image id read from state:", imageId)

	if imageId != 0 {
		url = ReadImageIntoCache(imageId)
		pr("read into cache, url:", url)
	}
	return url
}
