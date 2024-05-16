package rom

import (
	"fmt"
	"strings"
	"testing"
)

// Encode a header from a structure (in json), then
// decode it from bytes.  Compare the result.
func TestHeaderRecode(t *testing.T) {
	h := &Header{
		PrgSize:        32 * 1024,
		ChrSize:        8 * 1024,
		MiscSize:       0,
		TrainerPresent: true,
		Mirroring:      M_HORIZONTAL,
		Nes2:           true,
		Mapper:         12,

		Nes2Mapper: 16,
		SubMapper:  4,

		Console: CT_EXTENDED,
	}

	raw := h.Bytes()
	if len(raw) != 16 {
		t.Errorf("Invalid header length of %d bytes", len(raw))
	}

	h2, err := ParseHeader(raw)
	if err != nil {
		t.Fatal(err)
	}

	beforeStr := []string{}
	for _, b := range raw {
		beforeStr = append(beforeStr, fmt.Sprintf("%02X", b))
	}
	t.Logf("h: %s", strings.Join(beforeStr, " "))

	afterStr := []string{}
	for _, b := range raw {
		afterStr = append(afterStr, fmt.Sprintf("%02X", b))
	}
	t.Logf("h: %s", strings.Join(afterStr, " "))

	equTest(t, h, h2)
}

func equTest(t *testing.T, a, b *Header) {
	t.Helper()

	if a == nil || b == nil {
		t.Fatalf("nil header object: %v vs %v", a, b)
	}

	if a.PrgSize != b.PrgSize {
		t.Errorf("PrgSize mismatch: %d vs %d", a.PrgSize, b.PrgSize)
	}

	if a.ChrSize != b.ChrSize {
		t.Errorf("ChrSize mismatch: %d vs %d", a.ChrSize, b.ChrSize)
	}

	if a.MiscSize != b.MiscSize {
		t.Errorf("MiscSize mismatch: %d vs %d", a.MiscSize, b.MiscSize)
	}

	if a.TrainerPresent != b.TrainerPresent {
		t.Errorf("TrainerPresent mismatch: %t vs %t", a.TrainerPresent, b.TrainerPresent)
	}

	if a.PersistentMemory != b.PersistentMemory {
		t.Errorf("PersistentMemory mismatch: %t vs %t", a.PersistentMemory, b.PersistentMemory)
	}

	if a.Mirroring != b.Mirroring {
		t.Errorf("Mirroring mismatch: %s vs %s", a.Mirroring, b.Mirroring)
	}

	if a.Console != b.Console {
		t.Errorf("Console mismatch: %s vs %s", a.Console, b.Console)
	}

	if a.Nes2 != b.Nes2 {
		t.Errorf("Nes2 mismatch: %t vs %t", a.Nes2, b.Nes2)
	}

	if a.Mapper != b.Mapper {
		t.Errorf("Mapper mismatch: %d vs %d", a.Mapper, b.Mapper)
	}

	t.Log(a.Debug())
	t.Log(b.Debug())
}
