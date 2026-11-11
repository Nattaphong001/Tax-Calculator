package tax

import "math"

// bracket คือหนึ่งขั้นของภาษีขั้นบันได
type bracket struct {
	label string  // ป้ายช่วงที่แสดงใน taxLevel
	upper float64 // ขอบบนของช่วง (เงินได้สุทธิ)
	rate  float64 // อัตราภาษีของช่วงนี้
}

// Level คือภาษีของหนึ่งขั้นบันได
type Level struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

// Result คือผลคำนวณพร้อมรายละเอียดรายขั้น
type Result struct {
	Tax      float64 `json:"tax"`
	TaxLevel []Level `json:"taxLevel"`
}

// Allowance คือค่าลดหย่อนหนึ่งรายการ
type Allowance struct {
	Type   string
	Amount float64
}

// PersonalAllowance คือค่าลดหย่อนส่วนตัว (ค่าคงที่)
const PersonalAllowance = 60000

// DonationCap คือเพดานค่าลดหย่อนเงินบริจาค
const DonationCap = 100000

// ขอบบนและอัตราภาษีของแต่ละช่วง (เรียงจากน้อยไปมาก)
var brackets = []bracket{
	{"0-150,000", 150000, 0.00},
	{"150,001-500,000", 500000, 0.10},
	{"500,001-1,000,000", 1000000, 0.15},
	{"1,000,001-2,000,000", 2000000, 0.20},
	{"2,000,001 ขึ้นไป", math.Inf(1), 0.35}, // ช่วงสุดท้ายไม่มีขอบบน → ใช้ Inf
}

// Calculate รับรายได้รวม WHT และค่าลดหย่อน คืนเฉพาะภาษีที่ต้องจ่ายจริง
func Calculate(totalIncome, wht float64, allowances []Allowance) float64 {
	return CalculateDetailed(totalIncome, wht, allowances).Tax
}

// CalculateDetailed รับรายได้รวม WHT และค่าลดหย่อน คืนภาษีที่ต้องจ่ายจริงพร้อมรายละเอียดรายขั้น
func CalculateDetailed(totalIncome, wht float64, allowances []Allowance) Result {
	netIncome := totalIncome - PersonalAllowance - totalAllowance(allowances)
	levels := make([]Level, 0, len(brackets))
	total := 0.0
	lower := 0.0

	// คำนวณภาษีของแต่ละช่วง
	for _, b := range brackets {
		// เงินส่วนที่ตกในช่วงนี้ (ถ้าเป็นลบ → 0)
		portion := math.Max(0, min(netIncome, b.upper)-lower)
		levelTax := round2(portion * b.rate)
		// เพิ่ม Level ของช่วงนี้เข้าไปในผลลัพธ์
		levels = append(levels, Level{Level: b.label, Tax: levelTax})
		total += levelTax
		// อัปเดตขอบล่างสำหรับช่วงถัดไป
		lower = b.upper
	}

	return Result{Tax: total - wht, TaxLevel: levels}
}

// totalAllowance รวมค่าลดหย่อนจากรายการ โดยรวมเงินบริจาคก่อนแล้วค่อยตัดเพดาน
func totalAllowance(allowances []Allowance) float64 {
	donation := 0.0
	for _, a := range allowances {
		if a.Type == "donation" {
			donation += a.Amount
		}
	}
	return math.Min(donation, DonationCap)
}

// round2 ปัดเศษเป็นทศนิยม 2 ตำแหน่ง
func round2(x float64) float64 {
	return math.Round(x*100) / 100
}
