package debug

import (
	"bytes"
	"fmt"
	r "reflect"
	"time"
)

func Dump(value interface{}) string {
	r.TypeOf(value)
	v := r.ValueOf(value)
	k := v.Kind()
	for k == r.Ptr {
		v = v.Elem()
		k = v.Kind()
	}

	buf := &bytes.Buffer{}

	if v.IsValid() {
		dumpValue(v, "", buf, true)
	} else {
		fmt.Fprint(buf, v.String())
	}
	result := buf.String()
	return result[:len(result)-1]
}

var timeType = r.TypeOf(time.Time{})

func dumpValue(v r.Value, prefix string, buf *bytes.Buffer, withType bool) {
	if v.Type() == timeType {
		t := v.Interface().(time.Time)
		fmt.Fprintf(buf, "%v", t.Format("time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),"))
		return
	}

	k := v.Kind()

	switch k {
	case r.String:
		fmt.Fprintf(buf, "%q,", v)

	case r.Bool:
		fmt.Fprintf(buf, "%v,", v)

	case r.Int, r.Int8, r.Int16, r.Int32, r.Int64, r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
		fmt.Fprintf(buf, "%d,", v.Int())

	case r.Float32, r.Float64, r.Complex64, r.Complex128:
		fmt.Fprintf(buf, "%g,", v.Float())

	case r.Array, r.Slice:
		dumpSlice(v, prefix, buf, withType)

	case r.Map:
		dumpMap(v, prefix, buf, withType)

	case r.Struct:
		dumpStruct(v, prefix, buf, true)

	case r.Ptr:
		if v.IsNil() {
			buf.WriteString("nil,")
		} else {
			buf.WriteString("&")
			dumpValue(v.Elem(), prefix, buf, withType)
		}

	case r.Interface:
		if v.IsNil() {
			buf.WriteString("nil,")
		} else {
			elem := v.Elem()
			dumpValue(elem, prefix, buf, true)
		}

	case r.Chan, r.Func, r.UnsafePointer:
		// Skip those

	case r.Invalid:
		fmt.Fprintf(buf, " INVALID KIND: %v,", v)

	default:
		fmt.Fprintf(buf, "Unsupported kind: %v", k)
	}
}

func dumpStruct(v r.Value, prefix string, buf *bytes.Buffer, withType bool) {
	t := v.Type()
	if withType {
		fmt.Fprintf(buf, "%s{", t)
	} else {
		fmt.Fprintf(buf, "{")
	}
	for fn := 0; fn < v.NumField(); fn++ {
		ft := t.Field(fn)
		fmt.Fprintf(buf, "\n%s\t%s: ", prefix, ft.Name)
		dumpValue(v.Field(fn), prefix+"\t", buf, true)
	}
	fmt.Fprintf(buf, "\n%s},", prefix)
}

func dumpMap(v r.Value, prefix string, buf *bytes.Buffer, withType bool) {
	keys := v.MapKeys()
	if withType {
		fmt.Fprintf(buf, "%s{", v.Type())
	} else {
		fmt.Fprintf(buf, "{")
	}
	for _, key := range keys {
		value := v.MapIndex(key)
		fmt.Fprintf(buf, "\n%s\t%q: ", prefix, key)
		dumpValue(value, prefix+"\t", buf, false)
	}
	if len(keys) > 0 {
		fmt.Fprintf(buf, "\n%s", prefix)
	}

	fmt.Fprint(buf, "},")
}

func dumpSlice(v r.Value, prefix string, buf *bytes.Buffer, withType bool) {
	if v.IsNil() {
		buf.WriteString("nil,")
		return
	}
	if withType {
		fmt.Fprintf(buf, "%s{", v.Type())
	} else {
		fmt.Fprintf(buf, "{")
	}

	for idx := 0; idx < v.Len(); idx++ {
		fmt.Fprintf(buf, "\n\t%s", prefix)
		dumpValue(v.Index(idx), prefix+"\t", buf, false)
	}
	if v.Len() > 0 {
		fmt.Fprintf(buf, "\n%s", prefix)
	}
	buf.WriteString("},")
}
