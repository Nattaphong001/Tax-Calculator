package handler

// validate ตรวจ request คืนข้อความ error หรือ "" ถ้าผ่าน
func validate(req calculationRequest) string {
	// TotalIncome เป็น required field ถ้าไม่ส่งมาจะเป็น nil
	if req.TotalIncome == nil {
		return "totalIncome is required"
	}
	if *req.TotalIncome < 0 {
		return "totalIncome must be a positive number"
	}
	// WHT (ภาษีหัก ณ ที่จ่าย) เป็น optional ถ้าไม่ส่งมาถือเป็น 0
	if req.WHT != nil {
		if *req.WHT < 0 {
			return "wht must be a non-negative number"
		}
		// WHT ต้องไม่มากกว่า TotalIncome
		if *req.WHT > *req.TotalIncome {
			return "wht cannot be greater than totalIncome"
		}
	}
	if req.Allowances == nil {
		return "allowances is required and must be an array"
	}
	// ตรวจค่าลดหย่อนทีละรายการ: ตรวจ amount ก่อน type
	// (ถ้าผิดทั้งคู่ จะแจ้งเรื่อง amount ก่อน)
	for _, a := range req.Allowances {
		if a.Amount < 0 {
			return "allowance amount must be a non-negative number"
		}
		// ตอนนี้รองรับเฉพาะ donation
		if a.AllowanceType != "donation" {
			return "unknown allowanceType: " + a.AllowanceType
		}
	}
	return ""
}
