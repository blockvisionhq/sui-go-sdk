package mystenbcs

import (
	"io"
	"runtime"
	"testing"
)

// A tiny payload whose ULEB128 length prefix claims ~2GB, followed by no data.
// ULEB128 for 0x7FFFFFFF is FF FF FF FF 07.
var oversizedLenPrefix = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x07}

func TestDecodeStringRejectsOversizedLengthWithoutAllocating(t *testing.T) {
	var s string
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, err := Unmarshal(oversizedLenPrefix, &s)
	runtime.ReadMemStats(&after)

	allocated := int64(after.TotalAlloc - before.TotalAlloc)
	if err == nil {
		t.Fatal("expected an error for a length exceeding the input")
	}
	if allocated > 10*1024*1024 {
		t.Fatalf("allocated %d bytes for a 5-byte payload (unbounded-allocation DoS)", allocated)
	}
}

func TestDecodeByteSliceRejectsOversizedLengthWithoutAllocating(t *testing.T) {
	var b []byte
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, err := Unmarshal(oversizedLenPrefix, &b)
	runtime.ReadMemStats(&after)

	allocated := int64(after.TotalAlloc - before.TotalAlloc)
	if err == nil {
		t.Fatal("expected an error for a length exceeding the input")
	}
	if allocated > 10*1024*1024 {
		t.Fatalf("allocated %d bytes for a 5-byte payload (unbounded-allocation DoS)", allocated)
	}
}

func TestDecodeStringRoundTrip(t *testing.T) {
	want := "hello, sui"
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got string
	if _, err := Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != want {
		t.Fatalf("round-trip mismatch: got %q, want %q", got, want)
	}
}

// oneByteReader returns data one byte per Read call — a valid io.Reader that a
// single r.Read cannot fully satisfy but io.ReadFull can.
type oneByteReader struct {
	data []byte
	pos  int
}

func (o *oneByteReader) Read(p []byte) (int, error) {
	if o.pos >= len(o.data) {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = o.data[o.pos]
	o.pos++
	return 1, nil
}

func TestDecodeStringChunkedReader(t *testing.T) {
	// Regression: decodeString used r.Read (may return < size bytes) instead of
	// io.ReadFull, so a valid chunked reader failed with "wrong number of bytes".
	data, err := Marshal("chunked")
	if err != nil {
		t.Fatal(err)
	}
	var got string
	if _, err := NewDecoder(&oneByteReader{data: data}).Decode(&got); err != nil {
		t.Fatalf("chunked decode failed: %v", err)
	}
	if got != "chunked" {
		t.Fatalf("chunked decode mismatch: got %q, want %q", got, "chunked")
	}
}
