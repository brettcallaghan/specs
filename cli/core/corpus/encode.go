package corpus

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/brettcallaghan/specs/cli/core/envelope"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// The index is encoded through a view of [specmodel.SpecIndex] whose
// entries carry their frontmatter pre-encoded by [envelope.Encode], and
// whose every other field is the generated field itself.
//
// The generated frontmatter encoder refuses a required enum left at its
// empty value, and a member that omits one — `status`, a blueprint's
// `kind` — is a finding for check, never a reason the index cannot be
// written (AC-SPEC-002.4, AC-SPEC-002.5). envelope.Encode omits such a
// property instead, and is byte-identical to the generated encoder
// everywhere else, so the index of a corpus with no such member is
// unchanged.
//
// The view is derived from the generated types by reflection rather than
// declared by hand, so a field added to either generated struct reaches
// the index without this file changing. Neither generated struct encodes
// itself, so a struct of the same fields encodes as it does.

const (
	entriesField     = "Entries"
	frontmatterField = "Frontmatter"
)

type indexViewTypes struct {
	index reflect.Type
	entry reflect.Type
}

var indexView = sync.OnceValue(func() indexViewTypes {
	raw := reflect.TypeOf(json.RawMessage(nil))
	entry := reflect.StructOf(replaceField(reflect.TypeOf(specmodel.IndexEntry{}), frontmatterField, raw))
	index := reflect.StructOf(replaceField(reflect.TypeOf(specmodel.SpecIndex{}), entriesField, reflect.SliceOf(entry)))
	return indexViewTypes{index: index, entry: entry}
})

// replaceField returns t's fields with the one named name retyped as
// replacement, keeping its tag. It panics when t has no such field, which
// would mean the generated model no longer matches this file.
func replaceField(t reflect.Type, name string, replacement reflect.Type) []reflect.StructField {
	fields := make([]reflect.StructField, t.NumField())
	found := false
	for i := range fields {
		fields[i] = t.Field(i)
		if fields[i].Name == name {
			fields[i].Type = replacement
			found = true
		}
	}
	if !found {
		panic(fmt.Sprintf("corpus: %s has no field %s to encode through", t, name))
	}
	return fields
}

// encodable returns index as its encoding view.
func encodable(index specmodel.SpecIndex) (any, error) {
	types := indexView()
	src := reflect.ValueOf(index)
	dst := reflect.New(types.index).Elem()
	for i := 0; i < src.NumField(); i++ {
		if types.index.Field(i).Name != entriesField {
			dst.Field(i).Set(src.Field(i))
			continue
		}
		entries, err := encodableEntries(src.Field(i), types.entry)
		if err != nil {
			return nil, err
		}
		dst.Field(i).Set(entries)
	}
	return dst.Interface(), nil
}

func encodableEntries(src reflect.Value, entryType reflect.Type) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(reflect.SliceOf(entryType)), nil
	}
	dst := reflect.MakeSlice(reflect.SliceOf(entryType), src.Len(), src.Len())
	for j := 0; j < src.Len(); j++ {
		from, to := src.Index(j), dst.Index(j)
		for k := 0; k < from.NumField(); k++ {
			if entryType.Field(k).Name != frontmatterField {
				to.Field(k).Set(from.Field(k))
				continue
			}
			fm := from.Field(k)
			if fm.IsNil() {
				continue
			}
			encoded, err := envelope.Encode(*fm.Interface().(*specmodel.TypedFrontmatter))
			if err != nil {
				path := from.FieldByName("Path").String()
				return reflect.Value{}, fmt.Errorf("%s: %w", path, err)
			}
			to.Field(k).Set(reflect.ValueOf(json.RawMessage(encoded)))
		}
	}
	return dst, nil
}

// decodeIndex is the inverse of [encodable]: it reads raw through the
// same view and decodes each entry's frontmatter with [envelope.Decode],
// which accepts the required enum the write omitted while still decoding
// every property through its generated type — an out-of-vocabulary value
// or a malformed timestamp is still an error. The generated decoder would
// refuse the omission, so an index the writer produced could not be read
// back by its own reader.
func decodeIndex(raw []byte) (specmodel.SpecIndex, error) {
	types := indexView()
	view := reflect.New(types.index)
	if err := json.Unmarshal(raw, view.Interface()); err != nil {
		return specmodel.SpecIndex{}, err
	}
	src := view.Elem()
	var index specmodel.SpecIndex
	dst := reflect.ValueOf(&index).Elem()
	for i := 0; i < src.NumField(); i++ {
		if types.index.Field(i).Name != entriesField {
			dst.Field(i).Set(src.Field(i))
			continue
		}
		entries, err := decodedEntries(src.Field(i), dst.Field(i).Type())
		if err != nil {
			return specmodel.SpecIndex{}, err
		}
		dst.Field(i).Set(entries)
	}
	return index, nil
}

func decodedEntries(src reflect.Value, sliceType reflect.Type) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(sliceType), nil
	}
	dst := reflect.MakeSlice(sliceType, src.Len(), src.Len())
	for j := 0; j < src.Len(); j++ {
		from, to := src.Index(j), dst.Index(j)
		for k := 0; k < from.NumField(); k++ {
			if sliceType.Elem().Field(k).Name != frontmatterField {
				to.Field(k).Set(from.Field(k))
				continue
			}
			raw := from.Field(k).Interface().(json.RawMessage)
			if len(raw) == 0 || string(raw) == "null" {
				continue
			}
			decoded, err := envelope.Decode(raw)
			if err != nil {
				path := from.FieldByName("Path").String()
				return reflect.Value{}, fmt.Errorf("%s: %w", path, err)
			}
			to.Field(k).Set(reflect.ValueOf(&decoded))
		}
	}
	return dst, nil
}
