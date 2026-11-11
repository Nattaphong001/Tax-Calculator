package tax

import (
	"math"
	"testing"
)

func TestCalculateDetailedBySum(t *testing.T) {
	tests := []struct {
		name      string
		netIncome float64 // เงินได้สุทธิ (หลังหักลดหย่อน)
		want      float64 // ภาษีรวมที่คำนวณมือไว้แล้ว
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
			// ฟังก์ชันรับรายได้รวม จึงบวกลดหย่อนกลับเข้าไปให้ได้เงินได้สุทธิตามต้องการ
			got := CalculateDetailed(tc.netIncome+PersonalAllowance, 0)

			sum := 0.0
			for _, l := range got.TaxLevel {
				sum += l.Tax
			}
			// เทียบ float ด้วย tolerance ไม่ใช้ == เพราะ 0.1 และ 0.15 เก็บเป็นเลขฐานสองไม่ได้พอดี
			if math.Abs(sum-tc.want) > 0.01 {
				t.Errorf("sum of TaxLevel = %v, want %v", sum, tc.want)
			}
			if math.Abs(got.Tax-tc.want) > 0.01 {
				t.Errorf("Tax = %v, want %v", got.Tax, tc.want)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name        string
		totalIncome float64
		wht         float64
		want        float64
	}{
		{"no tax, no wht", 60000, 0, 0},
		{"example 1 in the brief", 750000, 0, 63500},
		{"example 2 with wht", 600000, 15000, 26000},
		{"wht reduces tax", 450000, 8000, 16000},
		{"wht more than tax gives refund", 100000, 5000, -5000},
		{"wht equals tax", 600000, 41000, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Calculate(tc.totalIncome, tc.wht)
			if math.Abs(got-tc.want) > 0.01 {
				t.Errorf("Calculate(%v, %v) = %v, want %v", tc.totalIncome, tc.wht, got, tc.want)
			}
		})
	}
}

func TestCalculateDetailed(t *testing.T) {
	got := CalculateDetailed(750000, 0)

	wantLevels := []Level{
		{"0-150,000", 0},
		{"150,001-500,000", 35000},
		{"500,001-1,000,000", 28500},
		{"1,000,001-2,000,000", 0},
		{"2,000,001 ขึ้นไป", 0},
	}

	if math.Abs(got.Tax-63500) > 0.01 {
		t.Errorf("Tax = %v, want 63500", got.Tax)
	}
	if len(got.TaxLevel) != len(wantLevels) {
		t.Fatalf("len(TaxLevel) = %d, want %d", len(got.TaxLevel), len(wantLevels))
	}
	for i, w := range wantLevels {
		g := got.TaxLevel[i]
		if g.Level != w.Level || math.Abs(g.Tax-w.Tax) > 0.01 {
			t.Errorf("TaxLevel[%d] = %+v, want %+v", i, g, w)
		}
	}
}

func TestCalculateDetailedWithWHT(t *testing.T) {
	got := CalculateDetailed(600000, 15000)

	// tax คือภาษีหลังหัก WHT แต่ taxLevel รวมกันได้ภาษีก่อนหัก WHT
	if math.Abs(got.Tax-26000) > 0.01 {
		t.Errorf("Tax = %v, want 26000", got.Tax)
	}
	sum := 0.0
	for _, l := range got.TaxLevel {
		sum += l.Tax
	}
	if math.Abs(sum-41000) > 0.01 {
		t.Errorf("sum of TaxLevel = %v, want 41000", sum)
	}
}

func TestCalculateDetailedLowIncome(t *testing.T) {
	// รายได้ต่ำกว่าลดหย่อน: เงินได้สุทธิติดลบ ทุกช่วงต้องเป็น 0 แต่ยังต้องมีครบ 5 ช่วง
	got := CalculateDetailed(50000, 3000)

	if len(got.TaxLevel) != 5 {
		t.Fatalf("len(TaxLevel) = %d, want 5", len(got.TaxLevel))
	}
	for i, l := range got.TaxLevel {
		if l.Tax != 0 {
			t.Errorf("TaxLevel[%d].Tax = %v, want 0", i, l.Tax)
		}
	}
	if math.Abs(got.Tax-(-3000)) > 0.01 {
		t.Errorf("Tax = %v, want -3000 (refund)", got.Tax)
	}
}
