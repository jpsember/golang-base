package sample

import (
	. "github.com/jpsember/golang-base/base"
)

type sDemoConfig struct {
	simulate bool
	name     string
	greeting string
	target   int
}

type DemoConfigBuilderObj struct {
	// We embed the static struct
	sDemoConfig
}

type DemoConfigBuilder = *DemoConfigBuilderObj

// ---------------------------------------------------------------------------------------
// DemoConfig interface
// ---------------------------------------------------------------------------------------

type DemoConfig interface {
	DataClass
	Simulate() bool
	Name() string
	Greeting() string
	Target() int
	Build() DemoConfig
	ToBuilder() DemoConfigBuilder
}

var DefaultDemoConfig = newDemoConfig()

// Convenience method to get a fresh builder.
func NewDemoConfig() DemoConfigBuilder {
	return DefaultDemoConfig.ToBuilder()
}

// Construct a new static object, with fields initialized appropriately
func newDemoConfig() DemoConfig {
	var m = sDemoConfig{}
	m.greeting = "hello"
	m.target = 12
	return &m
}

// ---------------------------------------------------------------------------------------
// Implementation of static (built) object
// ---------------------------------------------------------------------------------------

func (v *sDemoConfig) Simulate() bool {
	return v.simulate
}

func (v *sDemoConfig) Name() string {
	return v.name
}

func (v *sDemoConfig) Greeting() string {
	return v.greeting
}

func (v *sDemoConfig) Target() int {
	return v.target
}

func (v *sDemoConfig) Build() DemoConfig {
	// This is already the immutable (built) version.
	return v
}

func (v *sDemoConfig) ToBuilder() DemoConfigBuilder {
	return &DemoConfigBuilderObj{sDemoConfig: *v}
}

func (v *sDemoConfig) ToJson() JSEntity {
	var m = NewJSMap()
	m.Put(DemoConfig_Simulate, v.simulate)
	m.Put(DemoConfig_Name, v.name)
	m.Put(DemoConfig_Greeting, v.greeting)
	m.Put(DemoConfig_Target, v.target)
	return m
}

func (v *sDemoConfig) Parse(source JSEntity) DataClass {
	var s = source.AsJSMap()
	var n = newDemoConfig().(*sDemoConfig)
	n.simulate = s.OptBool(DemoConfig_Simulate, false)
	n.name = s.OptString(DemoConfig_Name, "")
	n.greeting = s.OptString(DemoConfig_Greeting, "hello")
	n.target = s.OptInt(DemoConfig_Target, 12)
	return n
}

func (v *sDemoConfig) String() string {
	var x = v.ToJson().AsJSMap()
	return PrintJSEntity(x, true)
}

// ---------------------------------------------------------------------------------------
// Implementation of builder
// ---------------------------------------------------------------------------------------

func (v DemoConfigBuilder) Simulate() bool {
	return v.simulate
}

func (v DemoConfigBuilder) Name() string {
	return v.name
}

func (v DemoConfigBuilder) Greeting() string {
	return v.greeting
}

func (v DemoConfigBuilder) Target() int {
	return v.target
}

func (v DemoConfigBuilder) SetSimulate(simulate bool) DemoConfigBuilder {
	v.simulate = simulate
	return v
}

func (v DemoConfigBuilder) SetName(name string) DemoConfigBuilder {
	v.name = name
	return v
}

func (v DemoConfigBuilder) SetGreeting(greeting string) DemoConfigBuilder {
	v.greeting = greeting
	return v
}

func (v DemoConfigBuilder) SetTarget(target int) DemoConfigBuilder {
	v.target = target
	return v
}

func (v DemoConfigBuilder) Build() DemoConfig {
	// Construct a copy of the embedded static struct
	var b = v.sDemoConfig
	return &b
}

func (v DemoConfigBuilder) ToBuilder() DemoConfigBuilder {
	return v
}

func (v DemoConfigBuilder) ToJson() JSEntity {
	return v.Build().ToJson()
}

func (v DemoConfigBuilder) Parse(source JSEntity) DataClass {
	return DefaultDemoConfig.Parse(source)
}

func (v DemoConfigBuilder) String() string {
	return v.Build().String()
}

const DemoConfig_Simulate = "simulate"
const DemoConfig_Name = "name"
const DemoConfig_Greeting = "greeting"
const DemoConfig_Target = "target"

// Convenience method to parse a DemoConfig from a JSMap
func ParseDemoConfig(jsmap JSEntity) DemoConfig {
	m := jsmap.(JSMap)
	return DefaultDemoConfig.Parse(m).(DemoConfig)
}
