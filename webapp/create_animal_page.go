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
	m := p.GenerateHeader()

	Todo("!Have ajax listener that can show advice without an actual error, e.g., if user left some fields blank")
	m.Label("Create New Animal Record").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("Name").Id(id_animal_name).AddInput(AnimalNameListener)

		m.Label("Summary").Id(id_animal_summary).AddInput(p.AnimalTextListener)
		m.Size(SizeTiny).Label("A brief paragraph to appear in the 'card' view.").AddText()
		m.Label("Details").Id(id_animal_details).AddInput(p.AnimalTextListener)
		m.Size(SizeTiny).Label("Additional paragraphs to appear on the 'details' view.").AddText()

		m.Id(id_add).Label("Create").AddButton(p.addButtonListener)
	}
	m.Close()

	m.Open()
	m.Id(id_animal_uploadpic).Label("Photo").AddFileUpload(p.uploadPhotoListener)
	imgWidget := m.Id(id_animal_display_pic).AddImage()
	imgWidget.URLProvider = p.provideURL

	// Scale the photos based on browser resolution
	{
		Todo("?Investigate relationship between pixel ratio, screen size")
		w := p.session.BrowserInfo.ScreenSizeX()
		modifiedSize := AnimalPicSizeNormal.ScaledBy(float64(w) / 4500)
		imgWidget.FixedSize = modifiedSize
	}
	m.Close()
}

func AnimalNameListener(s Session, widget InputWidget, value string) (string, error) {
	return ValidateAnimalName(value, VALIDATE_EMPTYOK)
}

func (p CreateAnimalPage) AnimalTextListener(sess Session, widget InputWidget, value string) (string, error) {
	if widget.Id() == id_animal_summary {
		return animalInfoListener(sess, widget, value, 20, 200, true)
	} else {
		return animalInfoListener(sess, widget, value, 200, 2000, true)
	}
}

func (p CreateAnimalPage) addButtonListener(s Session, widget Widget) error {
	pr := PrIf(true)
	//pr("state:", INDENT, s.State)

	//p.session.DeleteStateErrors()

	wName := getWidget(s, id_animal_name)
	wSummary := getWidget(s, id_animal_summary)
	wDetails := getWidget(s, id_animal_details)
	wPhoto := getWidget(s, id_animal_display_pic)
	mUpload := getWidget(s, id_animal_uploadpic)

	{
		text := s.WidgetStrValue(wName)
		result, err := ValidateAnimalName(text, 0)
		s.SetWidgetProblem(wName, err)
		Todo("We need to store new animal name value here perhaps", result)
	}

	preSubmitValidateText(s, wSummary, 20, 200, 0)
	preSubmitValidateText(s, wDetails, 200, 2000, 0)
	ValidateAnimalPhoto(s, wPhoto, mUpload)

	errcount := WidgetErrorCount(p.parentPage, s.State)
	pr("error count:", errcount)
	if errcount != 0 {
		return nil
	}

	b := NewAnimal()
	b.SetName(strings.TrimSpace(s.WidgetStrValue(wName)))
	b.SetSummary(strings.TrimSpace(s.WidgetStrValue(wSummary)))
	b.SetDetails(strings.TrimSpace(s.WidgetStrValue(wDetails)))
	b.SetPhotoThumbnail(s.WidgetIntValue(wPhoto))
	b.SetManagerId(SessionUser(s).Id())
	ub, err := CreateAnimal(b)
	CheckOk(err)

	pr("created animal:", INDENT, ub)

	Todo("discard state, i.e. the edited fields; use a common prefix to simplify")
	Todo("Do a 'back' operation to go back to the previous page")
	NewManagerPage(s, p.parentPage).Generate()
	return nil
}

func animalInfoListener(s Session, widget Widget, n string, minLength int, maxLength int, emptyOk bool) (string, error) {
	errStr := ""

	if Alert("?Allowing zero characters in summary, details fields") {
		minLength = 0
	}
	for {
		ln := len(n)

		errStr = "Please add more info here."
		if ln < minLength && !(ln == 0 && emptyOk) {
			break
		}

		errStr = "Please type no more than " + IntToString(maxLength) + " characters."
		if ln > maxLength {
			break
		}

		errStr = ""
		break
	}
	var err error
	if errStr != "" {
		err = Error(errStr)
	}
	return n, err
}

func preSubmitValidateText(s Session, widget Widget, minLength int, maxLength int, flags ValidateFlag) {
	n := s.WidgetStrValue(widget)
	n, err := animalInfoListener(s, widget, n, minLength, maxLength, flags.Has(VALIDATE_EMPTYOK))
	s.SetWidgetProblem(widget, err)
}

func ValidateAnimalPhoto(s Session, valueWidget Widget, reportWidget Widget) {
	n := s.WidgetIntValue(valueWidget)
	if n == 0 {
		s.SetWidgetProblem(reportWidget, "Please upload a photo")
	}
}

func (p CreateAnimalPage) uploadPhotoListener(s Session, widget FileUpload, by []byte) error {
	pr := PrIf(false)

	m := s.WidgetManager()

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

	err = UpdateErrorWithString(err, problem)

	if err == nil {
		// Discard the old blob whose id we are now replacing
		DiscardBlob(s.SessionIntValue(id_animal_display_pic))

		// Store the id of the blob in the image widget
		s.State.Put(id_animal_display_pic, imageId)

		pr("repainting animal_display_pic")
		m.RepaintIds(id_animal_display_pic)

		pr("state:", s.State)
	}
	return err
}

func (p CreateAnimalPage) provideURL() string {
	pr := PrIf(false)
	url := ""
	s := p.session
	imageId := s.SessionIntValue(id_animal_display_pic)

	if imageId == 0 {
		imageId = 1 // This is the default placeholder blob id
	}
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
