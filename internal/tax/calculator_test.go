package tax

import (
	"math"
	"testing"
)

// ทดสอบว่าผลรวมของ taxLevel เท่ากับภาษีรวม ในแต่ละช่วงเงินได้สุทธิ (ตั้งแต่ 0 จนถึงช่วง 35%)
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
			got := CalculateDetailed(tc.netIncome+PersonalAllowance, 0, nil)

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

// ทดสอบ Calculate: ภาษีสุทธิหลังหัก WHT รวมถึงกรณี WHT มากกว่าภาษี (ได้เงินคืน เป็นค่าติดลบ)
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
			got := Calculate(tc.totalIncome, tc.wht, nil)
			if math.Abs(got-tc.want) > 0.01 {
				t.Errorf("Calculate(%v, %v) = %v, want %v", tc.totalIncome, tc.wht, got, tc.want)
			}
		})
	}
}

// ทดสอบว่า taxLevel ครบ 5 ช่วง พร้อมป้ายช่วงและภาษีของแต่ละช่วง (ตัวอย่างรายได้ 750,000)
func TestCalculateDetailed(t *testing.T) {
	got := CalculateDetailed(750000, 0, nil)

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

// ทดสอบกรณีมี WHT: ภาษีหลังหัก WHT ต้องลดลง แต่ taxLevel รวมกันได้ภาษีก่อนหัก WHT
func TestCalculateDetailedWithWHT(t *testing.T) {
	got := CalculateDetailed(600000, 15000, nil)

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

// ทดสอบกรณีเงินได้สุทธิติดลบ: รายได้รวมต่ำกว่าค่าลดหย่อนส่วนตัว
func TestCalculateDetailedLowIncome(t *testing.T) {
	// รายได้ต่ำกว่าลดหย่อน: เงินได้สุทธิติดลบ ทุกช่วงต้องเป็น 0 แต่ยังต้องมีครบ 5 ช่วง
	got := CalculateDetailed(50000, 3000, nil)

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

// ทดสอบค่าลดหย่อนหลายกรณี: ไม่มี, บริจาค 0, บริจาคไม่เกินเพดาน, บริจาคเกินเพดาน, หลายรายการรวมกัน, มี WHT, รายได้สุทธิติดลบ
func TestCalculateDetailedWithAllowances(t *testing.T) {
	tests := []struct {
		name        string
		totalIncome float64
		wht         float64
		allowances  []Allowance
		want        float64
	}{
		{"nil allowances", 750000, 0, nil, 63500},
		{"donation zero", 750000, 0, []Allowance{{"donation", 0}}, 63500},
		{"donation under cap", 700000, 0, []Allowance{{"donation", 50000}}, 48500},
		{"donation exactly at cap", 700000, 0, []Allowance{{"donation", 100000}}, 41000},
		{"donation over cap is capped", 700000, 0, []Allowance{{"donation", 120000}}, 41000},
		{"over cap, example 4 income", 850000, 0, []Allowance{{"donation", 150000}}, 63500},
		{"several donations are summed then capped", 700000, 0,
			[]Allowance{{"donation", 60000}, {"donation", 60000}}, 41000},
		{"donation with wht", 700000, 10000, []Allowance{{"donation", 120000}}, 31000},
		{"donation makes net income negative", 100000, 0, []Allowance{{"donation", 50000}}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateDetailed(tc.totalIncome, tc.wht, tc.allowances)

			sum := 0.0
			for _, l := range got.TaxLevel {
				sum += l.Tax
			}
			if math.Abs(got.Tax-tc.want) > 0.01 {
				t.Errorf("Tax = %v, want %v", got.Tax, tc.want)
			}
			// taxLevel รวมกันได้ภาษีก่อนหัก WHT
			if math.Abs(sum-(tc.want+tc.wht)) > 0.01 {
				t.Errorf("sum of TaxLevel = %v, want %v", sum, tc.want+tc.wht)
			}
		})
	}
}

// ทดสอบการปัดทศนิยม: ทุกค่าใน response ต้องมีไม่เกิน 2 ตำแหน่ง
// และผลรวมของ taxLevel ต้องตรงกับ tax + wht พอดี
func TestCalculateDetailedRounding(t *testing.T) {
	tests := []struct {
		name        string
		totalIncome float64
		wht         float64
		allowances  []Allowance
		wantTax     float64
		wantLevels  []float64 // ภาษีรายช่วง ตามลำดับ 5 ช่วง
	}{
		{"cents in 15% bracket", 600000.33, 0, nil, 41000.05, []float64{0, 35000, 6000.05, 0, 0}},
		{"cents with wht", 600000.33, 0.07, nil, 40999.98, []float64{0, 35000, 6000.05, 0, 0}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateDetailed(tc.totalIncome, tc.wht, tc.allowances)

			if got.Tax != tc.wantTax {
				t.Errorf("Tax = %v, want exactly %v", got.Tax, tc.wantTax)
			}
			sum := 0.0
			for i, l := range got.TaxLevel {
				if l.Tax != tc.wantLevels[i] {
					t.Errorf("TaxLevel[%d] = %v, want exactly %v", i, l.Tax, tc.wantLevels[i])
				}
				sum += l.Tax
			}
			// sum ต้องตรงกับ tax + wht (เทียบด้วย tolerance เล็กมาก เพราะการบวก float ยังมีเศษระดับ 1e-9)
			if math.Abs(sum-(got.Tax+tc.wht)) > 0.001 {
				t.Errorf("sum of TaxLevel = %v, want tax+wht = %v", sum, got.Tax+tc.wht)
			}
		})
	}
}
