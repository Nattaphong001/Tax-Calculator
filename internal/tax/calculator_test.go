package tax

import (
	"math"
	"testing"
)

func TestCalculateTax(t *testing.T) {
	tests := []struct {
		name      string
		netIncome float64
		want      float64
	}{
		{"zero", 0, 0},
		{"negative treated as zero", -1000, 0},
		{"exactly top of 0% bracket", 150000, 0},
		{"single bracket", 290000, 14000},
		{"two brackets", 540000, 41000},
		{"example 1 in the brief", 690000, 63500},
		{"three brackets", 1140000, 138000},
		{"top of 20% bracket", 2000000, 310000},
		{"into 35% bracket", 2500000, 485000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateTax(tc.netIncome)
			// เทียบ float ด้วย tolerance ไม่ใช้ == เพราะ 0.1 และ 0.15 เก็บเป็นเลขฐานสองไม่ได้พอดี
			if math.Abs(got-tc.want) > 0.01 {
				t.Errorf("CalculateTax(%v) = %v, want %v", tc.netIncome, got, tc.want)
			}
		})
	}
}