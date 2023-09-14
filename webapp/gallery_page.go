package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

// ------------------------------------------------------------------------------------
// Page implementation
// ------------------------------------------------------------------------------------

const GalleryPageName = "gallery"

var GalleryPageTemplate = NewGalleryPage(nil)

func (p GalleryPage) GetBasicPage() BasicPage {
	return &p.BasicPageStruct
}

func (p GalleryPage) Constructor() PageConstructFunc {
	return NewGalleryPage
}

// ------------------------------------------------------------------------------------

type GalleryPage = *GalleryPageStruct

type GalleryPageStruct struct {
	BasicPageStruct
}

func NewGalleryPage(sess Session) Page {
	t := &GalleryPageStruct{}
	InitPage(&t.BasicPageStruct, GalleryPageName, sess, t.generate)
	return t
}

const sampleImageId = "sample_image"

var alertWidget AlertWidget
var myRand = NewJSRand().SetSeed(1234)

func (p GalleryPage) generate() {
	m := p.GenerateHeader()

	alertWidget = NewAlertWidget("sample_alert", AlertInfo)
	alertWidget.SetVisible(false)
	m.Add(alertWidget)

	m.Open()
	m.Col(6)
	m.Id("sample_upload").Label("Photo").AddFileUpload(p.uploadListener)
	imgWidget := m.Id("sample_image").AddImage()
	imgWidget.URLProvider = p.provideURL
	m.Close()

	m.Col(4)
	m.Label("uniform delta").AddText()
	m.Col(8)
	m.Id("x58").Label(`X58`).AddButton(buttonListener).SetEnabled(false)

	m.Col(2).AddSpace()
	m.Col(3).AddSpace()
	m.Col(3).AddSpace()
	m.Col(4).AddSpace()

	m.Col(6)
	m.Label("Bird").Id("bird")
	m.AddInput(birdListener)

	m.Col(6)
	m.Open()
	m.Id("x59").Label(`Label for X59`).AddCheckbox(p.checkboxListener)
	m.Id("x60").Label(`With fruit`).AddSwitch(p.checkboxListener)
	m.Close()

	m.Col(4)
	m.Id("launch").Label(`Launch`).AddButton(buttonListener)

	m.Col(8)
	m.Label(`Sample text; is 5 < 26? A line feed
"Quoted string"
Multiple line feeds:


   an indented final line`)
	m.AddText()

	m.Col(4)
	m.Label("Animal").Id("zebra").AddInput(zebraListener)
}

func birdListener(s Session, widget InputWidget, newVal string) (string, error) {
	var err error
	Todo("?can we have sessions produce listener functions with appropriate handling of sess any?")
	if newVal == "parrot" {
		err = Error("No parrots, please!")
	}
	return newVal, err
}

func zebraListener(s Session, widget InputWidget, newVal string) (string, error) {

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal

	alertWidget.SetVisible(newVal != "")

	s.State.Put(alertWidget.BaseId,
		strings.TrimSpace(newVal+" "+
			RandomText(myRand, 55, false)))
	s.WidgetManager().Repaint(alertWidget)
	return newVal, nil
}

func buttonListener(s Session, widget Widget) {
	wid := widget.Id()
	newVal := "Clicked: " + wid

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal
	alertWidget.SetVisible(true)

	s.State.Put(alertWidget.BaseId,
		strings.TrimSpace(newVal))
	s.WidgetManager().Repaint(alertWidget)
}

func (p GalleryPage) checkboxListener(s Session, widget CheckboxWidget, state bool) (bool, error) {
	Pr("new state:", state)
	return state, nil
}

func (p GalleryPage) uploadListener(s Session, fileUploadWidget FileUpload, value []byte) error {
	pr := PrIf(false)

	m := s.WidgetManager()

	var jpeg []byte
	var imageId int
	var img jimg.JImage
	var err error

	problem := ""
	for {
		problem = "Decoding image"
		if img, err = jimg.DecodeImage(value); err != nil {
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

	var errOut error

	if problem != "" {
		Pr("Problem with upload:", problem)
		if err != nil {
			Pr("...error was:", err)
		}
		errOut = Error(problem)
	} else {
		// Store the id of the blob in the image widget
		s.State.Put(sampleImageId, imageId)
		m.RepaintIds(sampleImageId)
	}
	return errOut
}

func (p GalleryPage) provideURL() string {
	pr := PrIf(true)
	url := ""
	s := p.Session
	imageId := s.State.OptInt(sampleImageId, 0)

	pr("provideURL, image id read from state:", imageId)

	if imageId != 0 {
		url = ReadImageIntoCache(imageId)
		pr("read into cache, url:", url)
	}
	return url
}
