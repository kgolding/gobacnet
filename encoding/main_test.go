package encoding

import "testing"

func TestNPDU(t *testing.T) {
	n := encodeNPDU(false, Normal)
	_, err := EncodePDU(&n, &BacnetAddress{}, &BacnetAddress{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadProperty(t *testing.T) {

}

func TestSegsApduEncode(t *testing.T) {
	// Test is structured as parameter 1, parameter 2, output
	tests := [][]int{
		[]int{0, 1, 0},
		[]int{64, 60, 0x61},
		[]int{80, 205, 0x72},
		[]int{80, 405, 0x73},
		[]int{80, 1005, 0x74},
		[]int{3, 1035, 0x15},
		[]int{9, 1035, 0x35},
	}

	for _, test := range tests {
		d := int(encodeMaxSegsMaxApdu(test[0], test[1]))
		if d != test[2] {
			t.Fatalf("Input was Segments %d and Apdu %d: Expected %x got %x", test[0], test[1], test[2], d)
		}
	}
}

func TestObject(t *testing.T) {
	e := NewEncoder()
	var inObjectType uint16 = 17
	var inInstance uint32 = 23
	e.objectId(inObjectType, inInstance)
	b := e.Bytes()
	t.Log(b)

	d := NewDecoder(b)
	outObject, outInstance := d.objectId()

	if inObjectType != outObject {
		t.Fatalf("There was an issue encoding/decoding objectType. Input value was %d and output value was %d", inObjectType, outObject)
	}

	if inInstance != outInstance {
		t.Fatalf("There was an issue encoding/decoding objectType. Input value was %d and output value was %d", inInstance, outInstance)
	}

	if err := d.Error(); err != nil {
		t.Fatal(err)
	}
}

func TestEnumerated(t *testing.T) {
	lengths := []int{size8, size16, size24, size32, size32}
	tests := []uint32{1, 2 << 8, 3 << 17, 7 << 25, 8 << 26}
	e := NewEncoder()
	for _, val := range tests {
		e.enumerated(val)
	}
	b := e.Bytes()
	d := NewDecoder(b)
	for i, val := range tests {
		x := d.enumerated(lengths[i])
		if x != val {
			t.Fatalf("Test[%d]:Decoded value %d doesn't match encoded value %d", i+1, x, val)
		}
	}

	d = NewDecoder(b)
	// 1000 is not a valid length
	x := d.enumerated(1000)
	if x != 0 {
		t.Fatalf("For invalid lengths, the value 0 should be decoded. The value %d was decoded", x)
	}
}

const compareErrFmt = "Mismatch in %s when decoding values. Expected: %d, recieved: %d"

func compare(t *testing.T, name string, a uint, b uint) {
	// See if the initial read property data matches the output read property
	if a != b {
		t.Fatal(compareErrFmt, name, a, b)
	}
}

func TestReadingProperty(t *testing.T) {
	e := NewEncoder()
	rd := ReadPropertyData{
		ObjectType:     37,
		ObjectInstance: 1000,
		ObjectProperty: 3921,
		ArrayIndex:     0,
	}
	e.readProperty(10, rd)
	if err := e.Error(); err != nil {
		t.Fatal(err)
	}

	b := e.Bytes()
	d := NewDecoder(b)

	// Read Property reads 4 extra fields that are not original encoded. Need to
	//find out where these 4 fields come from
	d.buff.Read(make([]uint8, 4))
	err, outRd := d.readProperty()
	if err != nil {
		t.Fatal(err)
	}

	// See if the initial read property data matches the output read property
	compare(t, "object instance", uint(rd.ObjectInstance), uint(outRd.ObjectInstance))
	compare(t, "boject type", uint(rd.ObjectType), uint(outRd.ObjectType))
	compare(t, "object property", uint(rd.ObjectProperty), uint(outRd.ObjectProperty))
	compare(t, "array index", uint(rd.ArrayIndex), uint(outRd.ArrayIndex))
}
