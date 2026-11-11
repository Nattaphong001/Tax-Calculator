# Tax Calculator API

REST API สำหรับคำนวณภาษีเงินได้บุคคลธรรมดา พัฒนาด้วยภาษา Go โดยใช้ Standard Library (`net/http`, `encoding/json`) ครอบคลุมการคำนวณภาษีแบบขั้นบันได การหัก ณ ที่จ่าย (WHT) ค่าลดหย่อนส่วนตัว และค่าลดหย่อนเงินบริจาค

---

## เทคโนโลยีที่ใช้

- **Language:** Go 1.27.2+
- **Dependencies:** Standard Library (`net/http`, `encoding/json`)

---

## โครงสร้างโปรเจกต์

```text
tax-calculator/
├── examples/
│   ├── tc1-no-tax.json
│   ├── tc2-single-stage.json
│   ├── tc3-multi-stage.json
│   ├── tc4-wht.json
│   ├── tc5-allowances.json
│   ├── tc6-bad-negative-income.json
│   └── tc7-bad-wht-exceed.json
├── internal/
│   ├── handler/
│   │   ├── calculation.go
│   │   ├── handler_test.go
│   │   ├── health.go
│   │   ├── response.go
│   │   ├── routes.go
│   │   └── validation.go
│   └── tax/
│       ├── calculator.go
│       └── calculator_test.go
├── .gitignore
├── go.mod
├── go.sum
├── main.go
└── README.md
```

---

## วิธีติดตั้งและรันโปรแกรม

### ข้อกำหนด

- ติดตั้ง Go เวอร์ชัน 1.27.2 ขึ้นไป
- โปรเจกต์ใช้ Go Standard Library เท่านั้น จึงไม่มี third-party dependencies ที่ต้องติดตั้งเพิ่มเติม

### วิธีรันโปรแกรม

```bash
go run .
```

API ใช้พอร์ต `8080` เป็นค่าเริ่มต้น และสามารถกำหนดพอร์ตผ่าน environment variable `PORT` ได้

### วิธีรัน Unit Tests

```bash
go test ./... -v
```

---

## หลักการคำนวณภาษี

ระบบหักค่าลดหย่อนส่วนตัว 60,000 บาท และค่าลดหย่อนเงินบริจาค (สูงสุด 100,000 บาท) จากรายได้ก่อนคำนวณภาษีตามขั้น:

| เงินได้สุทธิในแต่ละขั้น | อัตราภาษี |
|---|---:|
| 0–150,000 บาท | 0% |
| 150,001–500,000 บาท | 10% |
| 500,001–1,000,000 บาท | 15% |
| 1,000,001–2,000,000 บาท | 20% |
| มากกว่า 2,000,000 บาท | 35% |

จากนั้นนำ WHT มาหักจากภาษีรวม ผลลัพธ์ `tax` ที่น้อยกว่า 0 หมายถึงยอดภาษีคืน

---

## API Endpoints

### GET `/health`
เช็คสถานะการทำงานของ HTTP Server

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

### POST `/tax/calculations` — endpoint เดียวสำหรับคำนวณภาษี

โจทย์แบ่งการใช้งานเป็น 4 กรณีด้านล่าง ทั้งหมดใช้ endpoint เดียวกัน ไม่ใช่คนละ route โดยทุก response จะมี `tax` และ `taxLevel`

#### กรณีที่ 1: คำนวณภาษีพื้นฐาน

ส่งรายได้โดยไม่มี WHT และไม่มีค่าลดหย่อนเพิ่มเติม:

```json
{
  "totalIncome": 750000,
  "wht": 0,
  "allowances": []
}
```

ผลลัพธ์ `tax` คือ `63500`

#### กรณีที่ 2: คำนวณภาษีพร้อมรายละเอียดแต่ละขั้น

ใช้ request เดียวกับกรณีพื้นฐาน ระบบส่ง `taxLevel` แยกภาษีทั้ง 5 ขั้นมาพร้อมกับ `tax`:

```json
{
  "tax": 63500,
  "taxLevel": [
    { "level": "0-150,000", "tax": 0 },
    { "level": "150,001-500,000", "tax": 35000 },
    { "level": "500,001-1,000,000", "tax": 28500 },
    { "level": "1,000,001-2,000,000", "tax": 0 },
    { "level": "2,000,001 ขึ้นไป", "tax": 0 }
  ]
}
```

#### กรณีที่ 3: คำนวณภาษีพร้อม WHT

ส่ง WHT ใน request ระบบหัก WHT จากภาษีรวม:

```json
{
  "totalIncome": 600000,
  "wht": 15000,
  "allowances": []
}
```

ผลลัพธ์ `tax` คือ `26000`; ส่วน `taxLevel` ยังคงแสดงภาษีรายขั้นก่อนหัก WHT

#### กรณีที่ 4: รองรับค่าลดหย่อน (Bonus)

ระบุเงินบริจาคใน `allowances`; หักลดหย่อนได้สูงสุด 100,000 บาท:

```json
{
  "totalIncome": 700000,
  "wht": 0,
  "allowances": [
    {
      "allowanceType": "donation",
      "amount": 120000
    }
  ]
}
```

ระบบใช้เงินบริจาคลดหย่อน 100,000 บาท จึงได้ `tax` เท่ากับ `41000`

**ข้อมูลใน Request และเงื่อนไข validation:**

| Field | ชนิด | เงื่อนไข |
|---|---|---|
| `totalIncome` | number | จำเป็น ต้องไม่ติดลบ (ค่า `0` ใช้ได้) |
| `wht` | number | ไม่บังคับ ถ้าไม่ส่งจะถือเป็น `0`; ต้องไม่ติดลบและต้องไม่เกิน `totalIncome` |
| `allowances` | array | จำเป็น ต้องส่งเป็น array (ใช้ `[]` หากไม่มีรายการ); รองรับ `allowanceType: "donation"` และ `amount` ต้องไม่ติดลบ |

ค่าลดหย่อนส่วนตัว 60,000 บาทจะถูกนำไปหักให้อัตโนมัติ ปัจจุบันรองรับค่าลดหย่อนประเภท `donation` เท่านั้น

ค่า `tax` คือภาษีหลังหัก WHT แล้ว จึงอาจเป็นค่าติดลบได้เมื่อ WHT มากกว่าภาษีที่คำนวณได้ ส่วน `taxLevel` แสดงภาษีแยกตามขั้นก่อนหัก WHT

**Error Response (400 Bad Request) — ตัวอย่างเมื่อ `totalIncome` ติดลบ:**
```json
{
  "error": "totalIncome must be a positive number"
}
```
> หมายเหตุ: API อนุญาต `totalIncome` เท่ากับ `0` ข้อความ error ข้างต้นเป็นข้อความที่ API ส่งกลับเมื่อ `totalIncome` ติดลบ

กรณี JSON ไม่ถูกต้อง, ไม่มี `totalIncome` หรือ `allowances`, ค่า `wht` ไม่ถูกต้อง, หรือข้อมูลค่าลดหย่อนไม่ผ่าน validation จะตอบ `400 Bad Request` พร้อม JSON รูปแบบ `{"error":"ข้อความอธิบายข้อผิดพลาด"}` ส่วนการใช้ HTTP method ไม่ตรงกับ endpoint จะตอบ `405 Method Not Allowed`

---

## ทดสอบ API จริงผ่าน PowerShell หรือ Postman

### 1. ทดสอบผ่าน PowerShell

รันคำสั่งจากโฟลเดอร์หลักของโปรเจกต์ ขณะ API ทำงานอยู่ใน Terminal อีกหน้าต่างหนึ่ง:

```powershell
function Test-TaxCase {
    param(
        [string]$Name,
        [string]$File
    )

    Write-Host "`n=== $Name ==="
    $body = Get-Content -Raw $File
    try {
        Invoke-RestMethod -Method Post `
            -Uri "http://localhost:8080/tax/calculations" `
            -ContentType "application/json" `
            -Body $body | ConvertTo-Json -Depth 5
    }
    catch {
        if ($_.ErrorDetails.Message) {
            Write-Host "HTTP error response:"
            Write-Host $_.ErrorDetails.Message
        }
        else {
            throw
        }
    }
}

# Test Case 1: ไม่ต้องเสียภาษี (tax = 0)
Test-TaxCase "Test Case 1" "examples/tc1-no-tax.json"

# Test Case 2: คำนวณภาษีขั้นเดียว (tax = 14000)
Test-TaxCase "Test Case 2" "examples/tc2-single-stage.json"

# Test Case 3: คำนวณภาษีหลายขั้น (tax = 138000)
Test-TaxCase "Test Case 3" "examples/tc3-multi-stage.json"

# Test Case 4: มี WHT (tax = 16000)
Test-TaxCase "Test Case 4" "examples/tc4-wht.json"

# Test Case 5: มีค่าลดหย่อน (tax = 41000)
Test-TaxCase "Test Case 5" "examples/tc5-allowances.json"

# Test Case 6: totalIncome ติดลบ (คาดหวัง HTTP 400)
Test-TaxCase "Test Case 6" "examples/tc6-bad-negative-income.json"

# Test Case 7: wht มากกว่า totalIncome (คาดหวัง HTTP 400)
Test-TaxCase "Test Case 7" "examples/tc7-bad-wht-exceed.json"
```

---

### 2. ทดสอบผ่าน Postman

1. เปิดแอป **Postman** แล้วสร้าง Request ใหม่
2. เลือก Method เป็น `POST` และใส่ URL: `http://localhost:8080/tax/calculations`
3. ไปที่แท็บ **Headers** และเพิ่ม key-value:
   - `Content-Type`: `application/json`
4. ไปที่แท็บ **Body** -> เลือก **raw** -> เลือกชนิดเป็น **JSON**
5. วางข้อมูล Body ดังนี้:
   ```json
   {
     "totalIncome": 750000,
     "wht": 0,
     "allowances": [
       {
         "allowanceType": "donation",
         "amount": 0
       }
     ]
   }
   ```
6. กดปุ่ม **Send** เพื่อดู Response ผลการคำนวณ