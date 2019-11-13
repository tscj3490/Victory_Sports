package l10n

import (
	"strings"
	"bytes"
	"github.com/jinzhu/gorm"
	"reflect"
	"log"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n"
	"github.com/ugorji/go/codec"
	"github.com/fatih/structs"
)

/*
instead of using a multiqiery i18n thingy we store the strings in via msgpack in the string field
and provide a helper method to read from it.
*/
// the interface helps to detect the model that the user
// added the Methods to via struct composition
var mh = codec.MsgpackHandle{}
func Get(localized string, locale string) string {
	if len(locale) > 2 {
		locale = convertLang2Locale(locale)
	}
	var out map[string]string

	dec := codec.NewDecoderBytes([]byte(localized), &mh)
	if err := dec.Decode(&out); err != nil {
		log.Printf("l10n.Get Err: %v Value: %v", err, localized)
		return ""
	}

	//if err := msgpack.Unmarshal([]byte(localized), &out); err != nil {
	//	log.Printf("l10n.Get Err: %v Value: %v", err, localized)
	//	return ""
	//}
	resp, ok := out[locale];
	if !ok {
		return ""
	}
	return resp
}
func Set(localized string, locale string, value string) string {
	if len(locale) > 2 {
		locale = convertLang2Locale(locale)
	}
	var out map[string]string
	dec := codec.NewDecoderBytes([]byte(localized), &mh)
	if err := dec.Decode(&out); err != nil {
		log.Printf("l10n.Set Err: %v Value: %v", err, localized)
		return ""
	}
	//if err := msgpack.Unmarshal([]byte(localized), &out); err != nil {
	//	return ""
	//}
	out[locale] = value


	var (
		b []byte
	)
	enc := codec.NewEncoderBytes(&b, &mh)
	if err := enc.Encode(out); err != nil {
		return ""
	}

	//b, err := msgpack.Marshal(out)
	//if err != nil {
	//	return ""
	//}
	return bytes.NewBuffer(b).String()
}
func SetAll(localized map[string]string) string {
	var (
		out []byte
	)
	enc := codec.NewEncoderBytes(&out, &mh)
	if err := enc.Encode(localized); err != nil {
		return ""
	}
	//
	//
	//b, err := msgpack.Marshal(localized)
	//if err != nil {
	//	return ""
	//}
	return bytes.NewBuffer(out).String()
}
func GetAll(localized string) map[string]string {
	var out map[string]string

	dec := codec.NewDecoderBytes([]byte(localized), &mh)
	if err := dec.Decode(&out); err != nil {
		log.Printf("l10n.Get Err: %v Value: %v", err, localized)
		return out
	}
	return out
}
func GetByFieldName(m interface{}, fieldName string, locale string) string {
	var out string
	if len(locale) > 2 {
		locale = convertLang2Locale(locale)
	}
	s := structs.New(m)
	f := s.Field(fieldName)
	//log.Printf("GBFN 1: %v", f)
	if f == nil {
		return out
	}
	target := f.Tag("l10nTarget")
	//log.Printf("GBFN 2: %v", target)
	if target == "" {
		return out
	}
	t := s.Field(target)
	//log.Printf("GBFN 3: %v", t)
	if t == nil {
		return out
	}
	value, ok := t.Value().(string)
	//log.Printf("GBFN 4: %v", value)
	if !ok {
		return out
	}

	var localizedMap map[string]string
	if value != "" {
		localizedMap = GetAll(value)
	}
	//log.Printf("GBFN 5: %v", localizedMap)

	out, ok = localizedMap[locale]
	if !ok {
		//log.Printf("%v:%v", f.Name(), f.Value())
		out, ok = f.Value().(string)
	}
	//log.Printf("GBFN 6: %v %v", out, locale)

	return out
}

func convertLang2Locale(lang string) string {
	locale, _ := i18n.Locales[lang]
	return locale
}

type L10NModel interface {
	L10NFields() []string
	Map() map[string]interface{}
}
func L10NFieldsInner(m interface{})[]string {
	s := structs.New(m)
	fields := []string{}
	for _, f := range s.Fields() {
		tagValue := f.Tag("l10n")
		if tagValue == "" {
			continue
		}
		fields = append(fields, f.Name())
	}
	return fields
}
func L10NMapInner(m interface{}) map[string]interface{} {
	s := structs.New(m)
	resp := map[string]interface{}{}
	for _, f := range s.Fields() {
		if f.IsEmbedded() {
			for _, ef := range f.Fields() {
				resp[ef.Name()] = ef.Value()
			}
			continue
		}
		resp[f.Name()] = f.Value()
	}
	return resp
}

// takes a list of structs and converts them into a list of map[string]interface{}
// that include the translation
func List(l10nModel L10NModel, l10nList []L10NModel) []map[string]interface{} {
	resp := []map[string]interface{}{}
	fields := l10nModel.L10NFields()
	fl := map[string]bool{}
	for _, f := range fields {
		fl[f] = true
	}

	// loop over all list items
	for _, itm := range l10nList {
		resp = append(resp, Unpack(itm))
	}
	return resp
}
func Unpack(itm L10NModel) map[string]interface{} {
	fields := itm.L10NFields()
	fl := map[string]bool{}
	for _, f := range fields {
		fl[f] = true
	}
	m := itm.Map()
	n := map[string]interface{}{}

	for k, v := range m {
		if _, ok := fl[k]; !ok {
			// regular field found
			n[k] = v
			continue
		}
		// translatable field found
		var out map[string]string
		localized, ok := v.(string)
		if !ok {
			log.Printf("l10n.Unpack Err: String Convert Failed Value: %v", v)
			n[k] = v
			continue
		}

		dec := codec.NewDecoderBytes([]byte(localized), &mh)
		if err := dec.Decode(&out); err != nil {
			log.Printf("l10n.Unpack Err: %v Value: %v", err, v)
			n[k] = v
			continue
		}
		//if err := msgpack.Unmarshal([]byte(localized), &out); err != nil {
		//	log.Printf("l10n.Unpack Err: %v Value: %v", err, v)
		//	n[k] = v
		//	continue
		//}
		n[k] = out
	}
	return n
}


// not sure if the rest below will ever be used...

type Localizable interface {
	LocalizeGet(string, string) string
}

type Localize struct {
	/*
	@description: Localized returns all fields that have localized content.
	This is raw content coming from a DB or somewhere else ...
	*/
	Localized func() map[string]string `gorm:"-"`
	SetValue func(fieldName string, value string) bool `gorm:"-"`
}


func init() {
}

// key: string is the struct field name,
// locale: is en-US, ar-AE, ...
func (l Localize) LocalizeGet(key string, locale string) string {
	resp := ""
	b, ok := l.Localized()[key]
	if !ok {
		return resp
	}
	/*
	{
		"en-US": "my Name",
		"de-DE": "mein Name",
	}
	*/
	var out map[string]string
	dec := codec.NewDecoderBytes([]byte(b), &mh)
	if err := dec.Decode(&out); err != nil {
		return resp
	}
	//if err := msgpack.Unmarshal([]byte(b), &out); err != nil {
	//	return resp
	//}
	resp, ok = out[locale]
	if !ok {
		return resp
	}

	return resp
}
// key: string is the struct field name,
// value: is the value to be set for that field name
// locale: is en-US, ar-AE, ...
func (l *Localize) LocalizeUpdate(key string, value string, locale string) bool {
	resp := false
	// #fieldName: > #locales: > #content
	fields := map[string]map[string]string{}

	for fieldName, encMap := range l.Localized() {
		var content map[string]string
		dec := codec.NewDecoderBytes([]byte(encMap), &mh)
		if err := dec.Decode(&content); err != nil {
			return resp
		}
		//if err := msgpack.Unmarshal([]byte(encMap), &content); err != nil {
		//	continue
		//}
		// if fieldName matches key replace the value at the given locale
		if fieldName == key {
			content[fieldName] = value
			resp = true
		}
		fields[fieldName] = content
	}

	// at this point we processed the data

	return resp
}
// key: string is the struct field name,
// content: map[string]string{"en-US": "Hello", "de-DE": "Hallo"}
func (l *Localize) LocalizeSet(key string, content map[string]string) bool {
	var (
		out []byte
	)
	enc := codec.NewEncoderBytes(&out, &mh)
	if err := enc.Encode(content); err != nil {
		return false
	}
	newValue := bytes.NewBuffer(out).String()

	//b, err := msgpack.Marshal(content)
	//if err != nil {
	//	return false
	//}
	//newValue := bytes.NewBuffer(b).String()
	return l.SetValue(key, newValue)
}

// ParseTagOption parse tag options to hash
func ParseTagOption(str string) map[string]string {
	tags := strings.Split(str, ";")
	setting := map[string]string{}
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}
	return setting
}

func IsLocalizable(scope *gorm.Scope) (IsLocalizable bool) {
	if scope.GetModelStruct().ModelType == nil {
		return false
	}
	//log.Printf("ModelStruct %v", scope.GetModelStruct().ModelType)
	_, IsLocalizable = reflect.New(scope.GetModelStruct().ModelType).Interface().(Localizable)
	return
}

func afterQuery(scope *gorm.Scope) {
	if IsLocalizable(scope) {
		log.Printf("L10N.afterQuery found localizable model")
	}
	log.Printf("L10N.afterQuery not localizable model")
}

// RegisterCallbacks register callback into GORM DB
func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()

	callback.Query().After("gorm:query").Register("l10n:after_query", afterQuery)
}
