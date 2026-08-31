package mystenbcs

// https://github.com/fardream/go-bcs/blob/main/bcs/decode.go#L19

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// Unmarshal unmarshals the bcs serialized data into v.
//
// Refer to notes in [Marshal] for details how data serialized/deserialized.
//
// During the unmarshalling process
//  1. if [Unmarshaler], use "UnmarshalBCS" method.
//  2. if not [Unmarshaler] but [Enum], use the specialization for [Enum].
//  3. otherwise standard process.
func Unmarshal(data []byte, v any) (int, error) {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

// Decoder takes an [io.Reader] and decodes value from it.
type Decoder struct {
	reader     io.Reader
	byteBuffer [1]byte
}

// NewDecoder creates a new [Decoder] from an [io.Reader]
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

// DecodeWithSize decodes a value from the decoder, and returns the number of bytes it consumed from the decoder.
//
//   - If the value is [Unmarshaler], the corresponding UnmarshalBCS will be called.
//   - If the value is [Enum], it will be special handled for [Enum]
func (d *Decoder) Decode(v any) (int, error) {
	reflectValue := reflect.ValueOf(v)
	if reflectValue.Kind() != reflect.Pointer || reflectValue.IsNil() {
		return 0, fmt.Errorf("not a pointer or nil pointer")
	}

	return d.decode(reflectValue)
}

// decode is the main lifter, it first checks if a value can be [reflect.Value.CanInterface],
// then checks if the value implements [Unmarshaler] or [Enum], and then switch on the kind of the value:
// - pointer, create a new one and decode into its element.
// - interface, decode into element.
// - function, channel, unsafe pointers, ignore
// - otherwise call [decodeVanilla].
func (d *Decoder) decode(v reflect.Value) (int, error) {
	// if v cannot interface, ignore
	if !v.CanInterface() {
		return 0, nil
	}

	// Unmarshaler
	if i, isUnmarshaler := v.Interface().(Unmarshaler); isUnmarshaler {
		return i.UnmarshalBCS(d.reader)
	}

	// Enum
	if _, isEnum := v.Interface().(Enum); isEnum {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			return d.decodeEnum(v.Elem())
		default:
			return d.decodeEnum(v)
		}
	}

	// switch kind
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return d.decode(v.Elem())

	case reflect.Interface:
		if v.IsNil() {
			return 0, fmt.Errorf("cannot decode into nil interface")
		}
		return d.decode(v.Elem())

	case reflect.Chan, reflect.Func, reflect.Uintptr, reflect.UnsafePointer:
		// silently ignore
		return 0, nil
	default:
		return d.decodeVanilla(v)
	}
}

// decodeVanilla decodes bool, ints, slice, struct, array, and string.
func (d *Decoder) decodeVanilla(v reflect.Value) (int, error) {
	kind := v.Kind()
	if !v.CanSet() {
		return 0, fmt.Errorf("cannot change value of kind %s", kind.String())
	}

	switch kind {
	case reflect.Bool:
		t, n, err := d.readByte()
		if err != nil {
			return n, err
		}

		if t == 0 {
			v.SetBool(false)
		} else {
			v.SetBool(true)
		}

		return n, nil

	case reflect.Int8, reflect.Uint8:
		return 1, binary.Read(d.reader, binary.LittleEndian, v.Addr().Interface())
	case reflect.Int16, reflect.Uint16:
		return 2, binary.Read(d.reader, binary.LittleEndian, v.Addr().Interface())
	case reflect.Int32, reflect.Uint32:
		return 4, binary.Read(d.reader, binary.LittleEndian, v.Addr().Interface())
	case reflect.Int64, reflect.Uint64:
		return 8, binary.Read(d.reader, binary.LittleEndian, v.Addr().Interface())

	case reflect.Struct:
		return d.decodeStruct(v)

	case reflect.Slice:
		sliceType := v.Type().Elem()
		if sliceType.Kind() == reflect.Uint8 {
			return d.decodeByteSlice(v)
		}

		return d.decodeSlice(v)

	case reflect.Array:
		arrayType := v.Type().Elem()
		if arrayType.Kind() == reflect.Uint8 {
			return d.decodeByteArray(v)
		}
		return d.decodeArray(v)

	case reflect.String:
		return d.decodeString(v)

	default:
		return 0, fmt.Errorf("unsupported vanilla decoding type: %s", kind.String())
	}
}

// decodeString
func (d *Decoder) decodeString(v reflect.Value) (int, error) {
	size, n, err := ULEB128Decode[int](d.reader)
	if err != nil {
		return n, err
	}

	if size == 0 {
		v.SetString("")
		return n, nil
	}

	tmp, err := d.readSized(size)
	if err != nil {
		return n, err
	}
	n += size

	v.SetString(string(tmp))

	return n, nil
}

// readSized reads exactly size bytes from the reader.
//
// It guards against a malicious length prefix: BCS length fields are
// attacker-controlled, and allocating make([]byte, size) up front lets a tiny
// payload (a 5-byte ULEB128 length of ~2GB) exhaust memory before any data is
// read — an unbounded-allocation DoS. When the reader can report how many bytes
// remain (bytes.Reader / bytes.Buffer, which the public Unmarshal always uses),
// reject an oversized size before allocating. For an opaque stream, read through
// an io.LimitReader with io.ReadFull so a short read is a clean error rather than
// a spurious "wrong number of bytes" and the buffer still can't outgrow size.
func (d *Decoder) readSized(size int) ([]byte, error) {
	if size < 0 {
		return nil, fmt.Errorf("negative size %d", size)
	}
	if lr, ok := d.reader.(interface{ Len() int }); ok && size > lr.Len() {
		return nil, fmt.Errorf("size %d exceeds %d remaining bytes", size, lr.Len())
	}
	tmp := make([]byte, size)
	if _, err := io.ReadFull(d.reader, tmp); err != nil {
		return nil, err
	}
	return tmp, nil
}

// readByte reads one byte from the input, error if no byte is read.
func (d *Decoder) readByte() (byte, int, error) {
	b := d.byteBuffer[:]
	n, err := d.reader.Read(b)
	if err != nil {
		return 0, n, err
	}
	if n == 0 {
		return 0, n, io.ErrUnexpectedEOF
	}

	return b[0], n, nil
}

func (d *Decoder) decodeStruct(v reflect.Value) (int, error) {
	t := v.Type()

	var n int

fieldLoop:
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue fieldLoop
		}
		tag, err := parseTagValue(t.Field(i).Tag.Get(tagName))
		if err != nil {
			return n, err
		}

		switch {
		case tag.isIgnored(): // ignored
			continue fieldLoop
		case tag.isOptional(): // optional
			isOptional, k, err := d.readByte()
			n += k
			if err != nil {
				return n, err
			}
			if isOptional == 0 {
				field.Set(reflect.Zero(field.Type()))
			} else {
				field.Set(reflect.New(field.Type().Elem()))
				k, err := d.decode(field.Elem())
				n += k
				if err != nil {
					return n, err
				}
			}
		default:
			k, err := d.decode(field)
			n += k
			if err != nil {
				return n, err
			}
		}
	}

	return n, nil
}

func (d *Decoder) decodeEnum(v reflect.Value) (int, error) {
	if v.Kind() != reflect.Struct {
		return 0, fmt.Errorf("only support struct for Enum, got %s", v.Kind().String())
	}
	enumId, n, err := ULEB128Decode[int](d.reader)
	if err != nil {
		return n, err
	}

	field := v.Field(enumId)

	// Handle interface{} fields (like None any in TransactionExpiration)
	// For empty enum variants, BCS only encodes the enum ID, no data follows
	if field.Kind() == reflect.Interface {
		if field.IsNil() {
			// For empty variants (like None), set to an empty struct as a marker
			// and return immediately without decoding (no data to decode)
			field.Set(reflect.ValueOf(struct{}{}))
			return n, nil
		}
		// If field is not nil, it means it's a non-empty variant, decode normally
	}

	k, err := d.decode(field)
	n += k

	return n, err
}

func (d *Decoder) decodeByteSlice(v reflect.Value) (int, error) {
	size, n, err := ULEB128Decode[int](d.reader)
	if err != nil {
		return n, err
	}

	if size == 0 {
		return n, nil
	}

	tmp, err := d.readSized(size)
	if err != nil {
		return n, err
	}
	n += size

	v.Set(reflect.ValueOf(tmp))

	return n, nil
}

func (d *Decoder) decodeByteArray(v reflect.Value) (int, error) {
	arraySize := v.Len()

	if arraySize == 0 {
		return 0, nil
	}

	tmp := make([]byte, arraySize)

	read, err := d.reader.Read(tmp)
	if err != nil {
		return read, err
	}

	if arraySize != read {
		return read, fmt.Errorf("wrong number of bytes read for [%d]byte, want: %d, got %d", arraySize, arraySize, read)
	}

	for i := 0; i < arraySize; i++ {
		v.Index(i).SetUint(uint64(tmp[i]))
	}

	return read, nil
}

func (d *Decoder) decodeArray(v reflect.Value) (int, error) {
	size := v.Len()
	t := v.Type()
	elementType := t.Elem()

	var n int
	if elementType.Kind() == reflect.Pointer {
		for i := 0; i < size; i++ {
			idx := reflect.New(elementType.Elem())
			k, err := d.decode(idx.Elem())
			n += k
			if err != nil {
				return n, err
			}
			v.Index(i).Set(idx)
		}
	} else {
		for i := 0; i < size; i++ {
			idx := reflect.New(elementType)
			k, err := d.decode(idx.Elem())
			n += k
			if err != nil {
				return n, err
			}
			v.Index(i).Set(idx.Elem())
		}
	}

	return n, nil
}

func (d *Decoder) decodeSlice(v reflect.Value) (int, error) {
	// get the length of the slice.
	size, n, err := ULEB128Decode[int](d.reader)
	if err != nil {
		return n, err
	}

	// element type of the slice
	elementType := v.Type().Elem()
	// make a new slice
	tmp := reflect.MakeSlice(v.Type(), 0, size)

	if elementType.Kind() == reflect.Pointer {
		for i := 0; i < size; i++ {
			ind := reflect.New(elementType.Elem())
			k, err := d.decode(ind)
			n += k
			if err != nil {
				return n, err
			}
			tmp = reflect.Append(tmp, ind)
		}
	} else {
		for i := 0; i < size; i++ {
			ind := reflect.New(elementType)
			k, err := d.decode(ind.Elem())
			n += k
			if err != nil {
				return n, err
			}
			tmp = reflect.Append(tmp, ind.Elem())
		}
	}

	v.Set(tmp)

	return n, nil
}
