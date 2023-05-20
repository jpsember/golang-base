package app

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)
import . "github.com/jpsember/golang-base/files"

type CmdLineArgs struct {
	logger Logger
	banner string
	locked bool

	opt             *Option
	namedOptionMap  map[string]*Option
	optionList      *Array[string]
	extraArgsCursor int
	exArgs          []string
	helpShown       bool
}

func NewCmdLineArgs() *CmdLineArgs {
	var c = new(CmdLineArgs)
	c.logger = NewLogger(c)
	c.namedOptionMap = make(map[string]*Option)
	c.optionList = NewArray[string]()
	return c
}

func (c *CmdLineArgs) WithBanner(banner string) *CmdLineArgs {
	c.banner = banner
	return c
}

func (c *CmdLineArgs) Parse(args []string) {
	c.lock()
}

func (c *CmdLineArgs) lock() {
	if c.locked {
		return
	}
	c.Add("help").Desc("Show this message")
	// Reserve the 'h' short name for the help option
	c.ShortName("h")

	c.locked = true
	c.chooseShortNames()
}

func (c *CmdLineArgs) claimName(name string, owner *Option) {
	if value, hasKey := c.namedOptionMap[name]; hasKey {
		BadState("option already exists:", name, "for:", value.Description)
	}
	c.namedOptionMap[name] = owner
}

func (c *CmdLineArgs) Add(longName string) *CmdLineArgs {
	c.checkNotLocked()
	c.opt = NewOption(longName)
	c.claimName(longName, c.opt)
	c.optionList.Add(longName)
	return c
}

func (c *CmdLineArgs) ShortName(shortName string) *CmdLineArgs {
	c.opt.ShortName = shortName
	c.claimName(shortName, c.opt)
	return c
}

func (c *CmdLineArgs) checkNotLocked() {
	CheckState(!c.locked)
}

func (c *CmdLineArgs) Desc(description string) *CmdLineArgs {
	c.opt.Description = description
	return c
}
func (c *CmdLineArgs) chooseShortNames() {
	for _, key := range c.optionList.Array() {
		opt := c.namedOptionMap[key]

		j := 0
		// If option has prefix "no", it's probably 'noXXX', so avoid
		// deriving short name from 'n' or 'o'
		if strings.HasPrefix(key, "no") {
			j = 2
		}
		for ; opt.ShortName == ""; j++ {
			if j >= len(key) {
				// choose first unused character

				poss := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
				for k := 0; k < len(poss); k++ {
					candidate := poss[k : k+1]
					if !HasKey(c.namedOptionMap, candidate) {
						c.claimName(candidate, opt)
						opt.ShortName = candidate
						break
					}
				}
				break
			}

			candidate := key[j : j+1]
			if !HasKey(c.namedOptionMap, candidate) {
				c.claimName(candidate, opt)
				opt.ShortName = candidate
				break
			}
			candidate = strings.ToUpper(candidate)
			if !HasKey(c.namedOptionMap, candidate) {
				c.claimName(candidate, opt)
				opt.ShortName = candidate
				break
			}
		}
		c.validate(opt.ShortName != "", "can't find short name for", key)
	}

}

func (c *CmdLineArgs) validate(condition bool, message ...any) {
	if !condition {
		Die(message...)
	}
}

type OptType int

const (
	Unknown = iota
	Bool
)

// Representation of a command line option
type Option struct {
	LongName    string
	ShortName   string
	Type        OptType
	typeDefined bool
	Description string

	/**
		<pre>
		    public boolean hasValue() {
	      return !mValues.isEmpty();
	    }

	    public String mLongName;
	    public String mShortName;
	    public Object mDefaultValue;
	    public String mDescription = "*** No description! ***";
	    public int mType;
	    public boolean mArray;
	    // Number of values expected; -1 if variable-length array
	    public int mExpectedValueCount = 1;
	    public boolean mTypeDefined;
	    public String mInvocation;
	    public ArrayList<Object> mValues = arrayList();
	  }

		</pre>
	*/
}

func NewOption(longName string) *Option {
	var opt = Option{
		LongName: longName,
		Type:     Bool,
	}
	return &opt
}

func (opt *Option) SetType(t OptType) {
	CheckState(opt.Type == Unknown)
	opt.Type = t
}

// // ------------------------------------------------------------------
// // Additional functionality moved from App class
// // ------------------------------------------------------------------
//
// private String[] mExtraArgs;
// private int mExtraArgsCursor;
// private Map<String, Object> mArgValueMap = hashMap();
// private boolean mStillHandlingArgs;
//
//	private int extraArgsCursor() {
//	  return mExtraArgsCursor;
//	}
func (c *CmdLineArgs) ExtraArgs() []string {
	if c.exArgs == nil {
		c.exArgs = []string{}
		Todo("determine extra args")
	}
	return c.exArgs
}

//	private String[] extraArgs() {
//	  if (mExtraArgs == null)
//	    mExtraArgs = getExtras();
//	  return mExtraArgs;
//	}
func (c *CmdLineArgs) HasNextArg() bool {
	return c.extraArgsCursor < len(c.ExtraArgs())
}

//	  return extraArgsCursor() < extraArgs().length;
//	}
//
//	public final void assertArgsDone() {
//	  if (hasNextArg())
//	    fail("Unexpected extra argument(s): " + nextArg());
//	}
//
//	public final String peekNextArg() {
//	  if (!hasNextArg())
//	    fail("missing argument(s)");
//	  return extraArgs()[mExtraArgsCursor];
//	}
func (c *CmdLineArgs) PeekNextArg() string {
	if !c.HasNextArg() {
		BadState("missing argument(s)")
	}
	return c.ExtraArgs()[c.extraArgsCursor]
}

func (c *CmdLineArgs) NextArg() string {
	arg := c.PeekNextArg()
	c.extraArgsCursor++
	return arg
}

// public final String nextArg(String defaultValue) {
//   String value = defaultValue;
//   if (hasNextArg() || value == null)
//     value = nextArg();
//   return value;
// }

func (c *CmdLineArgs) HelpShown() bool {
	return c.helpShown
}

// Get the boolean value supplied for an option, or its default if none was given. If no default was specified, assume it was false.
func (c *CmdLineArgs) Get(optionName string) bool {

	var opt = c.findOption(optionName)
	Pr("opt:", opt)
	CheckState(opt.Type == Bool, "type mismatch", optionName)
	Todo("do we need to store a default value somewhere?")
	return false
}

func (c *CmdLineArgs) findOption(optionName string) *Option {
	opt := c.namedOptionMap[optionName]
	CheckState(opt != nil, "unrecognized option:", optionName)
	return opt
}
