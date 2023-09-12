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
	id_animal_uploadpic   = anim_state_prefix + "photo"
	id_animal_display_pic = anim_state_prefix + "pic"
)

type CreateAnimalPageStruct struct {
	BasicPage
	editId int
}

type CreateAnimalPage = *CreateAnimalPageStruct

func NewCreateAnimalPage(sess Session, parentWidget Widget) AbstractPage {
	t := &CreateAnimalPageStruct{
		BasicPage: NewBasicPage(sess, parentWidget),
	}
	t.devLabel = "create_animal_page"
	return t
}

func NewEditAnimalPage(sess Session, parentWidget Widget, animalId int) AbstractPage {
	t := &CreateAnimalPageStruct{
		BasicPage: NewBasicPage(sess, parentWidget),
		editId:    animalId,
	}
	t.devLabel = "edit_animal_page"
	return t
}

func (p CreateAnimalPage) readStateFromAnimal() {
	a := DefaultAnimal
	if p.editId != 0 {
		var err error
		a, err = ReadAnimal(p.editId)
		if err != nil || a.Id() == 0 {
			Alert("Trouble reading animal with id", p.editId, "; err:", err)
		}
	}
	s := p.session.State
	s.Put(id_animal_name, a.Name())
	s.Put(id_animal_summary, a.Summary())
	s.Put(id_animal_details, a.Details())
	s.Put(id_animal_display_pic, a.PhotoThumbnail())
}

func (p CreateAnimalPage) Generate() {
	//SetWidgetDebugRendering()
	p.session.SetClickListener(nil)
	p.session.DeleteStateFieldsWithPrefix(anim_state_prefix)
	m := p.GenerateHeader()

	p.readStateFromAnimal()

	Todo("!Have ajax listener that can show advice without an actual error, e.g., if user left some fields blank")

	//m.Label("Create New Animal Record").Size(SizeLarge).AddHeading()

	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("Name").Id(id_animal_name).AddInput(AnimalNameListener)

		m.Label("Summary").Id(id_animal_summary).AddInput(p.AnimalTextListener)
		m.Size(SizeTiny).Label("A brief paragraph to appear in the 'card' view.").AddText()
		m.Label("Details").Id(id_animal_details).AddInput(p.AnimalTextListener)
		m.Size(SizeTiny).Label("Additional paragraphs to appear on the 'details' view.").AddText()

		if p.editId != 0 {
			m.Label("Done").AddButton(p.doneEditListener)
		} else {
			m.Label("Create").AddButton(p.createAnimalButtonListener)
		}
	}
	m.Close()

	m.Open()
	m.Id(id_animal_uploadpic).Label("Photo").AddFileUpload(p.uploadPhotoListener)
	imgWidget := m.Id(id_animal_display_pic).AddImage()
	imgWidget.URLProvider = p.provideURL

	// Scale the photos based on browser resolution
	imgWidget.SetSize(AnimalPicSizeNormal, 0.6)

	m.Close()
}

func AnimalNameListener(s Session, widget InputWidget, value string) (string, error) {
	return ValidateAnimalName(value, VALIDATE_EMPTYOK)
}

func (p CreateAnimalPage) AnimalTextListener(sess Session, widget InputWidget, value string) (string, error) {
	if widget.Id() == id_animal_summary {
		return animalInfoListener(value, 20, 200, true)
	} else {
		return animalInfoListener(value, 200, 2000, true)
	}
}

func (p CreateAnimalPage) createAnimalButtonListener(s Session, widget Widget) {
	pr := PrIf(true)

	if !p.validateAll() {
		return
	}
	//{
	//	text := s.WidgetStrValue(id_animal_name)
	//	_, err := ValidateAnimalName(text, 0)
	//	s.SetWidgetProblem(id_animal_name, err)
	//}
	//
	//preCreateValidateText(s, id_animal_summary, 20, 200, 0)
	//preCreateValidateText(s, id_animal_details, 200, 2000, 0)
	//{
	//	picId := s.WidgetIntValue(id_animal_display_pic)
	//	if picId == 0 {
	//		s.SetWidgetProblem(id_animal_uploadpic, "Please upload a photo")
	//	}
	//}
	//
	//errcount := WidgetErrorCount(p.parentPage, s.State)
	//pr("error count:", errcount)
	//if errcount != 0 {
	//	return nil
	//}

	b := NewAnimal()
	p.writeStateToAnimal(b)
	//b.SetName(strings.TrimSpace(s.WidgetStrValue(id_animal_name)))
	//b.SetSummary(strings.TrimSpace(s.WidgetStrValue(id_animal_summary)))
	//b.SetDetails(strings.TrimSpace(s.WidgetStrValue(id_animal_details)))
	//b.SetPhotoThumbnail(s.WidgetIntValue(id_animal_display_pic))
	//b.SetManagerId(SessionUser(s).Id())
	ub, err := CreateAnimal(b)
	if ReportIfError(err, "CreateAnimal after editing") {
		return
	}

	pr("created animal:", INDENT, ub)
	p.exit()
	//s.DeleteStateFieldsWithPrefix(anim_state_prefix)
	//
	//Todo("Discard any existing manager animal list, as its contents have now changed")
	//s.DeleteSessionData(SessionKey_MgrList)
	//
	//Todo("Do a 'back' operation to go back to the previous page")
	//NewManagerPage(s, p.parentPage).Generate()
}

func (p CreateAnimalPage) validateAll() bool {
	pr := PrIf(true)
	s := p.session

	{
		text := s.WidgetStrValue(id_animal_name)
		_, err := ValidateAnimalName(text, 0)
		s.SetWidgetProblem(id_animal_name, err)
	}

	preCreateValidateText(s, id_animal_summary, 20, 200, 0)
	preCreateValidateText(s, id_animal_details, 200, 2000, 0)
	{
		picId := s.WidgetIntValue(id_animal_display_pic)
		if picId == 0 {
			s.SetWidgetProblem(id_animal_uploadpic, "Please upload a photo")
		}
	}

	errcount := WidgetErrorCount(p.parentPage, s.State)
	pr("error count:", errcount)
	return errcount == 0
}

func (p CreateAnimalPage) doneEditListener(s Session, widget Widget) {
	pr := PrIf(true)

	if !p.validateAll() {
		return
	}
	//{
	//	text := s.WidgetStrValue(id_animal_name)
	//	_, err := ValidateAnimalName(text, 0)
	//	s.SetWidgetProblem(id_animal_name, err)
	//}
	//
	//preCreateValidateText(s, id_animal_summary, 20, 200, 0)
	//preCreateValidateText(s, id_animal_details, 200, 2000, 0)
	//{
	//	picId := s.WidgetIntValue(id_animal_display_pic)
	//	if picId == 0 {
	//		s.SetWidgetProblem(id_animal_uploadpic, "Please upload a photo")
	//	}
	//}
	//
	//errcount := WidgetErrorCount(p.parentPage, s.State)
	//pr("error count:", errcount)
	//if errcount != 0 {
	//	return nil
	//}

	a, err := ReadAnimal(p.editId)
	if ReportIfError(err, "ReadAnimal after editing") {
		return
	}
	b := a.ToBuilder()
	p.writeStateToAnimal(b)

	err = UpdateAnimal(b)
	if ReportIfError(err, "UpdateAnimal after editing") {
		return
	}
	pr("updated animal", b)
	p.exit()
}

func (p CreateAnimalPage) writeStateToAnimal(b AnimalBuilder) {
	s := p.session
	b.SetName(strings.TrimSpace(s.WidgetStrValue(id_animal_name)))
	b.SetSummary(strings.TrimSpace(s.WidgetStrValue(id_animal_summary)))
	b.SetDetails(strings.TrimSpace(s.WidgetStrValue(id_animal_details)))
	b.SetPhotoThumbnail(s.WidgetIntValue(id_animal_display_pic))
	b.SetManagerId(SessionUser(s).Id())
}

func (p CreateAnimalPage) exit() {
	s := p.session

	s.DeleteStateFieldsWithPrefix(anim_state_prefix)

	Todo("Discard any existing manager animal list, as its contents have now changed")
	s.DeleteSessionData(SessionKey_MgrList)

	Todo("Do a 'back' operation to go back to the previous page")
	NewManagerPage(s, p.parentPage).Generate()
}

func animalInfoListener(n string, minLength int, maxLength int, emptyOk bool) (string, error) {
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

func preCreateValidateText(s Session, widgetId string, minLength int, maxLength int, flags ValidateFlag) {
	n := s.WidgetStrValue(widgetId)
	n, err := animalInfoListener(n, minLength, maxLength, flags.Has(VALIDATE_EMPTYOK))
	s.SetWidgetProblem(widgetId, err)
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

	// The listener widget id is the *upload* widget.  We want to store the
	// appropriately transformed image in the *display_pic* widget (and save
	// that version to the database).
	//
	if err == nil {
		picId := id_animal_display_pic
		// Discard the old blob whose id we are now replacing
		DiscardBlob(s.WidgetIntValue(picId))

		// Store the id of the blob in the image widget
		s.State.Put(picId, imageId)

		pr("repainting animal_display_pic")
		m.RepaintIds(picId)

		pr("state:", s.State)
	}
	return err
}

func (p CreateAnimalPage) provideURL() string {
	pr := PrIf(false)
	url := ""
	s := p.session
	imageId := s.WidgetIntValue(id_animal_display_pic)

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

func ReportIfError(err error, msg ...any) bool {
	if err != nil {
		Alert("#50<1Error occurred, ignoring!  Error:", err, INDENT, "Message:", ToString(msg...))
		return true
	}
	return false
}
