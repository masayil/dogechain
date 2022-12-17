package types

import (
	"bytes"
	"testing"
)

func TestStringToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		arr []byte
		exp []byte
	}{
		{StringToBytes("0x00ffff00ff0000"), []byte{0x00, 0xff, 0xff, 0x00, 0xff, 0x00, 0x00}},
		{StringToBytes("0x00000000000000"), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{StringToBytes("0xff"), []byte{0xff}},
		{[]byte{}, []byte{}},
		{StringToBytes("0x00ffffffffffff"), []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	for i, test := range tests {
		if !bytes.Equal(test.arr, test.exp) {
			t.Errorf("test %d, got %x exp %x", i, test.arr, test.exp)
		}
	}
}

func TestTrimLeftZeroes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		arr []byte
		exp []byte
	}{
		{StringToBytes("0x00ffff00ff0000"), StringToBytes("0xffff00ff0000")},
		{StringToBytes("0x00000000000000"), []byte{}},
		{StringToBytes("0xff"), StringToBytes("0xff")},
		{[]byte{}, []byte{}},
		{StringToBytes("0x00ffffffffffff"), StringToBytes("0xffffffffffff")},
	}

	for i, test := range tests {
		got := TrimLeftZeroes(test.arr)
		if !bytes.Equal(got, test.exp) {
			t.Errorf("test %d, got %x exp %x", i, got, test.exp)
		}
	}
}

func TestTrimRightZeroes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		arr []byte
		exp []byte
	}{
		{StringToBytes("0x00ffff00ff0000"), StringToBytes("0x00ffff00ff")},
		{StringToBytes("0x00000000000000"), []byte{}},
		{StringToBytes("0xff"), StringToBytes("0xff")},
		{[]byte{}, []byte{}},
		{StringToBytes("0x00ffffffffffff"), StringToBytes("0x00ffffffffffff")},
	}

	for i, test := range tests {
		got := TrimRightZeroes(test.arr)
		if !bytes.Equal(got, test.exp) {
			t.Errorf("test %d, got %x exp %x", i, got, test.exp)
		}
	}
}
