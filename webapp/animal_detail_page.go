package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const anim_state_prefix = "_create_animal_:"

// // Use the field names that Animal produces as JSMaps
const (
	id_animal_uploadpic   = "photo"
	id_animal_display_pic = anim_state_prefix + "photo_thumbnail"
)

type AnimalDetailPageStruct struct {
	animalId int
	editing  bool
	name     string

	editor DataEditor

	nameWidget      InputWidget
	summaryWidget   InputWidget
	detailsWidget   InputWidget
	uploadPicWidget Widget
	imgWidget       Widget
}

type AnimalDetailPage = *AnimalDetailPageStruct

var CreateAnimalPageTemplate = &AnimalDetailPageStruct{name: "new"}
var EditAnimalPageTemplate = &AnimalDetailPageStruct{name: "edit"}
var ViewAnimalPageTemplate = &AnimalDetailPageStruct{name: "view"}

func (p AnimalDetailPage) ConstructPage(s Session, args PageArgs) Page {
	pr := PrIf("AnimDetailPage.ConstructPage", true)

	// Construct a copy of the template
	t := *p
	var result Page

	switch t.name {
	case "new":
		if args.Done() {
			t.editing = true
		}
		result = &t
		break
	case "view", "edit":
		t.animalId = args.PositiveInt()
		if args.Problem() {
			break
		}
		anim := ReadAnimalIgnoreError(t.animalId)
		if anim.Id() == 0 {
			break
		}
		user := SessionUser(s)
		if user.UserClass() == UserClassDonor {
			if p.name == "edit" {
				break
			}
			t.editing = false
		} else {
			if anim.ManagerId() != user.Id() {
				break
			}
			t.editing = true
		}
		t.prepareAnimal()
		result = &t
		break
	}

	if result != nil {
		pr("constructed page:", result.Name())
		pr("generating widgets")
		t.generateWidgets(s)
		pr("done generating")
	}
	return result
}

func (p AnimalDetailPage) Name() string {
	return p.name
}

func (p AnimalDetailPage) Args() []string {
	if p.animalId != 0 {
		return []string{IntToString(p.animalId)}
	}
	return nil
}

func (p AnimalDetailPage) prepareAnimal() {
	anim, err := ReadAnimal(p.animalId)
	if ReportIfError(err, "NewEditAnimalPage") {
		BadState(err)
	}
	p.editor = NewDataEditor(anim)
}

func (p AnimalDetailPage) viewing() bool {
	return p.name == "view"
}

func (p AnimalDetailPage) readStateFromAnimal(sess Session) {
	a := DefaultAnimal
	if p.animalId != 0 {
		var err error
		a, err = ReadActualAnimal(p.animalId)
		if ReportIfError(err, "AnimalDetailPage readStateFromAnimal") {
			return
		}
	}
	p.editor = NewDataEditor(a)
	//s.Put(id_animal_display_pic, a.PhotoThumbnail())
}

func (p AnimalDetailPage) generateWidgets(s Session) {
	if s == nil {
		return
	}
	s.DeleteStateFieldsWithPrefix(anim_state_prefix)
	GenerateHeader(s, p)
	if p.viewing() {
		AddUserHeaderWidget(s)
	}

	p.readStateFromAnimal(s)

	Todo("!Have ajax listener that can show advice without an actual error, e.g., if user left some fields blank")

	s.WidgetManager().PushStateProvider(p.editor.WidgetStateProvider)
	if p.editing {
		p.generateForEditing(s)
	} else {
		p.generateForViewing(s)
	}
	s.WidgetManager().PopStateProvider()
}

func (p AnimalDetailPage) generateForEditing(s Session) {
	m := s.WidgetManager()
	m.Col(6).Open()
	{
		m.Col(12)

		Todo("!datagen option to generate field name constants for use here")

		p.nameWidget = m.Label("Name").Id(Animal_Name).AddInput(AnimalNameListener)

		p.summaryWidget = m.Label("Summary").Id(Animal_Summary).AddInput(p.animalSummaryListener)
		m.Size(SizeTiny).Label("A brief paragraph to appear in the 'card' view.").AddText()

		p.detailsWidget = m.Label("Details").Id(Animal_Details).AddInput(p.animalDetailsListener)
		m.Size(SizeTiny).Label("Additional paragraphs to appear on the 'details' view.").AddText()

		/**
			name            string
		summary         string
		details         string
		campaignTarget  int
		campaignBalance int
		photoThumbnail  int
		*/

		m.Col(6)
		if p.animalId != 0 {
			m.Label("Done").AddButton(p.doneEditListener)
			m.Label("Abort").AddButton(p.abortEditListener)
		} else {
			m.Label("Create").AddButton(p.createAnimalButtonListener)
			m.Label("Abort").AddButton(p.abortEditListener)
		}
	}
	m.Close()

	m.Open()
	//m.Id(id_animal_uploadpic)
	p.uploadPicWidget =
		m.Label("Photo").Id(id_animal_uploadpic).AddFileUpload(p.uploadPhotoListener)

	imgWidget := m.Id(id_animal_display_pic).AddImage()
	imgWidget.URLProvider = p.provideURL
	// Scale the photos based on browser resolution
	imgWidget.SetSize(AnimalPicSizeNormal, 0.6)
	p.imgWidget = imgWidget

	m.Close()
}

func (p AnimalDetailPage) generateForViewing(s Session) {
	m := s.WidgetManager()

	m.Col(6).Open()
	{
		Todo("!Flesh this out some")
		m.Col(12)
		m.Id("name").AddText()

		//m.Label("Summary").Id(id_animal_summary).AddInput(p.AnimalTextListener)
		//m.Size(SizeTiny).Label("A brief paragraph to appear in the 'card' view.").AddText()
		//m.Label("Details").Id(id_animal_details).AddInput(p.AnimalTextListener)
		//m.Size(SizeTiny).Label("Additional paragraphs to appear on the 'details' view.").AddText()
		//
		m.Col(6)
		m.Label("Done").AddButton(p.doneViewListener)

	}
	m.Close()

	m.Open()
	imgWidget := m.AddImage()
	imgWidget.URLProvider = p.provideURL

	// Scale the photos based on browser resolution
	imgWidget.SetSize(AnimalPicSizeNormal, 0.6)

	p.imgWidget = imgWidget

	m.Close()
}

func AnimalNameListener(s Session, widget InputWidget, value string) (string, error) {
	return ValidateAnimalName(value, VALIDATE_EMPTYOK)
}

func (p AnimalDetailPage) animalSummaryListener(sess Session, widget InputWidget, value string) (string, error) {
	return animalInfoListener(value, 20, 200, true)
}

func (p AnimalDetailPage) animalDetailsListener(sess Session, widget InputWidget, value string) (string, error) {
	return animalInfoListener(value, 200, 2000, true)
}

func (p AnimalDetailPage) createAnimalButtonListener(s Session, widget Widget, arg string) {
	pr := PrIf("", true)

	if !p.validateAll(s) {
		return
	}

	b := NewAnimal()
	p.writeStateToAnimal(s, b)

	ub, err := CreateAnimal(b)
	if ReportIfError(err, "CreateAnimal after editing") {
		return
	}

	pr("created animal:", INDENT, ub)
	p.exit(s)
}

func (p AnimalDetailPage) validateAll(s Session) bool {
	pr := PrIf("", false)

	{
		text := s.WidgetStringValue(p.nameWidget)
		_, err := ValidateAnimalName(text, 0)
		s.SetProblem(p.nameWidget, err)
	}

	preCreateValidateText(s, p.summaryWidget, 20, 200, 0)
	preCreateValidateText(s, p.detailsWidget, 200, 2000, 0)
	{
		picId :=
			s.WidgetIntValue(p.imgWidget)
		//s.IntValue(id_animal_display_pic)
		if picId == 0 {
			s.SetProblem(p.uploadPicWidget, "Please upload a photo")
		}
	}

	errcount := WidgetErrorCount(s.PageWidget, s.State)
	pr("error count:", errcount)
	return errcount == 0
}

func (p AnimalDetailPage) doneEditListener(s Session, widget Widget, arg string) {
	pr := PrIf("", false)

	if !p.validateAll(s) {
		return
	}

	a, err := ReadActualAnimal(p.animalId)
	if ReportIfError(err, "ReadAnimal after editing") {
		return
	}
	b := a.ToBuilder()
	p.writeStateToAnimal(s, b)

	err = UpdateAnimal(b)
	if ReportIfError(err, "UpdateAnimal after editing") {
		return
	}
	pr("updated animal", b.ToJson().AsJSMap().CompactString())
	p.exit(s)
}

func (p AnimalDetailPage) doneViewListener(s Session, widget Widget, arg string) {
	p.exit(s)
}

func (p AnimalDetailPage) abortEditListener(s Session, widget Widget, arg string) {
	p.exit(s)
}

func (p AnimalDetailPage) writeStateToAnimal(s Session, b AnimalBuilder) {
	pr := PrIf("writeStateToAnimal", true)
	pr("builder, before:", INDENT, b)
	b.SetName(p.nameWidget.Value(s))
	b.SetSummary(p.summaryWidget.Value(s))
	b.SetDetails(p.detailsWidget.Value(s))
	b.SetPhotoThumbnail(s.IntValue(id_animal_display_pic))
	b.SetManagerId(SessionUser(s).Id())
	pr("builder, after:", INDENT, b)
}

func (p AnimalDetailPage) exit(s Session) {
	pr := PrIf("AnimalDetailPage.exit", true)
	pr("state fields:", INDENT, s.State)
	s.DeleteStateFieldsWithPrefix(anim_state_prefix)
	pr("state fields after deleting:", INDENT, s.State)
	if SessionUser(s).UserClass() == UserClassDonor {
		s.SwitchToPage(FeedPageTemplate, nil)
	} else {
		s.SwitchToPage(ManagerPageTemplate, nil)
	}
	pr("state fields after switching pages:", INDENT, s.State)
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

func preCreateValidateText(s Session, widget Widget, minLength int, maxLength int, flags ValidateFlag) {
	n := s.WidgetStringValue(widget)
	n, err := animalInfoListener(n, minLength, maxLength, flags.Has(VALIDATE_EMPTYOK))
	s.SetProblem(widget, err)
}

func (p AnimalDetailPage) uploadPhotoListener(s Session, widget FileUpload, by []byte) error {
	pr := PrIf("", false)

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
		DiscardBlob(s.IntValue(picId))

		// Store the id of the blob in the image widget
		s.State.Put(picId, imageId)

		pr("repainting animal_display_pic")
		s.RepaintIds(picId)

		pr("state:", s.State)
	}
	return err
}

func (p AnimalDetailPage) provideURL(s Session) string {
	pr := PrIf("", false)
	url := ""

	// We need to access the state directly, without a widget.
	imageId := s.IntValue(id_animal_display_pic)

	if imageId == 0 {
		imageId = 1 // This is the default placeholder blob id
	}
	pr("provideURL, image id read from state:", imageId)

	if imageId != 0 {
		url = SharedWebCache.GetBlobURL(imageId)
		pr("read into cache, url:", url)
	}
	return url
}

func DiscardBlob(id int) {
	if id != 0 {
		Todo("#50Discard blob id", id)
	}
}
