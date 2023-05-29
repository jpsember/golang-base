package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"math/rand"
	"strings"
)

type WidgetObj struct {
	Id string
}

type Widget = *WidgetObj

func (w Widget) WriteValue(v JSEntity) {
	NotImplemented("WriteValue")
}

func (w Widget) ReadValue() JSEntity {
	NotImplemented("ReadValue")
	return JBoolFalse
}

type WidgetManagerObj struct {
	rand                 *rand.Rand
	pendingColumnWeights []int
	// Note: this was a sorted map in the Java code
	widgetMap   map[string]Widget
	mSpanXCount int

	mGrowXFlag                  int
	mGrowYFlag                  int
	mPendingSize                int
	mPendingAlignment           int
	mPendingGravity             int
	mPendingMinWidthEm          float64
	mPendingMinHeightEm         float64
	mPendingMonospaced          bool
	mLineCount                  int
	mComboChoices               *Array[string]
	mPendingBooleanDefaultValue bool
	mPendingStringDefaultValue  string
	mPendingLabel               string
	mPendingTabTitle            string
	mPendingFloatingPointFlag   bool
	mPendingDefaultFloatValue   float64
	mPendingDefaultIntValue     int


	mPanelStack *Array[Grid]

}

type WidgetManager = *WidgetManagerObj

// Determine if a widget exists
func (m WidgetManager) Exists(id string) bool {
	return m.find(id) != nil
}

func (m WidgetManager) Get(id string) Widget {
	w := m.find(id)
	if w == nil {
		BadState("Can't find widget with id:", Quoted(id))
	}
	return w
}

func (m WidgetManager) find(id string) Widget {
	return m.widgetMap[id]
}

// ------------------------------------------------------------------
// Accessing widget values
// ------------------------------------------------------------------

/**
 * Set widgets' values. Used to restore app widgets to a previously saved
 * state
 */
func (m WidgetManager) SetWidgetValues(js *JSMap) {
	for id, val := range js.WrappedMap() {
		if m.Exists(id) {
			m.Get(id).WriteValue(val)
		}
	}
}

/**
 * Read widgets' values. Doesn't include widgets that have no ids, or whose
 * ids start with "."
 */
func (m WidgetManager) ReadWidgetValues() *JSMap {
	mp := NewJSMap()

	for id, w := range m.widgetMap {
		if strings.HasPrefix(id, ".") {
			continue
		}
		v := w.ReadValue()
		if v != nil {
			mp.Put(id, v)
		}
	}
	return mp
}

/**
 * Get value of string-valued widget
 */
func (m WidgetManager) Vs(id string) string {
	return m.Get(id).ReadValue().ToString()
}

/**
 * Set value of string-valued widget
 */
func (m WidgetManager) Sets(id string, v string) {
	m.Get(id).WriteValue(JString(v))
}

/**
 * Get value of boolean-valued widget
 */
func (m WidgetManager) Vb(id string) bool {
	result := false
	g := m.Get(id)
	if g != nil {
		result = g.ReadValue().ToBool()
	}
	return result
}

/**
 * Set value of boolean-valued widget
 */
func (m WidgetManager) Setb(id string, boolValue bool) bool {
	m.Get(id).WriteValue(JBool(boolValue))
	return boolValue
}

/**
 * Toggle value of boolean-valued widget
 *
 */
func (m WidgetManager) Toggle(id string) bool {
	return m.Setb(id, !m.Vb(id))
}

/**
 * Get value of integer-valued widget
 */
func (m WidgetManager) Vi(id string) int {
	return int(m.Get(id).ReadValue().ToInteger())
}

/**
 * Set value of integer-valued widget
 */
func (m WidgetManager) Seti(id string, v int) int {
	m.Get(id).WriteValue(JInteger(v))
	return v
}

/**
 * Get value of float-valued widget
 */
func (m WidgetManager) Vf(id string) float64 {
	return m.Get(id).ReadValue().ToFloat()
}

/**
 * Set value of double-valued widget
 */
func (m WidgetManager) SetF(id string, v float64) float64 {
	m.Get(id).WriteValue(JFloat(v))
	return v
}

// ------------------------------------------------------------------------------------

var digitsExpr = Regexp(`^\d+$`)

/**
 * <pre>
 *
 * Set the number of columns, and which ones can grow, for the next view in
 * the hierarchy. The columns expression is a string of column expressions,
 * which may be one of:
 *
 *     "."   a column with weight zero
 *     "x"   a column with weight 100
 *     "\d+" column with integer weight
 *
 * Spaces are ignored, except to separate integer weights from each other.
 * </pre>
 */
func (m WidgetManager) Columns(columnsExpr string) WidgetManager {
	CheckState(m.pendingColumnWeights == nil, "previous column weights were never used")

	columnSizes := NewArray[int]()
	for _, word := range strings.Split(columnsExpr, " ") {
		size := 0
		if digitsExpr.MatchString(word) {
			size = ParseIntM(word)
			columnSizes.Add(size)
		} else {
			w := []byte(word)
			for _, c := range w {
				if c == '.' {
					size = 0
				} else if c == 'x' {
					size = 100
				} else {
					BadArg("Can't parse columns expression:", Quoted(columnsExpr))
				}
				columnSizes.Add(size)
			}
		}
	}
	m.pendingColumnWeights = columnSizes.Array()
	return m
}

const (
	SIZE_DEFAULT = iota
	SIZE_TINY
	SIZE_SMALL
	SIZE_LARGE
	SIZE_HUGE
	SIZE_MEDIUM
)

const (
	ALIGNMENT_DEFAULT = iota
	ALIGNMENT_LEFT
	ALIGNMENT_CENTER
	ALIGNMENT_RIGHT
)

/**
 * Make next component added occupy remaining columns in its row
 */
func (m WidgetManager) Spanx() WidgetManager {
	m.mSpanXCount = -1
	return m
}

/**
 * Make next component added occupy some number of columns in its row
 */
func (m WidgetManager) SpanxCount(count int) WidgetManager {
	CheckArg(count > 0)
	m.mSpanXCount = count
	return m
}

/**
 * Skip a single cell
 */
func (m WidgetManager) skip() WidgetManager {
	Todo("skip()")
	//m.add(m.wrap(nil))
	return m
}

/**
 * Skip one or more cells
 */
func (m WidgetManager) skipN(count int) WidgetManager {
	m.SpanxCount(count)
	return m.skip()
}

/**
 * Set pending component, and the column it occupies, as 'growable'. This can
 * also be accomplished by using an 'x' when declaring the columns.
 * <p>
 * Calls growX(100)...
 */
func (m WidgetManager) GrowX() WidgetManager {
	return m.GrowXBy(100)
}

/**
 * Set pending component, and the column it occupies, as 'growable'. This can
 * also be accomplished by using an 'x' when declaring the columns.
 * <p>
 * Calls growY(100)...
 */
func (m WidgetManager) GrowY() WidgetManager {
	return m.GrowYBy(100)
}

/**
 * Set pending component's horizontal weight to a value > 0 (if it is already
 * less than this value)
 */
func (m WidgetManager) GrowXBy(weight int) WidgetManager {
	m.mGrowXFlag = MaxInt(m.mGrowXFlag, weight)
	return m
}

/**
 * Set pending component's vertical weight to a value > 0 (if it is already
 * less than this value)
 */
func (m WidgetManager) GrowYBy(weight int) WidgetManager {
	m.mGrowYFlag = MaxInt(m.mGrowYFlag, weight)
	return m
}

/**
 * Specify the component to use for the next open() call, instead of
 * generating one
 */
func (m WidgetManager) SetPendingContainer(component any) WidgetManager {
	//
	//public WidgetManager setPendingContainer(JComponent component) {
	//  checkState(mPanelStack.isEmpty(), "current panel stack isn't empty");
	//  mPendingContainer = component;
	//  return m
	return m
}

func (m WidgetManager) setPendingSize(value int) WidgetManager {
	m.mPendingSize = value
	return m
}

func (m WidgetManager) setPendingAlignment(value int) WidgetManager {
	m.mPendingAlignment = value
	return m
}

func (m WidgetManager) Small() WidgetManager {
	return m.setPendingSize(SIZE_SMALL)
}

func (m WidgetManager) Large() WidgetManager {
	return m.setPendingSize(SIZE_LARGE)
}

func (m WidgetManager) medium() WidgetManager {
	return m.setPendingSize(SIZE_MEDIUM)
}

func (m WidgetManager) tiny() WidgetManager {
	return m.setPendingSize(SIZE_TINY)
}

func (m WidgetManager) huge() WidgetManager {
	return m.setPendingSize(SIZE_HUGE)
}

func (m WidgetManager) left() WidgetManager {
	return m.setPendingAlignment(ALIGNMENT_LEFT)
}

func (m WidgetManager) right() WidgetManager {
	return m.setPendingAlignment(ALIGNMENT_RIGHT)
}

func (m WidgetManager) center() WidgetManager {
	return m.setPendingAlignment(ALIGNMENT_CENTER)
}

/**
 * Have next widget use a monospaced font
 */
func (m WidgetManager) Monospaced() WidgetManager {
	m.mPendingMonospaced = true
	return m
}

func (m WidgetManager) MinWidth(ems float64) WidgetManager {
	m.mPendingMinWidthEm = ems
	return m
}

func (m WidgetManager) MinHeight(ems float64) WidgetManager {
	m.mPendingMinHeightEm = ems
	return m
}

func (m WidgetManager) gravity(gravity int) WidgetManager {
	m.mPendingGravity = gravity
	return m
}

func (m WidgetManager) LineCount(numLines int) WidgetManager {
	CheckArg(numLines > 0)
	m.mLineCount = numLines
	return m
}

func (m WidgetManager) addLabel(id string) WidgetManager {
	text := m.ConsumePendingLabel()
	Todo("addLabel", text)
	//add(new LabelWidget(id, mPendingGravity, mLineCount, text, mPendingSize, mPendingMonospaced,
	//    mPendingAlignment));
	return m
}

func (m WidgetManager) Floats() WidgetManager {
	m.mPendingFloatingPointFlag = true
	return m
}

/**
 * Set default value for next boolean-valued control
 */
func (m WidgetManager) DefaultBool(value bool) WidgetManager {
	m.mPendingBooleanDefaultValue = value
	return m
}

func (m WidgetManager) DefaultString(value string) WidgetManager {
	m.mPendingStringDefaultValue = value
	return m
}

func (m WidgetManager) Label(value string) WidgetManager {
	m.mPendingLabel = value
	return m
}

/**
 * Set default value for next double-valued control
 */
func (m WidgetManager) defaultFloat(value float64) WidgetManager {
	m.Floats()
	m.mPendingDefaultFloatValue = value
	return m
}

/**
 * Set default value for next integer-valued control
 */
func (m WidgetManager) defaultInt(value int) WidgetManager {
	m.mPendingDefaultIntValue = value
	return m
}

/**
 * Append some choices for the next ComboBox
 */
func (m WidgetManager) Choices(choices ...string) WidgetManager {
	for _, s := range choices {
		if m.mComboChoices == nil {
			m.mComboChoices = NewArray[string]()
		}
		m.mComboChoices.Add(s)
	}
	return m
}

func (m WidgetManager) ConsumePendingBooleanDefaultValue() bool {
	v := m.mPendingBooleanDefaultValue
	m.mPendingBooleanDefaultValue = false
	return v
}

func (m WidgetManager) ConsumePendingFloatingPointFlag() bool {
	v := m.mPendingFloatingPointFlag
	m.mPendingFloatingPointFlag = false
	return v
}

func (m WidgetManager) ConsumePendingLabel() string {
	lbl := m.mPendingLabel
	m.mPendingLabel = ""
	return lbl
}

func (m WidgetManager) ConsumePendingStringDefaultValue() string {
	s := m.mPendingStringDefaultValue
	m.mPendingStringDefaultValue = ""
	return s
}

func (m WidgetManager) consumePendingTabTitle() string {
	tabNameExpression := m.mPendingTabTitle
	m.mPendingTabTitle = ""
	CheckState(tabNameExpression != "", "no pending tab title")
	return tabNameExpression
}

func verifyUsed(flag bool, name string) {
	if flag {
		return
	}
	BadState("unused value:", name)
}

func (m WidgetManager) clearPendingComponentFields() {
	// If some values were not used, issue warnings
	verifyUsed(0 != len(m.pendingColumnWeights), "pending column weights")
	//verifyUsed(mComboChoices, "pending combo choices");
	verifyUsed(m.mPendingDefaultIntValue == 0, "pendingDefaultIntValue")
	verifyUsed(m.mPendingStringDefaultValue != "", "mPendingStringDefaultValue")
	verifyUsed(m.mPendingLabel != "", "mPendingLabel ")
	verifyUsed(!m.mPendingFloatingPointFlag, "mPendingFloatingPoint")

	m.pendingColumnWeights = nil
	m.mSpanXCount = 0
	m.mGrowXFlag = 0
	m.mGrowYFlag = 0
	m.mPendingSize = SIZE_DEFAULT
	m.mPendingAlignment = ALIGNMENT_DEFAULT
	m.mPendingGravity = 0
	m.mPendingMinWidthEm = 0
	m.mPendingMinHeightEm = 0
	m.mPendingMonospaced = false
	m.mLineCount = 0
	m.mComboChoices = nil
	m.mPendingDefaultIntValue = 0
	m.mPendingBooleanDefaultValue = false
	m.mPendingStringDefaultValue = ""
	m.mPendingLabel = ""
	m.mPendingFloatingPointFlag = false
}


func RandomText(rand *rand.Rand, maxLength int, withLinefeeds bool) string {

	sample := "orhxxidfusuytelrcfdlordburswfxzjfjllppdsywgswkvukrammvxvsjzqwplxcpkoekiznlgsgjfonlugreiqvtvpjgrqotzu"

	sb := strings.Builder{}
	length := MinInt(maxLength, rand.Intn(maxLength+2))
	for sb.Len() < length {
		wordSize := rand.Intn(8) + 2
		if withLinefeeds && rand.Intn(4) == 0 {
			sb.WriteByte('\n')
		} else {
			sb.WriteByte(' ')
		}
		c := rand.Intn(len(sample) - wordSize)
		sb.WriteString(sample[c : c+wordSize])
	}
	return strings.TrimSpace(sb.String())
}


















/**
func (m WidgetManager)  ( ) WidgetManager {
**/


  func (m WidgetManager)  open( ) WidgetManager {
    return m.openFor("<no context>");
  }







  /**
   * Create a child view and push onto stack
   */
  func (m WidgetManager)  openFor(debugContext string ) WidgetManager {
grid := NewGrid()
    grid.SetContext(debugContext);

    {
      if  m.pendingColumnWeights == nil {
		 m.Columns("x");
	  }
      grid.ColumnSizes = m.pendingColumnWeights
	  m.pendingColumnWeights = nil

      JComponent panel;
      if (mPendingContainer != null) {
        panel = mPendingContainer;
        log2("pending container:", panel.getClass());
        mPendingContainer = null;
      } else {
        log2("constructing JPanel");
        panel = new JPanel();
        applyMinDimensions(panel, mPendingMinWidthEm, mPendingMinHeightEm);
      }
      panel.setLayout(buildLayout());
      addStandardBorderForSpacing(panel);
      grid.setWidget(wrap(panel));
    }
    add(grid.widget());
    mPanelStack.add(grid);
    log2("added grid to panel stack, its widget:", grid.widget().getClass());
    return m
  }

  /**
   * Pop view from the stack
   */
  func (m WidgetManager)  close( ) WidgetManager {
    return closeFor("<no context>");
  }

  /**
   * Pop view from the stack
   */
  func (m WidgetManager) closeFor (string debugContext   ) WidgetManager {
    Grid parent = pop(mPanelStack);
    if (verbose())
      log2("close", debugContext, compInfo(gridComponent(parent)));
    endRow();

    if (!(parent.widget() instanceof TabbedPaneWidget))
      assignViewsToGridLayout(parent);
    return m
  }

  /**
   * Verify that no unused 'pending' arguments exist, calls are balanced, etc
   */
  func (m WidgetManager)  finish( ) WidgetManager {
    m.clearPendingComponentFields();
    if (!mPanelStack.isEmpty())
      badState("panel stack nonempty; size:", mPanelStack.size());
    if (!mListenerStack.isEmpty())
      badState("listener stack nonempty; size:", mListenerStack.size());
    return m;
  }

  /**
   * If current row is only partially complete, add space to its end
   */
  func (m WidgetManager)  EndRow( ) WidgetManager {
    if (mPanelStack.isEmpty())
      return m;
    Grid parent = last(mPanelStack);
    if (parent.nextCellLocation().x != 0)
      spanx().addHorzSpace();
    return m;
  }

  func (m WidgetManager)  AddText(id string ) WidgetManager {
    TextWidget t = new TextWidget(consumePendingListener(), id, consumePendingStringDefaultValue(),
        mLineCount, mEditableFlag, mPendingSize, mPendingMonospaced, mPendingMinWidthEm, mPendingMinHeightEm);
    consumeTooltip(t);
    return m.add(t);
  }

  func (m WidgetManager)  AddHeader(text string ) WidgetManager {
    m.spanx();
    JLabel label = new JLabel(text);
    label.setBorder(
        new CompoundBorder(buildStandardBorderWithZeroBottom(), BorderFactory.createEtchedBorder()));
    label.setHorizontalAlignment(SwingConstants.CENTER);
    add(wrap(label));
    return m;
  }

  /**
   * Add a horizontal space to occupy cell(s) in place of other widgets
   */
  func (m WidgetManager)  AddHorzSpace( ) WidgetManager {
    add(wrap(new JPanel()));
    return m;
  }

  /**
   * Add a horizontal separator that visually separates components above from
   * below
   */
  func (m WidgetManager)  AddHorzSep( ) WidgetManager {
    m.spanx();
    m.add(wrap(new JSeparator(JSeparator.HORIZONTAL)));
    return m
  }

  /**
   * Add a vertical separator that visually separates components left from right
   */
  func (m WidgetManager)  AddVertSep( ) WidgetManager {
   m. spanx();
   m. growY();
   m. add(m.wrap(new JSeparator(JSeparator.VERTICAL)));
    return m
  }

  /**
   * Add a row that can stretch vertically to occupy the available space
   */
  func (m WidgetManager)  AddVertGrow( ) WidgetManager {
    //JComponent panel;
    //if (verbose())
    //  panel = colorPanel();
    //else
    //  panel = new JPanel();
    //spanx().growY();
    //add(wrap(panel));
    return m
  }

  /**
   * Add widget to view hierarchy
   */
  func (m WidgetManager)  Add (widget Widget ) WidgetManager {
    String id = null;
    if (widget.hasId())
      id = widget.id();
    log2("add widget", id != null ? id : "<anon>");

    if (id != null) {
      if (exists(widget.id()))
        badState("attempt to add widget id:", widget.id(), "that already exists");
      mWidgetMap.put(id, widget);
    }
    JComponent tooltipOwner = widget.componentForTooltip();
    if (tooltipOwner != null)
      consumeTooltip(tooltipOwner);
    addView(widget);
    return m
  }



  /**
   * Add a component to the current panel. Process pending constraints
   */
  func (m WidgetManager)  addView(widget Widget ) WidgetManager {
    consumeTooltip(widget);

    if (!mPanelStack.isEmpty())
      auxAddComponent(widget);

    clearPendingComponentFields();
    return m
  }

  func  (m WidgetManager) auxAddComponent(widget Widget )
    JComponent component = widget.swingComponent();

    // If the parent grid's widget is a tabbed pane,
    // add the component to it

    Grid grid = last(mPanelStack);
    if (grid.widget() instanceof TabbedPaneWidget) {
      TabbedPaneWidget tabPane = grid.widget();
      String tabIdNameExpression = consumePendingTabTitle(component);
      log2("adding a tab with name:", tabIdNameExpression);
      tabPane.add(tabIdNameExpression, component);
      return;
    }

    GridCell cell = new GridCell();
    cell.view = widget;
    IPoint nextGridCellLocation = grid.nextCellLocation();
    cell.x = nextGridCellLocation.x;
    cell.y = nextGridCellLocation.y;

    // determine location and size, in cells, of component
    int cols = 1;
    if (mSpanXCount != 0) {
      int remainingCols = grid.numColumns() - cell.x;
      if (mSpanXCount < 0)
        cols = remainingCols;
      else {
        if (mSpanXCount > remainingCols)
          throw new IllegalStateException(
              "requested span of " + mSpanXCount + " yet only " + remainingCols + " remain");
        cols = mSpanXCount;
      }
    }
    cell.width = cols;

    cell.growX = mGrowXFlag;
    cell.growY = mGrowYFlag;

    // If any of the spanned columns have 'grow' flag set, set it for this component
    for (int i = cell.x; i < cell.x + cell.width; i++) {
      int colSize = grid.columnSizes()[i];
      cell.growX = Math.max(cell.growX, colSize);
    }

    // "paint" the cells this view occupies by storing a copy of the entry in each cell
    for (int i = 0; i < cols; i++)
      grid.addCell(cell);
  }


  func (m WidgetManager) AddButton ( id string) WidgetManager {
    ButtonWidget button = new ButtonWidget(consumePendingListener(), id, consumePendingLabel(true));
    return add(button);
  }


  func (m WidgetManager) AddToggleButton (id string ) WidgetManager {
    ToggleButtonWidget button = new ToggleButtonWidget(consumePendingListener(), id,
        consumePendingLabel(true), consumePendingBooleanDefaultValue());
    return add(button);
  }

  func (m WidgetManager) AddLabel (id string ) WidgetManager {
    String text = consumePendingLabel(true);
    add(new LabelWidget(id, mPendingGravity, mLineCount, text, mPendingSize, mPendingMonospaced,
        mPendingAlignment));
    return m
  }



  func (m WidgetManager) AddChoiceBox (id string ) WidgetManager {
    ComboBoxWidget c = new ComboBoxWidget(consumePendingListener(), id, mComboChoices);
    return add(c);
  }


  // ------------------------------------------------------------------
  // Layout manager
  // ------------------------------------------------------------------

  private LayoutManager buildLayout() {
    return new GridBagLayout();
  }

  private static <T extends JComponent> T gridComponent(Grid grid) {
    Widget widget = grid.widget();
    return widget.swingComponent();
  }

  private void assignViewsToGridLayout(Grid grid) {
    grid.propagateGrowFlags();
    Widget containerWidget = grid.widget();
    JComponent container = containerWidget.swingComponent();

    int gridWidth = grid.numColumns();
    int gridHeight = grid.numRows();
    for (int gridY = 0; gridY < gridHeight; gridY++) {
      for (int gridX = 0; gridX < gridWidth; gridX++) {
        GridCell cell = grid.cellAt(gridX, gridY);
        if (cell.isEmpty())
          continue;

        // If cell's coordinates don't match our iteration coordinates, we've
        // already added this cell
        if (cell.x != gridX || cell.y != gridY)
          continue;

        GridBagConstraints gc = new GridBagConstraints();

        float weightX = cell.growX;
        gc.weightx = weightX;
        gc.gridx = cell.x;
        gc.gridwidth = cell.width;
        gc.weighty = cell.growY;
        gc.gridy = cell.y;
        gc.gridheight = 1;

        Widget widget = (Widget) cell.view;
        JComponent component = widget.swingComponent();
        // Padding widgets have no views
        if (component == null)
          continue;

        // Not using gc.anchor
        gc.fill = GridBagConstraints.BOTH;

        // Not using gravity
        container.add(widget.swingComponent(), gc);
      }
    }
  }

  static Font getFont(boolean monospaced, int widgetFontSize) {
    int fontSize;
    switch (widgetFontSize) {
    case SIZE_DEFAULT:
      fontSize = 16;
      break;
    case SIZE_MEDIUM:
      fontSize = 16;
      break;
    case SIZE_SMALL:
      fontSize = 12;
      break;
    case SIZE_LARGE:
      fontSize = 22;
      break;
    case SIZE_HUGE:
      fontSize = 28;
      break;
    default:
      alert("unsupported widget font size:", widgetFontSize);
      fontSize = 16;
      break;
    }

    Integer mapKey = fontSize + (monospaced ? 0 : 1000);

    Font f = sFontMap.get(mapKey);
    if (f == null) {
      if (monospaced)
        f = new Font("Monaco", Font.PLAIN, fontSize);
      else
        f = new Font("Lucida Grande", Font.PLAIN, fontSize);
      sFontMap.put(mapKey, f);
    }
    return f;
  }
