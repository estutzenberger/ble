package att

import (
	"bytes"
	"io"
	"log"

	"github.com/currantlabs/bt/uuid"

	"golang.org/x/net/context"
)

// Attribute is a BLE attribute.
type Attribute struct {
	Handle       uint16    // Handle
	EndingHandle uint16    // Group EndingHandle
	Type         uuid.UUID // Type (in UUID)

	Value []byte // Staic and read-only Value
	Pvt   Value  // Handler from upper layer, such as GATT.
}

// Value ...
type Value interface {
	Handle(ctx context.Context, req []byte, resp *ResponseWriter) Error
}

// A Range is a contiguous range of attributes.
type Range struct {
	Attributes []Attribute
	Base       uint16 // handle for first Attribute in Attributes
}

const (
	tooSmall = -1
	tooLarge = -2
)

// idx returns the index into Attributes corresponding to Attribute a.
// If h is too small, idx returns tooSmall (-1).
// If h is too large, idx returns tooLarge (-2).
func (r *Range) idx(h int) int {
	if h < int(r.Base) {
		return tooSmall
	}
	if int(h) >= int(r.Base)+len(r.Attributes) {
		return tooLarge
	}
	return h - int(r.Base)
}

// At returns Attribute a.
func (r *Range) at(h uint16) (a Attribute, ok bool) {
	i := r.idx(int(h))
	if i < 0 {
		return Attribute{}, false
	}
	return r.Attributes[i], true
}

// Subrange returns attributes in range [start, end]; it may return an empty
// slice. Subrange does not panic for out-of-range start or end.
func (r *Range) subrange(start, end uint16) []Attribute {
	startidx := r.idx(int(start))
	switch startidx {
	case tooSmall:
		startidx = 0
	case tooLarge:
		return []Attribute{}
	}

	endidx := r.idx(int(end) + 1) // [start, end] includes its upper bound!
	switch endidx {
	case tooSmall:
		return []Attribute{}
	case tooLarge:
		endidx = len(r.Attributes)
	}
	return r.Attributes[startidx:endidx]
}

// DumpAttributes ...
func DumpAttributes(Attributes []Attribute) {
	log.Printf("Generating attribute table:")
	log.Printf("handle\tend\ttype\tvalue")
	for _, a := range Attributes {
		if a.Value != nil {
			log.Printf("0x%04X\t0x%04X\t0x%s\t[ % X ]", a.Handle, a.EndingHandle, a.Type, a.Value)
			continue
		}
		log.Printf("0x%04X\t0x%04X\t0x%s\t%T", a.Handle, a.EndingHandle, a.Type, a.Pvt)
	}
}

// ResponseWriter ...
type ResponseWriter struct {
	buf *bytes.Buffer
}

// Reset ...
func (r *ResponseWriter) Reset() {
	r.buf.Reset()
}

// Len ...
func (r *ResponseWriter) Len() int {
	return r.buf.Len()
}

// Cap ...
func (r *ResponseWriter) Cap() int {
	return r.buf.Cap()
}

// Write writes data to return as the characteristic value.
func (r *ResponseWriter) Write(b []byte) (int, error) {
	if len(b) > r.buf.Cap()-r.buf.Len() {
		return 0, io.ErrShortWrite
	}

	return r.buf.Write(b)
}
