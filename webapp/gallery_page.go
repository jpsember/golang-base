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

func (p GalleryPage) Args() []string { return EmptyStringSlice }

// ------------------------------------------------------------------------------------

const sampleImageId = "sample_image"

var alertWidget AlertWidget
var myRand = NewJSRand().SetSeed(1234)

const gallery_card_prefix = "gallery_card."

func (p GalleryPage) generateWidgets(sess Session) {

	trim := false && Alert("removing most widgets")

	m := GenerateHeader(sess, p)

	m.Open()
	{
		m.Col(6)

		x := m.Label("hello").AddText()
		//x.SetTrace(true)
		x.SetVisible(false)
		m.Label("Toggle Visibility").AddButton(func(s Session, w Widget) {
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
		x := NewGalleryListImplementation()

		if true && Alert("constructing list with default renderer") {
			p.list = m.AddList(x, nil, nil)

		} else {
			listItemWidget := m.Open()
			m.Id("foo_text").Height(3).AddText()
			m.Close()
			m.Detach(listItemWidget)

			listProvider := func(sess Session, widget *ListWidgetStruct, itemId int) WidgetStateProvider {
				json := NewJSMap()
				json.Put("foo_text", ToString("Item #", itemId, x.names[itemId]))
				return NewStateProvider("", json)
			}

			p.list = m.AddList(x, listItemWidget, listProvider)
		}
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

			cardListener := func(sess Session, widget AnimalCard) { Pr("card listener, animal id:", widget.Animal().Id()) }
			cardButtonListener := func(sess Session, widget AnimalCard) { Pr("card button listener, name:", widget.Animal().Name()) }

			Todo("We need to create a state provider for cards, when not in list (list handles that already somehow)")

			// Create a new card that will contain other widgets
			c1 := NewAnimalCard("gallery_card", ReadAnimalIgnoreError(3), cardListener, "Hello", cardButtonListener)
			//c1.SetTrace(true)

			m.Add(c1)
			m.Add(
				NewAnimalCard("gallery_card2", ReadAnimalIgnoreError(4), nil, "Bop", cardButtonListener))

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

		m.Id("sample_upload").Label("Photo").AddFileUpload(p.uploadListener)
		imgWidget := m.Id("sample_image").AddImage()
		Todo("!image widgets should have a state that is some sort of string, eg a blob name, or str(blob id); separately a URLProvider which may take the state as an arg")
		Todo("!give widgets values (state) in this way wherever appropriate")
		imgWidget.URLProvider = p.provideURL
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

	s.State.Put(alertWidget.BaseId,
		strings.TrimSpace(newVal+" "+
			RandomText(myRand, 55, false)))
	s.Repaint(alertWidget)
	return newVal, nil
}

func buttonListener(s Session, widget Widget) {
	Pr("buttonListener, widget:", widget.Id())
	wid := widget.Id()
	newVal := "Clicked: " + wid

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal
	alertWidget.SetVisible(true)

	s.State.Put(alertWidget.BaseId,
		strings.TrimSpace(newVal))
	s.Repaint(alertWidget)
}

func (p GalleryPage) checkboxListener(s Session, widget CheckboxWidget, state bool) (bool, error) {
	Pr("gallery, id", widget.Id(), "new state:", state)
	return state, nil
}

func (p GalleryPage) uploadListener(s Session, fileUploadWidget FileUpload, value []byte) error {
	pr := PrIf(false)

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
		s.RepaintIds(sampleImageId)
	}
	return errOut
}

func (p GalleryPage) provideURL(s Session) string {
	pr := PrIf(false)
	url := ""
	imageId := s.State.OptInt(sampleImageId, 0)

	pr("provideURL, image id read from state:", imageId)

	if imageId != 0 {
		url = SharedWebCache.GetBlobURL(imageId)
		pr("read into cache, url:", url)
	}
	return url
}

func (p GalleryPage) clickListener(sess Session, message string) bool {
	Todo("This explicit handler probably not required")
	//
	//if p.list.HandleClick(sess, message) {
	//	return true
	//}

	if arg, f := TrimIfPrefix(message, gallery_card_prefix); f {
		Pr("card click, remaining arg:", arg)
		return true

	}
	return false
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
