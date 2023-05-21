package app

import (
	. "github.com/jpsember/golang-base/base"
	"strconv"
	"strings"
)

type CmdLineArgs struct {
	logger            Logger
	banner            string
	locked            bool
	opt               *Option
	namedOptionMap    map[string]*Option
	optionList        *Array[string]
	extraArgsCursor   int
	exArgs            []string
	helpShown         bool
	extraArguments    *Array[string]
	stillHandlingArgs bool
}

func NewCmdLineArgs() *CmdLineArgs {
	var c = new(CmdLineArgs)
	c.logger = NewLogger(c)
	c.namedOptionMap = make(map[string]*Option)
	c.optionList = NewArray[string]()
	c.extraArguments = NewArray[string]()
	return c
}

func (c *CmdLineArgs) WithBanner(banner string) *CmdLineArgs {
	c.banner = banner
	return c
}

func (c *CmdLineArgs) Parse(args []string) {
	c.lock()
	var argList = c.unpackArguments(args)
	c.readArgumentValues(argList)
}

// Read the arguments, and return an array that contains
// [*Option, value, *Option, *Option, value, ...]
func (c *CmdLineArgs) unpackArguments(args []string) *Array[any] {
	var pattern = Regexp(`^--?[a-z_A-Z][a-z_A-Z\-]*$`)

	var argList = NewArray[any]()
	for _, arg := range args {
		if pattern.MatchString(arg) {
			if strings.HasPrefix(arg, "--") {
				var opt = c.findOption(arg[2:])
				opt.Invocation = arg
				argList.Add(opt)
			} else {
				for i := 1; i < len(arg); i++ {
					var opt = c.findOption(arg[i : i+1])
					opt.Invocation = arg
					argList.Add(opt)
				}
			}
			continue
		}
		argList.Add(arg)
	}
	return argList
}

func (c *CmdLineArgs) Help() {
	if c.helpShown {
		return
	}
	c.helpShown = true
	Pr("....not implemented; help...")
}

// Process the unpacked list of options and values, assigning values to the
// options
func (c *CmdLineArgs) readArgumentValues(args *Array[any]) {

	var pr func(message ...any)

	if false {
		pr = Pr
	} else {
		pr = func(message ...any) {}
	}

	pr("processing unpacked list of options and values")

	var cursor = 0
	for cursor < args.Size() {
		var arg = args.Get(cursor)
		pr("cursor", cursor, "arg:", arg)
		cursor++

		opt, ok := arg.(*Option)
		if ok {
			pr("...it is an option, type:", opt.Type, "name;", opt.LongName)
			if opt.Type == Bool {
				opt.BoolValue = true
				pr("set boolean value to true")
				continue
			}
			if opt.LongName == "help" {
				c.Help()
				break
			}

			//for
			{
				// We expect a string to be the next argument.
				// If it is missing, or is another option, that's a problem
				missing := true
				var value string
				if cursor < args.Size() {
					arg = args.Get(cursor)
					pr("cursor:", cursor, "read next arg as value for option:", arg)
					cursor++
					v, ok := arg.(string)
					if ok {
						missing = false
						value = v
					}
				}

				if missing {
					BadArg("Expected value for argument:", opt.Invocation)
				}

				switch opt.Type {
				case Int:
					intVal, err := strconv.ParseInt(value, 10, 64)
					CheckArg(err == nil, "Can't parse int:", value, "from", opt.Invocation)
					opt.IntValue = intVal
					pr("set int value", intVal)
				case Float:
					float64Val, err := strconv.ParseFloat(value, 64)
					CheckArg(err == nil, "Can't parse float:", value, "from", opt.Invocation)
					opt.FloatValue = float64Val
					pr("set float value", float64Val)
				case Str:
					opt.StringValue = value
					pr("set string value", value)
				default:
					Halt("unsupported type:", opt.Type)
				}
			}
		} else {
			// This was an argument not tied to an option;
			// add them to the extra arguments list
			c.extraArguments.Add(arg.(string))
		}
	}
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

func (c *CmdLineArgs) claimName(name string) {
	if value, hasKey := c.namedOptionMap[name]; hasKey {
		BadState("option already exists:", name, "for:", value.Description)
	}
	c.namedOptionMap[name] = c.option()
}

func (c *CmdLineArgs) Add(longName string) *CmdLineArgs {
	c.checkNotLocked()
	c.opt = NewOption(longName)
	c.claimName(longName)
	c.optionList.Add(longName)
	return c
}

func (c *CmdLineArgs) ShortName(shortName string) *CmdLineArgs {
	c.option().ShortName = shortName
	c.claimName(shortName)
	return c
}

func (c *CmdLineArgs) option() *Option {
	if c.opt == nil {
		BadState("No current Option")
	}
	return c.opt
}

// Set type of current option to int
func (c *CmdLineArgs) SetInt() *CmdLineArgs {
	c.option().Type = Int
	return c
}

func (c *CmdLineArgs) SetFloat() *CmdLineArgs {
	c.option().Type = Float
	return c
}

func (c *CmdLineArgs) SetString() *CmdLineArgs {
	c.option().Type = Str
	return c
}

func (c *CmdLineArgs) checkNotLocked() {
	CheckState(!c.locked)
}

func (c *CmdLineArgs) Desc(description string) *CmdLineArgs {
	c.option().Description = description
	return c
}

func (c *CmdLineArgs) chooseShortNames() {
	for _, key := range c.optionList.Array() {
		c.opt = c.namedOptionMap[key]

		j := 0
		// If option has prefix "no", it's probably 'noXXX', so avoid
		// deriving short name from 'n' or 'o'
		if strings.HasPrefix(key, "no") {
			j = 2
		}
		for ; c.option().ShortName == ""; j++ {
			if j >= len(key) {
				// choose first unused character

				poss := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
				for k := 0; k < len(poss); k++ {
					candidate := poss[k : k+1]
					if !HasKey(c.namedOptionMap, candidate) {
						c.claimName(candidate)
						c.option().ShortName = candidate
						break
					}
				}
				break
			}

			candidate := key[j : j+1]
			if HasKey(c.namedOptionMap, candidate) {
				candidate = strings.ToUpper(candidate)
			}
			if !HasKey(c.namedOptionMap, candidate) {
				c.claimName(candidate)
				c.option().ShortName = candidate
				break
			}
		}
		c.validate(c.option().ShortName != "", "can't find short name for", key)
	}

}

func (c *CmdLineArgs) validate(condition bool, message ...any) {
	if !condition {
		Die(message...)
	}
}

type OptType int

const (
	Bool = iota
	Int
	Float
	Str
)

// Representation of a command line option
type Option struct {
	LongName     string
	ShortName    string
	Type         OptType
	typeDefined  bool
	Description  string
	Invocation   string
	BoolValue    bool
	IntValue     int64
	FloatValue   float64
	StringValue  string
	DebugCounter int
}

var debugCounter = 500

func (x *Option) String() string {
	return "Option #" + strconv.Itoa(x.DebugCounter) + " '" + x.LongName + "' "
}

func NewOption(longName string) *Option {
	var opt = Option{
		LongName:     longName,
		Type:         Bool,
		DebugCounter: debugCounter,
	}
	debugCounter++
	return &opt
}

func (opt *Option) SetType(t OptType) {
	opt.Type = t
}

func (c *CmdLineArgs) HandlingArgs() bool {
	c.stillHandlingArgs = !c.stillHandlingArgs
	if !c.stillHandlingArgs {
		//ensureRequiredArgsProvided()
		if c.HasNextArg() {
			Pr("...done handling args; argument(s) remain:", c.PeekNextArg())
		}
	}
	return c.stillHandlingArgs
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
	return c.extraArguments.Array()
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
	CheckState(opt.Type == Bool, "type mismatch", optionName)
	return opt.BoolValue
}

// Get the value of a string option
func (c *CmdLineArgs) GetString(optionName string) string {
	var opt = c.findOption(optionName)
	CheckState(opt.Type == Str, "type mismatch", optionName)
	return opt.StringValue
}

func (c *CmdLineArgs) findOption(optionName string) *Option {
	opt := c.namedOptionMap[optionName]
	CheckState(opt != nil, "unrecognized option:", optionName)
	//Pr("found option", optionName, "as:", opt)
	return opt
}
