package transform

import (
	"regexp"
	"strings"

	"github.com/gowebapi/webidlgenerator/types"
)

type action interface {
	IsGlobal() bool
	ExecuteCallback(instance *types.Callback, trans *Transform)
	ExecuteDictionary(instance *types.Dictionary, trans *Transform)
	ExecuteEnum(instance *types.Enum, targets map[string]renameTarget, trans *Transform)
	ExecuteInterface(instance *types.Interface, targets map[string]renameTarget, trans *Transform)
	Reference() ref
}

// propary change on interface/enum/etc, like package name
type property struct {
	Name  string
	Value string
	Ref   ref
}

func (t *property) ExecuteCallback(instance *types.Callback, trans *Transform) {
	if f, ok := callbackProperties[t.Name]; ok {
		f(instance, t.Value)
	} else {
		trans.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(callbackPropertyNames, ", "))
	}
}

func (t *property) ExecuteDictionary(instance *types.Dictionary, trans *Transform) {
	if f, ok := dictionaryProperties[t.Name]; ok {
		f(instance, t.Value)
	} else {
		trans.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(dictionaryPropertyNames, ", "))
	}
}

func (t *property) ExecuteEnum(instance *types.Enum, targets map[string]renameTarget, trans *Transform) {
	if f, ok := enumProperties[t.Name]; ok {
		f(instance, t.Value)
	} else {
		trans.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(enumPropertyNames, ", "))
	}
}

func (t *property) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, trans *Transform) {
	if f, ok := interfaceProperties[t.Name]; ok {
		f(value, t.Value)
	} else {
		trans.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(interfacePropertyNames, ", "))
	}
}

func (t property) IsGlobal() bool {
	value, found := globalProperties[t.Name]
	return value && found
}

func (t property) Reference() ref {
	return t.Ref
}

// rename a method or attribute name
type rename struct {
	Name  string
	Value string
	Ref   ref
}

type renameTarget interface {
	Name() *types.MethodName
}

func (t *rename) ExecuteCallback(instance *types.Callback, trans *Transform) {
	trans.messageError(t.Ref, "callback doesn't have any attributes or methods that can be renamed")
}

func (t *rename) ExecuteDictionary(value *types.Dictionary, trans *Transform) {
	trans.messageError(t.Ref, "dictionary doesn't have any attributes or methods that can be renamed")
}

func (t *rename) ExecuteEnum(value *types.Enum, targets map[string]renameTarget, trans *Transform) {
	genericRename(t.Name, t.Value, t.Ref, targets, trans)
}

func (t *rename) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, trans *Transform) {
	genericRename(t.Name, t.Value, t.Ref, targets, trans)
}

func genericRename(name, value string, ref ref, targets map[string]renameTarget, trans *Transform) {
	if target, found := targets[name]; found {
		target.Name().Def = value
	} else {
		trans.messageError(ref, "unknown rename target '%s'", name)
	}
}

func (t *rename) IsGlobal() bool {
	return false
}

func (t rename) Reference() ref {
	return t.Ref
}

// do a command on multiple types at one
type globalRegExp struct {
	Match *regexp.Regexp
	Type  matchType
	What  action
	Ref   ref
}

func (t *globalRegExp) IsGlobal() bool {
	return true
}

func (t globalRegExp) Reference() ref {
	return t.Ref
}

func (t *globalRegExp) ExecuteCallback(instance *types.Callback, trans *Transform) {
	t.What.ExecuteCallback(instance, trans)
}

func (t *globalRegExp) ExecuteDictionary(value *types.Dictionary, trans *Transform) {
	t.What.ExecuteDictionary(value, trans)
}

func (t *globalRegExp) ExecuteEnum(value *types.Enum, targets map[string]renameTarget, trans *Transform) {
	t.What.ExecuteEnum(value, targets, trans)
}

func (t *globalRegExp) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, trans *Transform) {
	t.What.ExecuteInterface(value, targets, trans)
}
