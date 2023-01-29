package businesslogic

import (
	"io/ioutil"
	"reflect"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type timeDecodeExtension struct {
	jsoniter.DummyExtension
}

type decorateCodecTime struct {
	originDecoder jsoniter.ValDecoder
}

func (codec *decorateCodecTime) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	tStr := iter.ReadString()
	t, err := time.Parse("2006-01-02", tStr)
	if err != nil {
		codec.originDecoder.Decode(ptr, iter)
	} else {
		*((*time.Time)(ptr)) = t
	}

}

func (e *timeDecodeExtension) DecorateDecoder(typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	kind := typ.Kind()
	var t time.Time
	switch kind {
	case reflect.Struct:
		if typ == reflect2.TypeOf(t) {
			return &decorateCodecTime{decoder}
		}
	}
	return decoder
}

func ReadStruct[structType any](filename string) (*structType, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var s structType

	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
func MinDate(dates ...time.Time) time.Time {
	min := time.Time{}
	for _, d := range dates {
		if !d.IsZero() && (min.IsZero() || min.After(d)) {
			min = d
		}
	}
	return min
}

func MaxDate(dates ...time.Time) time.Time {
	min := time.Time{}
	for i, d := range dates {
		if min.Before(d) || i == 0 {
			min = d
		}
	}
	return min
}
