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

type GalleryPageStruct struct {
	fooMap JSMap
	list   ListWidget
	editor DataEditor
}

func NewGalleryPage(sess Session) Page {
	t := &GalleryPageStruct{
		fooMap: NewJSMap().Put("bar", "hello"),
	}
	if sess != nil {
		t.generateWidgets(sess)
	}
	return t
}

const GalleryPageName = "gallery"

var GalleryPageTemplate = NewGalleryPage(nil)

func (p GalleryPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		return NewGalleryPage(s)
	}
	return nil
}

func (p GalleryPage) Name() string {
	return GalleryPageName
}

func (p GalleryPage) Args() []string { return nil }

// ------------------------------------------------------------------------------------

var alertWidget AlertWidget
var myRand = NewJSRand().SetSeed(1234)

func (p GalleryPage) generateWidgets(sess Session) {

	trim := false && Alert("removing most widgets")

	m := GenerateHeader(sess, p)

	anim := ReadAnimalIgnoreError(3)
	if anim.Id() == 0 {
		Alert("No animals available")
	} else {
		m.Open()
		p.editor = NewDataEditor(anim)
		m.Id("foo").Label("Photo").AddFileUpload(p.uploadListener)

		// The image widget will have the same id as the animal field that is to store its photo blob id.
		//
		imgWidget := m.Id(Animal_PhotoThumbnail).AddImage()
		imgWidget.SetSize(AnimalPicSizeNormal, 0.3)

		Todo("!image widgets should have a state that is some sort of string, eg a blob name, or str(blob id); separately a URLProvider which may take the state as an arg")

		Todo("!give widgets values (state) in this way wherever appropriate")

		imgWidget.URLProvider = func(s Session) string {
			// We are storing the image's blob id in the animal's photo_thumbnail field,
			// so we can use the editor's embedded JSMap to access it.
			imageId := p.editor.GetInt(Animal_PhotoThumbnail)
			Pr("provideURL, image id read from state:", imageId)
			url := ""
			if imageId != 0 {
				url = SharedWebCache.GetBlobURL(imageId)
				Pr("read into cache, url:", url)
			}
			return url
		}
		m.Close()
	}

	m.Open()
	{
		m.Col(6)

		x := m.Label("hello").AddText()
		//x.SetTrace(true)
		x.SetVisible(false)
		m.Label("Toggle Visibility").AddButton(func(s Session, w Widget, arg string) {
			newState := !x.Visible()
			Pr("setting x visible:", newState)
			x.SetVisible(newState)
			s.Repaint(x)
		})
	}
	m.Close()

	m.Open()
	{
		m.Label("In HTML and CSS, background color is denoted by " +
			"the background-color property. To add or " +
			"change background color in HTML,").Size(SizeSmall).Height(5).AddText()

	}
	m.Close()

	if !trim {
		listItemWidget := m.Open()
		m.Id("foo_text").Height(3).AddText()
		m.Close()

		p.list = m.AddList(NewGalleryListImplementation(), listItemWidget)
		if trim {
			p.list.WithPageControls = false
		}
		Todo("!Add support for empty list items, to pad out page to full size")
	}

	if !trim {
		m.Open()

		m.Id("fred").Label(`Fred`).AddButton(buttonListener)

		{
			m.Col(4)

			cardListener := func(sess Session, widget AnimalCard, arg string) {
				Pr("card listener, animal id:", widget.Animal().Id(), "arg:", arg)
			}
			cardButtonListener := func(sess Session, widget AnimalCard, arg string) {
				Pr("card button listener, name:", widget.Animal().Name(), "arg:", arg)
			}

			Todo("We need to create a state provider for cards, when not in list (list handles that already somehow)")

			// Create a new card that will contain other widgets
			c1 := NewAnimalCard(m, ReadAnimalIgnoreError(3), cardListener, "Hello", cardButtonListener)
			//c1.SetTrace(true)

			m.Add(c1)
			m.Add(
				NewAnimalCard(m, ReadAnimalIgnoreError(4), nil, "Bop", cardButtonListener))

			m.Open()

			m.PushStateProvider(NewStateProvider("", p.fooMap))
			m.PushIdPrefix("")
			{

				m.Col(4)
				m.Label("Static text.").Height(5).AddText()
				m.Id("bar").Label("Bar:").AddInput(p.fooListener)
			}
			m.PopIdPrefix()
			m.PopStateProvider()
			m.Close()

		}

		m.Close()
	}

	// Open a container for all these various columns so we restore the default when it closes
	m.Open()
	{
		m.Col(4)
		m.Label("uniform delta").AddText()
		m.Col(8)
		m.Id("x58").Label(`Disabled`).AddButton(buttonListener).SetEnabled(false)

		m.Col(2).AddSpace()
		m.Col(3).Id("yz").Label(`Enabled`).AddButton(buttonListener)

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
	m.Close()

	if !trim {
		AddUserHeaderWidget(sess)
		alertWidget = NewAlertWidget("sample_alert", AlertInfo)
		alertWidget.SetVisible(false)
		m.Add(alertWidget)
	}

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

	s.State.Put(alertWidget.Id(),
		strings.TrimSpace(newVal+" "+
			RandomText(myRand, 55, false)))
	s.Repaint(alertWidget)
	return newVal, nil
}

func buttonListener(s Session, widget Widget, arg string) {
	Pr("buttonListener, widget:", widget.Id(), "arg:", arg)
	wid := widget.Id()
	newVal := "Clicked: " + wid

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal
	alertWidget.SetVisible(true)

	if Alert("Is this adequate?") {
		s.SetWidgetValue(alertWidget, newVal)
	} else {
		s.State.Put(alertWidget.Id(),
			strings.TrimSpace(newVal))
		s.Repaint(alertWidget)
	}
}

func (p GalleryPage) checkboxListener(s Session, widget CheckboxWidget, state bool) (bool, error) {
	Pr("gallery, id", widget.Id(), "new state:", state)
	return state, nil
}

func (p GalleryPage) uploadListener(s Session, fileUploadWidget FileUpload, value []byte) error {
	pr := PrIf("Gallery.uploadListener", true)

	//pr("who called this?", Callers(0, 8))
	Todo("!fileUploadWidget argument not used")

	Alert("For simplicity, maybe file upload widgets don't have values.  They just return byte arrays, and what are done with them is up to the client.")

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
		p.editor.Put(Animal_PhotoThumbnail, imageId)
		s.RepaintIds(Animal_PhotoThumbnail)
	}
	return errOut
}

func (p GalleryPage) fooListener(sess Session, widget InputWidget, value string) (string, error) {
	Todo("Clarify prefix role in provider, widget ids, and resolve confusion about add/subtract prefix")
	Pr("fooListener, id:", widget.Id(), "value:", value, CR, "current map:", INDENT, p.fooMap)
	return value, nil
}

// ------------------------------------------------------------------------------------
// List
// ------------------------------------------------------------------------------------

type GalleryListImplementationStruct struct {
	BasicListStruct
	names []string
}

type GalleryListImplementation = *GalleryListImplementationStruct

type GalleryPage = *GalleryPageStruct

func NewGalleryListImplementation() GalleryListImplementation {
	t := &GalleryListImplementationStruct{}
	t.ElementsPerPage = 3
	j := NewJSRand().SetSeed(1965)
	for i := 0; i < 50; i++ {
		t.names = append(t.names, RandomText(j, 12, false))
		t.ElementIds = append(t.ElementIds, i)
	}
	return t
}

func (g GalleryListImplementation) listItemRenderer(session Session, widget ListWidget, elementId int, m MarkupBuilder) {
	m.TgOpen(`div class="col-sm-4"`).TgContent()
	m.A(ESCAPED, ToString("#", elementId, g.names[elementId]))
	m.TgClose()
}

func (g GalleryListImplementation) ItemStateProvider(s Session, elementId int) WidgetStateProvider {
	json := NewJSMap()
	json.Put("foo_text", ToString("Item #", elementId, g.names[elementId]))
	return NewStateProvider("", json)
}
