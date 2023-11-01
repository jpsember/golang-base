package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

const (
	GDistinctDataObjects = false
	GList                = true
	GListMultiItems      = true
	GListPager           = false
	GAlert               = false
	GClickPic            = false
	GUploadPic           = false
	GVisibility          = false
	GTextArea            = false
	GColumns             = false
	GUserHeader          = false
)

type GalleryPageStruct struct {
	alertWidget AlertWidget
	list        ListWidget
	picEditor   DataEditor
	ourState    JSMap
	editorA     DataEditor
	editorB     DataEditor
	rand        JSRand
	clickpic    int
}

var GalleryPageTemplate = &GalleryPageStruct{}

func (p GalleryPage) ConstructPage(s Session, args PageArgs) Page {
	if !args.CheckDone() {
		return nil
	}
	x := &GalleryPageStruct{
		rand:     NewJSRand().SetSeed(1234),
		ourState: NewJSMap(),
	}
	x.generateWidgets(s)
	return x
}

func (p GalleryPage) Name() string {
	return "gallery"
}

func (p GalleryPage) Args() []string { return nil }

// ------------------------------------------------------------------------------------

func (p GalleryPage) generateWidgets(sess Session) {
	pr := PrIf("GalleryPage.generateWidgets", false)
	pr("generateWidgets")

	m := GenerateHeader(sess, p)

	// ------------------------------------------------------------------------------------
	// The list
	// ------------------------------------------------------------------------------------

	if GList {

		Alert("The list item has no listener; how does the element id get propagated to the list's listener?")
		listItemWidget := m.Open()
		// We want all the list item widgets to get their state from the list itself;
		// so we haven't pushed a state map yet
		m.Id("foo_text").Height(3).AddText()
		m.Close()

		glist := NewGalleryListImplementation()
		ourListListener := func(sess Session, widget *ListWidgetStruct, elementId int, args WidgetArgs) error {
			Pr("GList event, element:", elementId, "args:", args, "element state:", glist.ItemStateMap(sess, elementId))
			return nil
		}

		p.list = m.Id("pets").AddList(glist, listItemWidget, ourListListener)
		p.list.WithPageControls = GListPager
		Todo("!Add support for empty list items, to pad out page to full size")
	}

	// ------------------------------------------------------------------------------------
	// Two widget sets displaying a couple of data objects, each set with a unique prefix
	// ------------------------------------------------------------------------------------

	if GDistinctDataObjects {
		m.Open()
		p.editorA = NewDataEditorWithPrefix(NewAnimal().SetName("Andy"), "a")
		p.editorB = NewDataEditorWithPrefix(NewAnimal().SetName("Brian"), "b")

		nameListener := func(sess Session, widget InputWidget, value string) (string, error) {
			Pr("GDDistinctDataObjects listener, id:", QUO, widget.Id(), "new value:", QUO, value, "; current names:", INDENT, //
				p.editorA.GetString("name"), CR, p.editorB.GetString("name"))
			return value, nil
		}

		sess.PushEditor(p.editorA)
		m.Label("Name A").Id(Animal_Name).AddInput(nameListener)
		sess.PopEditor()

		sess.PushEditor(p.editorB)
		b := m.Label("Name B").Id(Animal_Name).AddInput(nameListener)
		sess.PopEditor()

		m.Listener(
			func(s Session, w Widget, args WidgetArgs) {
				s.SetWidgetValue(b, RandomText(p.rand, 3, false))
				b.Repaint()
			}).Label("Repaint B").AddBtn()
		m.Close()
	}

	// ------------------------------------------------------------------------------------
	// A clickable image; clicking on it changes the image
	// ------------------------------------------------------------------------------------

	if GClickPic {
		m.Open()

		var imgUrlProvider = func(s Session) string {
			pr := PrIf("imgURLProvider", false)
			// We are storing the image's blob id in the animal's photo_thumbnail field,
			// so we can use the editor's embedded JSMap to access it.
			imageId := p.clickpic
			pr("provideURL, image id read from state:", imageId)
			url := ""
			if imageId != 0 {
				url = SharedWebCache.GetBlobURL(imageId)
				pr("read into cache, url:", url)
			}
			return url
		}

		var imgWidget ImageWidget

		m.Listener(func(s Session, w Widget, args WidgetArgs) {
			for {
				anim := ReadAnimalIgnoreError(p.rand.Intn(8) + 1)
				if anim == nil || p.clickpic == anim.PhotoThumbnail() {
					continue
				}
				p.clickpic = anim.PhotoThumbnail()
				break
			}
			imgWidget.Repaint()
		})

		imgWidget = m.Id("clickpic").AddImage(imgUrlProvider)
		imgWidget.SetSize(AnimalPicSizeNormal, 0.3)

		m.Close()
	}

	// ------------------------------------------------------------------------------------

	m.PushStateMap(p.ourState)

	if GAlert {
		p.alertWidget = NewAlertWidget("sample_alert", AlertInfo)
		p.alertWidget.SetVisible(false)
		m.Add(p.alertWidget)
	}

	if GUploadPic {
		anim := ReadAnimalIgnoreError(3)
		if anim.Id() == 0 {
			Alert("No animals available")
		} else {
			m.Open()
			p.picEditor = NewDataEditor(anim)

			m.Id("foo").Label("Photo").AddFileUpload(p.uploadListener)

			// The image widget will have the same id as the animal field that is to store its photo blob id.
			//

			var imgUrlProvider = func(s Session) string {
				pr := PrIf("imgURLProvider", false)
				// We are storing the image's blob id in the animal's photo_thumbnail field,
				// so we can use the editor's embedded JSMap to access it.
				imageId := p.picEditor.GetInt(Animal_PhotoThumbnail)
				pr("provideURL, image id read from state:", imageId)
				url := ""
				if imageId != 0 {
					url = SharedWebCache.GetBlobURL(imageId)
					pr("read into cache, url:", url)
				}
				return url
			}

			imgWidget := m.Id(Animal_PhotoThumbnail).AddImage(imgUrlProvider)
			imgWidget.SetSize(AnimalPicSizeNormal, 0.3)

			Todo("!image widgets should have a state that is some sort of string, eg a blob name, or str(blob id); separately a URLProvider which may take the state as an arg")

			Todo("!give widgets values (state) in this way wherever appropriate")

			m.Close()
		}
	}

	if GVisibility {
		m.Open()
		{
			m.Col(6)

			x := m.Label("hello").AddText()
			//x.SetTrace(true)
			x.SetVisible(false)
			m.Label("Toggle Visibility").Listener(func(s Session, w Widget, args WidgetArgs) {
				newState := !x.Visible()
				Pr("setting x visible:", newState)
				x.SetVisible(newState)
				x.Repaint()
			}).AddBtn()
		}
		m.Close()
	}

	if GTextArea {
		m.Open()
		{
			m.Label("In HTML and CSS, background color is denoted by " +
				"the background-color property. To add or " +
				"change background color in HTML,").Size(SizeSmall).Height(5).AddText()

		}
		m.Close()
	}

	if GColumns {

		buttonListener := func(s Session, widget Widget, args WidgetArgs) {
			Pr("buttonListener, widget:", widget.Id(), "args:", args)
			wid := widget.Id()
			newVal := "Clicked: " + wid

			if GAlert {
				w := p.alertWidget
				// Increment the alert class, and update its message
				w.Class = (w.Class + 1) % AlertTotal
				w.SetVisible(true)
				s.SetWidgetValue(w, newVal)
			}
		}

		// Open a container for all these various columns so we restore the default when it closes
		m.Open()
		{
			m.Col(4)
			m.Label("uniform delta").AddText()
			m.Col(8)
			m.Id("x58").Label(`Disabled`).Listener(buttonListener).AddBtn().SetEnabled(false)

			m.Col(2).AddSpace()
			m.Col(3).Id("yz").Label(`Enabled`).Listener(buttonListener).AddBtn()

			m.Col(3).AddSpace()
			m.Col(4).AddSpace()

			m.Col(6)
			m.Label("Bird").Id("bird")
			var birdListener = func(s Session, widget InputWidget, newVal string) (string, error) {
				var err error
				Todo("?can we have sessions produce listener functions with appropriate handling of sess any?")
				if newVal == "parrot" {
					err = Error("No parrots, please!")
				}
				return newVal, err
			}
			m.AddInput(birdListener)

			m.Col(6)
			m.Open()

			cbListener := func(s Session, widget CheckboxWidget, state bool) (bool, error) {
				Pr("gallery, id", widget.Id(), "new state:", state)
				return state, nil
			}
			m.Id("x59").Label(`Label for X59`).AddCheckbox(cbListener)
			m.Id("x60").Label(`With fruit`).AddSwitch(cbListener)
			m.Close()

			m.Col(4)
			m.Id("launch").Label(`Launch`).Listener(buttonListener).AddBtn()

			m.Col(8)
			m.Label(`Sample text; is 5 < 26? A line feed
"Quoted string"
Multiple line feeds:


   an indented final line`)
			m.AddText()

			m.Col(4)
			sess.PushStateMap(p.ourState)

			var zebraListener = func(s Session, widget InputWidget, newVal string) (string, error) {

				if GAlert {
					w := p.alertWidget
					// Increment the alert class, and update its message
					w.Class = (w.Class + 1) % AlertTotal
					w.SetVisible(newVal != "")

					s.SetWidgetValue(w,
						strings.TrimSpace(newVal+" "+
							RandomText(p.rand, 55, false)))
					w.Repaint()
				}
				return newVal, nil
			}

			m.Label("Animal").Id("zebra").AddInput(zebraListener)
			sess.PopStateMap()
		}
		m.Close()

	}
	if GUserHeader {
		AddUserHeaderWidget(sess)
	}
	sess.PopStateMap()
}

func (p GalleryPage) uploadListener(s Session, source FileUpload, value []byte) error {
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
		p.picEditor.Put(Animal_PhotoThumbnail, imageId)
		s.Get(Animal_PhotoThumbnail).Repaint()
	}
	return errOut
}

// ------------------------------------------------------------------------------------
// List
// ------------------------------------------------------------------------------------

const galleryItemPrefix = "gallery_item:"

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
	maxItems := Ternary(GListMultiItems, 50, 1)
	for i := 0; i < maxItems; i++ {
		t.names = append(t.names, RandomText(j, 12, false))
		t.ElementIds = append(t.ElementIds, i)
	}
	return t
}

func (g GalleryListImplementation) ItemStateMap(s Session, elementId int) JSMap {
	json := NewJSMap()
	json.Put("foo_text", ToString("Item #", elementId, g.names[elementId]))
	return json
}
