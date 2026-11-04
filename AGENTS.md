# 🤖 AGENTS & DEVELOPER CONTEXT: BÁNH TRÁNG CẬU HAI (BACKEND API)

> **Tài liệu ngữ cảnh dành cho AI Coding Assistants (Antigravity, Cursor, Copilot, Claude Code) & Lập trình viên Backend Go.**

---

## 📌 1. TỔNG QUAN HỆ THỐNG
- **Dự án**: `member-shop-api` — Dịch vụ REST API hiệu năng cao cho hệ thống F&B Bánh Tráng Cậu Hai.
- **Runtime**: Go 1.24, Gin Web Framework, Docker Multi-stage static binary (`CGO_ENABLED=0`).
- **Database**: MongoDB Atlas (MongoDB Driver v2) & Redis Cache-Aside.
- **Production Host**: Render.com (`https://member-shop-api.onrender.com`).

---

## 🛠️ 2. KIẾN TRÚC MÃ NGUỒN (CLEAN ARCHITECTURE)

```
member-shop-api/
├── cmd/api/main.go            # Điểm khởi chạy HTTP Server & Graceful Shutdown
├── internal/
│   ├── app/                   # Khởi tạo DI ModuleContext, router, database & seeder
│   ├── config/                # Đọc biến môi trường (PORT, MONGODB_URI, JWT, REDIS)
│   ├── db/                    # Kết nối MongoDB Atlas
│   ├── handler/               # HTTP Handlers (Validation & Response Envelope)
│   ├── middlewares/           # CORS (Dynamic ALLOWED_ORIGINS), RateLimiter, Gzip, Auth, Recovery
│   ├── model/                 # BSON Entities & Business Logic (Customer, Order, Product, Staff)
│   ├── repository/            # Thao tác MongoDB queries, Atomic deduction, OCC versioning
│   ├── routes/                # Định tuyến /v1 công khai & /admin bảo mật
│   └── services/              # Business logic tính điểm, voucher, giao ca, JWT token
└── render.yaml                # Cấu hình Blueprint tự động deploy Render.com (Singapore Region)
```

---

## 💎 3. QUY TẮC BẢO MẬT & NGHIỆP VỤ BẮT BUỘC
1. **Bảo vệ mã PIN**: Băm 100% bằng `bcrypt.GenerateFromPassword`. Tuyệt đối không lưu plaintext.
2. **Khóa tạm Brute-force**: Sau 5 lần nhập sai mã PIN liên tiếp, khóa tạm 15 phút qua Redis key `failed_login:<user>`.
3. **Atomic Points Deduction (SEC-06)**: Trừ điểm trong `customer_repository.go` bắt buộc dùng điều kiện nguyên tử: `bson.M{"_id": id, "availablePoints": bson.M{"$gte": points}}`.
4. **Optimistic Concurrency Control (OCC)**: Cập nhật dữ liệu luôn kiểm tra `version == req.Version` và tăng `version + 1`.
5. **Đồng bộ hạng hội viên**: Mốc thăng hạng chuẩn:
   - Silver: từ 200.000đ (x1.15)
   - Gold: từ 600.000đ (x1.3)
   - Diamond: từ 1.500.000đ (x1.5)
