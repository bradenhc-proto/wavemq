package wavemq

import (
	"testing"
)

func TestEncodeRemainingLength(t *testing.T) {
	// Test only one byte
	var testLength uint32 = 34
	buf := encodeRemainingLength(testLength)
	resultLength, err := decodeRemainingLength(buf)
	if err != nil {
		t.Errorf("An error occurred while decoding the length: %v", err)
	}
	if testLength != resultLength {
		t.Errorf("The result was incorrect. Test length was %v and result was %v", testLength, resultLength)
	}

	// Test multiple bytes
	testLength = 1234567
	buf = encodeRemainingLength(testLength)
	resultLength, err = decodeRemainingLength(buf)
	if err != nil {
		t.Errorf("An error occurred while decoding the length: %v", err)
	}
	if testLength != resultLength {
		t.Errorf("The result was incorrect. Test length was %v and result was %v", testLength, resultLength)
	}

	// Test too many bytes
	testLength = 999999999
	buf = encodeRemainingLength(testLength)
	_, err = decodeRemainingLength(buf)
	if err == nil {
		t.Errorf("There should have been a malformed length error because it was too big")
	}

}
