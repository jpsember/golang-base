package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

const anim_state_prefix = "_create_animal_."
const (
	id_animal_name        = anim_state_prefix + "name"
	id_animal_summary     = anim_state_prefix + "summary"
	id_animal_details     = anim_state_prefix + "details"
	id_add                = anim_state_prefix + "add"
	id_animal_uploadpic   = anim_state_prefix + "photo"
	id_animal_display_pic = anim_state_prefix + "pic"
)

type CreateAnimalPageStruct struct {
	BasicPage
}

type CreateAnimalPage = *CreateAnimalPageStruct

func NewCreateAnimalPage(sess Session, parentWidget Widget) AbstractPage {
	t := &CreateAnimalPageStruct{
		BasicPage: NewBasicPage(sess, parentWidget),
	}
	t.devLabel = "create_animal_page"
	return t
}

func (p CreateAnimalPage) Generate() {
	//SetWidgetDebugRendering()
	p.session.DeleteStateFieldsWithPrefix(anim_state_prefix)
	//p.session.DeleteStateErrors()
	m := p.GenerateHeader()

	Todo("!Have ajax listener that can show advice without an actual error, e.g., if user left some fields blank")
	m.Label("Create New Animal Record").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		Todo("when does the INPUT store the state?")
		m.Label("Name").Id(id_animal_name).Listener(AnimalNameListener).AddInput()

		m.Label("Summary").Id(id_animal_summary).AddInput()
		m.Size(SizeTiny).Label("A brief paragraph to appear in the 'card' view.").AddText()
		m.Label("Details").Id(id_animal_details).AddInput()
		m.Size(SizeTiny).Label("Additional paragraphs to appear on the 'details' view.").AddText()

		m.Listener(p.addListener)
		m.Id(id_add).Label("Create").AddButton()
	}
	m.Close()

	m.Open()
	m.Id(id_animal_uploadpic).Label("Photo").Listener(p.uploadPhotoListener).AddFileUpload()
	imgWidget := m.Id(id_animal_display_pic).AddImage()
	imgWidget.URLProvider = p.provideURL
	m.Close()
}

func SessionStrValue(s Session, id string) string {
	return s.State.OptString(id, "")
}

func WidgetStrValue(s Session, widget Widget) string {
	return SessionStrValue(s, widget.Id())
}

func SessionIntValue(s Session, id string) int {
	return s.State.OptInt(id, 0)
}

func WidgetIntValue(s Session, widget Widget) int {
	return SessionIntValue(s, widget.Id())
}

func (p CreateAnimalPage) addListener(s Session, widget Widget) {
	pr := PrIf(true)
	pr("state:", INDENT, s.State)

	p.session.DeleteStateErrors()

	wName := getWidget(s, id_animal_name)
	wSummary := getWidget(s, id_animal_summary)
	wDetails := getWidget(s, id_animal_details)
	wPhoto := getWidget(s, id_animal_display_pic)
	mUpload := getWidget(s, id_animal_uploadpic)

	ValidateAnimalName(s, wName)
	ValidateAnimalInfo(s, wSummary, 20, 200)
	ValidateAnimalInfo(s, wDetails, 200, 2000)
	ValidateAnimalPhoto(s, wPhoto, mUpload)

	errcount := WidgetErrorCount(p.parentPage, s.State)
	if errcount != 0 {
		return
	}

	b := NewAnimal()
	b.SetName(strings.TrimSpace(WidgetStrValue(s, wName)))
	b.SetSummary(strings.TrimSpace(WidgetStrValue(s, wSummary)))
	b.SetDetails(strings.TrimSpace(WidgetStrValue(s, wDetails)))
	b.SetPhotoThumbnail(WidgetIntValue(s, wPhoto))
	ub, err := CreateAnimal(b)
	CheckOk(err)

	Pr("created animal:", INDENT, ub)

	Todo("discard state, i.e. the edited fields; use a common prefix to simplify")
	Todo("Do a 'back' operation to go back to the previous page")
	NewManagerPage(s, p.parentPage).Generate()
}

func AnimalNameListener(s Session, widget Widget) {
	Pr("validate animal name, widget id:", widget.Id(), "state:", s.State)

	// The requested value for the widget has been passed in the ajax map, but is not yet known to us otherwise.
	n := s.GetValueString()

	// Store the requested value to the widget, even if it subsequently fails validation.
	Todo("?A utility method for this")
	s.State.Put(widget.Id(), n)

	ValidateAnimalName(s, widget)
}

func ValidateAnimalName(s Session, widget Widget) {
	n := s.GetStateString(widget.Id())

	errStr := ""

	Todo("?A utility method for this")
	s.State.Put(widget.Id(), n)

	for {
		ln := len(n)
		if ln < 3 || ln > 20 {
			errStr = "Length should be 3...20 characters" + " (currently: " + n + ")"
			break
		}
		break
	}

	if errStr != "" {
		s.SetWidgetProblem(widget, errStr)
	}
}

func ValidateAnimalInfo(s Session, widget Widget, minLength int, maxLength int) {
	errStr := ""
	Todo("a utility method for reading widget values as strings")
	n := WidgetStrValue(s, widget)
	n = strings.TrimSpace(n)
	Pr("validateAnimalInfo, maxLen:", maxLength, "id:", widget.Id())
	for {
		ln := len(n)

		errStr = "Please add more info here."
		if false && Alert("allowing empty fields") {
			minLength = 0
		}
		if ln < minLength {
			break
		}
		errStr = "Please type no more than " + IntToString(maxLength) + " characters."
		if ln > maxLength {
			break
		}

		errStr = ""
		break
	}
	if errStr != "" {
		s.SetWidgetProblem(widget, errStr)
	}
}

func ValidateAnimalPhoto(s Session, valueWidget Widget, reportWidget Widget) {
	n := WidgetIntValue(s, valueWidget)
	Pr("validate animal photo, widget int value:", n, "widget:", valueWidget)
	if n == 0 {
		s.SetWidgetProblem(reportWidget, "Please upload a photo")
	}
}

func (p CreateAnimalPage) uploadPhotoListener(s Session, widget Widget) {
	pr := PrIf(true)

	m := s.WidgetManager()

	fu := widget.(FileUpload)
	by := fu.ReceivedBytes()

	var jpeg []byte
	var imageId int
	var img jimg.JImage
	var err error

	problem := ""
	for {
		problem = "Decoding image"
		if img, err = jimg.DecodeImage(by); err != nil {
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

		img = img.ScaleToSize(AnimalPicSizeNormal)

		problem = "Converting image"
		if jpeg, err = img.ToJPEG(); err != nil {
			break
		}
		pr("encoded as jpeg")

		problem = "Storing image"

		Todo("?Later, keep the original image around for crop adjustments; but for now, scale and store immediately")
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
		// Discard the old blob whose id we are now replacing
		DiscardBlob(SessionIntValue(s, id_animal_display_pic))

		// Store the id of the blob in the image widget
		s.State.Put(id_animal_display_pic, imageId)
	}
	pr("repainting animal_display_pic")
	m.RepaintIds(id_animal_display_pic)
}

func (p CreateAnimalPage) provideURL() string {
	pr := PrIf(false)
	url := ""
	s := p.session
	imageId := SessionIntValue(s, id_animal_display_pic)

	pr("provideURL, image id read from state:", imageId)

	if imageId != 0 {
		url = ReadImageIntoCache(imageId)
		pr("read into cache, url:", url)
	}
	return url
}

func DiscardBlob(id int) {
	if id != 0 {
		Todo("#50Discard blob id", id)
	}
}
