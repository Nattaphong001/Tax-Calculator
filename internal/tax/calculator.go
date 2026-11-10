package tax

import "math"

// bracket คือหนึ่งขั้นของภาษีขั้นบันได
type bracket struct {
	upper float64 // ขอบบนของช่วง (เงินได้สุทธิ)
	rate  float64 // อัตราภาษีของช่วงนี้
}

// ขั้นบันไดเก็บเป็นข้อมูล ไม่ hardcode ใน if/else
var brackets = []bracket{
	{150000, 0.00},
	{500000, 0.10},
	{1000000, 0.15},
	{2000000, 0.20},
	{math.Inf(1), 0.35}, // ช่วงสุดท้ายไม่มีขอบบน
}

// PersonalAllowance คือค่าลดหย่อนส่วนตัว (ค่าคงที่)
const PersonalAllowance = 60000

// Calculate รับรายได้รวมและ WHT คืนภาษีที่ต้องจ่ายจริง (ติดลบ = ได้เงินคืน)
func Calculate(totalIncome, wht float64) float64 {
	netIncome := totalIncome - PersonalAllowance
	return CalculateTax(netIncome) - wht
}

// CalculateTax รับเงินได้สุทธิ (หลังหักลดหย่อนแล้ว) แล้วคืนภาษีรวม
func CalculateTax(netIncome float64) float64 {
	if netIncome <= 0 {
		return 0
	}

	total := 0.0
	lower := 0.0 // ขอบล่างของช่วงปัจจุบัน

	for _, b := range brackets {
		if netIncome <= lower {
			break // เงินหมดแล้ว
		}
		portion := min(netIncome, b.upper) - lower // เงินส่วนที่ตกในช่วงนี้
		total += portion * b.rate
		lower = b.upper
	}
	return total
}
