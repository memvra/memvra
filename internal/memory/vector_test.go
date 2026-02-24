package memory

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestFloat32SliceToBlob(t *testing.T) {
	input := []float32{1.0, 2.0, 3.0}
	blob := float32SliceToBlob(input)

	if len(blob) != 12 { // 3 floats * 4 bytes each
		t.Fatalf("expected 12 bytes, got %d", len(blob))
	}

	// Verify first float.
	bits := binary.LittleEndian.Uint32(blob[0:4])
	val := math.Float32frombits(bits)
	if val != 1.0 {
		t.Errorf("first float: got %f, want 1.0", val)
	}
}

func TestBlobToFloat32Slice(t *testing.T) {
	original := []float32{1.5, -2.5, 3.14}
	blob := float32SliceToBlob(original)
	result := BlobToFloat32Slice(blob)

	if len(result) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(result), len(original))
	}
	for i, v := range result {
		if v != original[i] {
			t.Errorf("index %d: got %f, want %f", i, v, original[i])
		}
	}
}

func TestFloat32RoundTrip(t *testing.T) {
	input := []float32{0.0, -1.0, 1e-10, 1e10, math.MaxFloat32}
	blob := float32SliceToBlob(input)
	output := BlobToFloat32Slice(blob)

	for i := range input {
		if input[i] != output[i] {
			t.Errorf("round-trip failed at index %d: %f != %f", i, input[i], output[i])
		}
	}
}

func TestFloat32SliceToBlob_Empty(t *testing.T) {
	blob := float32SliceToBlob(nil)
	if len(blob) != 0 {
		t.Errorf("expected empty blob for nil input, got %d bytes", len(blob))
	}
}

func TestBlobToFloat32Slice_Empty(t *testing.T) {
	result := BlobToFloat32Slice(nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice for nil blob, got %d elements", len(result))
	}
}
