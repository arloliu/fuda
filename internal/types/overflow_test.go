package types

import (
	"reflect"
	"strings"
	"testing"
)

func TestConvert_OverflowAndSize(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		target    any
		wantErr   bool
		errMatch  string
		wantValue any
	}{
		// --- Int8 ---
		{
			name:      "int8 - valid",
			value:     "127",
			target:    new(int8),
			wantValue: int8(127),
		},
		{
			name:     "int8 - overflow upper",
			value:    "128",
			target:   new(int8),
			wantErr:  true,
			errMatch: "overflows int8",
		},
		{
			name:     "int8 - overflow lower",
			value:    "-129",
			target:   new(int8),
			wantErr:  true,
			errMatch: "overflows int8",
		},
		{
			name:     "int8 - valid size", // "1KiB" = 1024 -> overflow int8
			value:    "1KiB",
			target:   new(int8),
			wantErr:  true,
			errMatch: "overflows int8",
		},

		// --- Int ---
		{
			name:      "int - valid size",
			value:     "1KiB",
			target:    new(int),
			wantValue: int(1024),
		},
		{
			name:      "int - valid numeric",
			value:     "123456",
			target:    new(int),
			wantValue: int(123456),
		},
		{
			name:      "int - valid SI",
			value:     "1KB",
			target:    new(int),
			wantValue: int(1000),
		},

		// --- Uint8 ---
		{
			name:      "uint8 - valid",
			value:     "255",
			target:    new(uint8),
			wantValue: uint8(255),
		},
		{
			name:     "uint8 - overflow",
			value:    "256",
			target:   new(uint8),
			wantErr:  true,
			errMatch: "overflows uint8",
		},
		{
			name:     "uint8 - negative",
			value:    "-1",
			target:   new(uint8),
			wantErr:  true,
			errMatch: "cannot assign negative",
		},

		// --- Uint16 ---
		{
			name:      "uint16 - almost overflow",
			value:     "65535",
			target:    new(uint16),
			wantValue: uint16(65535),
		},
		{
			name:     "uint16 - overflow",
			value:    "65536",
			target:   new(uint16),
			wantErr:  true,
			errMatch: "overflows uint16",
		},
		{
			name:     "uint16 - 64KiB", // 65536 -> overflow uint16 (max 65535)
			value:    "64KiB",
			target:   new(uint16),
			wantErr:  true,
			errMatch: "overflows uint16",
		},
		{
			name:      "uint16 - 63KiB", // 64512 -> ok
			value:     "63KiB",
			target:    new(uint16),
			wantValue: uint16(64512),
		},

		// --- Uint64 special ---
		{
			name:      "uint64 - valid large size", // 4GiB
			value:     "4GiB",
			target:    new(uint64),
			wantValue: uint64(4294967296),
		},
		{
			name:      "uint64 - max raw",
			value:     "18446744073709551615",
			target:    new(uint64),
			wantValue: uint64(18446744073709551615),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetVal := reflect.ValueOf(tt.target).Elem()
			err := Convert(tt.value, targetVal)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMatch)
				} else if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("expected error containing %q, got %q", tt.errMatch, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if targetVal.Interface() != tt.wantValue {
					t.Errorf("expected value %v, got %v", tt.wantValue, targetVal.Interface())
				}
			}
		})
	}
}
