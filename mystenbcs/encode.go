package mystenbcs

// https://github.com/fardream/go-bcs/blob/main/bcs/encode.go

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// Marshaler customizes the marshalling behavior for a type
type Marshaler interface {
	MarshalBCS() ([]byte, error)
}

// Unmarshaler customizes the unmarshalling behavior for a type.
//
// Compared with other Unmarshalers in golang, the Unmarshaler here takes
// a [io.Reader] instead of []byte, since it is difficult to delimit the byte streams without unmarshalling.
// Method [UnmarshalBCS] returns the number of bytes read, and potentially an error.
type Unmarshaler interface {
	UnmarshalBCS(io.Reader) (int, error)
}

type Enum interface {
	// IsBcsEnum doesn't do anything. Its function is to indicate this is an enum for bcs de/serialization.
	IsBcsEnum()
}

// Encoder takes an [io.Writer] and encodes value into it.
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a new [Encoder] from an [io.Writer]
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

// Encode a value v into the encoder.
//
//   - If the value is [Marshaler], the corresponding
//     MarshalBCS implementation will be called.
//   - If the value is [Enum], it will be special handled for [Enum].
func (e *Encoder) Encode(v any) error {
	return e.encode(reflect.ValueOf(v))
}

// encode a value
func (e *Encoder) encode(v reflect.Value) error {
	// if v not CanInterface,
	// this value is an unexported value, skip it.
	if !v.IsValid() || !v.CanInterface() {
		return nil
	}

	// test for the two interfaces we defined.
	// 1. Marshaler
	// 2. Enum.
	i := v.Interface()
	if m, ismarshaler := i.(Marshaler); ismarshaler {
		bytes, err := m.MarshalBCS()
		if err != nil {
			return err
		}

		_, err = e.w.Write(bytes)

		return err
	}
	if _, isenum := i.(Enum); isenum {
		return e.encodeEnum(reflect.Indirect(v))
	}

	kind := v.Kind()

	switch kind {
	case reflect.Bool, // boolean
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, // all the ints
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // all the uints
		// use little endian to encode those.
		return binary.Write(e.w, binary.LittleEndian, v.Interface())

	case reflect.Pointer: // pointer
		// bcs has no encoding for an absent value outside of an Option, so a nil
		// pointer here would contribute no bytes at all and shift everything that
		// follows it. the optional flag is what turns a pointer into an Option, and
		// that is checked by encodeStruct before the field reaches this point.
		if v.IsNil() {
			return fmt.Errorf("nil pointer to %s cannot be encoded, consider marking the field optional to encode it as an Option", v.Type().Elem().String())
		}
		return e.encode(reflect.Indirect(v))

	case reflect.Interface:
		return e.encode(v.Elem())

	case reflect.Slice: // slices
		// check if the element is uint8 or byteslice
		if byteSlice, ok := (v.Interface()).([]byte); ok {
			return e.encodeByteSlice(byteSlice)
		}
		return e.encodeSlice(v)

	case reflect.Array: // encode array
		return e.encodeArray(v)

	case reflect.String:
		str := []byte(v.String())
		return e.encodeByteSlice(str)

	case reflect.Struct:
		return e.encodeStruct(v)

	case reflect.Chan, reflect.Func, reflect.Uintptr, reflect.UnsafePointer: // channel, func, pointers
		return nil

	default:
		return fmt.Errorf("unsupported kind: %s, consider make the field ignored by using - tag or provide a customized Marshaler implementation", kind.String())
	}
}

// encodeEnum encodes an [Enum]
func (e *Encoder) encodeEnum(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// ignore fields that are not exported
		if !field.CanInterface() {
			continue
		}

		fieldType := t.Field(i)
		// check the tag
		tag, err := parseTagValue(fieldType.Tag.Get(tagName))
		if err != nil {
			return err
		}
		if tag.isIgnored() {
			continue
		}
		fieldKind := field.Kind()
		if fieldKind != reflect.Pointer && fieldKind != reflect.Interface {
			return fmt.Errorf("enum only supports fields that are either pointers or interfaces, unless they are ignored")
		}
		if !field.IsNil() {
			if _, err := e.w.Write(ULEB128Encode(i)); err != nil {
				return err
			}
			if fieldKind == reflect.Pointer {
				return e.encode(reflect.Indirect(field))
			} else {
				// the variant index is already written, only the value the
				// interface holds is left. an empty variant holds an empty
				// struct and adds no further bytes.
				return e.encode(field)
			}
		}
	}

	return fmt.Errorf("no field is set in the enum")
}

// encodeByteSlice is specialized since bytes those can be simply put into the output.
func (e *Encoder) encodeByteSlice(b []byte) error {
	l := len(b)
	if _, err := e.w.Write(ULEB128Encode(l)); err != nil {
		return err
	}

	if _, err := e.w.Write(b); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) encodeArray(v reflect.Value) error {
	length := v.Len()
	for i := 0; i < length; i++ {
		if err := e.encode(v.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeSlice(v reflect.Value) error {
	length := v.Len()
	if _, err := e.w.Write(ULEB128Encode(length)); err != nil {
		return err
	}

	for i := 0; i < length; i++ {
		if err := e.encode(v.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// if a field is not exported, ignore
		if !field.CanInterface() {
			continue
		}
		tag, err := parseTagValue(t.Field(i).Tag.Get(tagName))
		if err != nil {
			return err
		}
		switch {
		case tag.isIgnored():
			continue
		case tag.isOptional():
			if field.Kind() != reflect.Pointer && field.Kind() != reflect.Interface {
				return fmt.Errorf("optional field can only be pointer or interface")
			}
			if field.IsNil() {
				_, err := e.w.Write([]byte{0})
				if err != nil {
					return err
				}
			} else {
				if _, err := e.w.Write([]byte{1}); err != nil {
					return err
				}
				if err := e.encode(field.Elem()); err != nil {
					return err
				}
			}
			continue
		default:
			// finally
			if err := e.encode(field); err != nil {
				return err
			}
		}
	}

	return nil
}

// Marshal a value into bcs bytes.
//
// Many constructs supported by bcs don't exist in golang or move-lang.
//
//   - [Enum] is used to simulate the effects of rust enum.
//   - Use tag `optional` to indicate an optional value in rust.
//     the field must be pointer or interface.
//   - Use tag `-` to ignore fields.
//   - Unexported fields are ignored.
//
// Note that bcs doesn't have schema, and field names are irrelevant. The fields
// of struct are serialized in the order that they are defined.
//
// Pointers are serialized as the type they point to. Nil pointers are an error
// unless they are marked as `optional`, in which case they are serialized as the
// `None` variant of an Option.
//
// Arrays are serialized as fixed length vector (or serialize the each object individually without prefixing
// the length of the array).
//
// Vanilla maps are not supported, however, the code will error if map is encountered to call out they are
// not supported and either ignore or provide a customized marshal function.
//
// Channels, functions are silently ignored.
//
// During marshalling process, how v is marshalled depends on if v implemented [Marshaler] or [Enum]
//  1. if [Marshaler], use "MarshalBCS" method.
//  2. if not [Marshaler] but [Enum], use specialization for [Enum].
//  3. otherwise standard process.
func Marshal(v any) ([]byte, error) {
	var b bytes.Buffer
	e := NewEncoder(&b)

	if err := e.Encode(v); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

type Option[T any] struct {
	Some T
	None bool
}

func (p *Option[T]) MarshalBCS() ([]byte, error) {
	if p.None {
		return []byte{0}, nil
	}
	b, err := Marshal(p.Some)
	return append([]byte{1}, b...), err
}

func (p *Option[T]) UnmarshalBCS(r io.Reader) (int, error) {
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	tmp := buf.Bytes()
	if len(tmp) == 1 {
		p.None = true
		return 1, nil
	}
	b := tmp[1:]
	return Unmarshal(b, &p.Some)
}

// MustMarshal [Marshal] v, and panics if error.
func MustMarshal(v any) []byte {
	result, err := Marshal(v)
	if err != nil {
		panic(err)
	}

	return result
}

func fixedByteArrayToSlice(v reflect.Value) []byte {
	if v.Kind() != reflect.Array {
		return nil
	}
	if v.Type().Elem().Kind() != reflect.Uint8 {
		return nil
	}

	// If the value is Sui address bytes, not to add uleb128 prefix
	if isCustomSuiAddressBytes(v) {
		return nil
	}

	length := v.Len()
	slice := make([]byte, length)
	for i := 0; i < length; i++ {
		slice[i] = byte(v.Index(i).Uint())
	}
	return slice
}

const SuiAddressBytesName = "SuiAddressBytes"

func isCustomSuiAddressBytes(v reflect.Value) bool {
	return v.Type().Name() == SuiAddressBytesName
}
