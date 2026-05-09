# write-vi-ui

Mode này dành cho:
- button
- label
- toast
- error message
- helper text
- empty state
- confirm dialogs

## Mục tiêu
- ngắn
- rõ động từ
- dễ scan
- ít nhập nhằng

## Nên làm
- ưu tiên động từ trực tiếp: `Lưu`, `Hủy`, `Thử lại`, `Tải xuống`, `Xem chi tiết`
- error message nên nói rõ:
  1. chuyện gì xảy ra
  2. user nên làm gì tiếp
- empty state nên nói gọn, hữu ích, không quá hoạt hình
- helper text nên giảm mơ hồ; nói đúng constraint hoặc expected format

## Tránh
- câu dài như văn email
- quá lịch sự làm UI nặng nề
- thay đổi cách gọi cùng một action giữa các màn
- thêm cảm xúc không cần thiết vào lỗi hoặc cảnh báo

## Pattern gợi ý
- lỗi: `Không tải được dữ liệu. Thử lại sau.`
- xác nhận: `Xóa mục này?` / `Bạn có chắc muốn xóa mục này?`
- helper: `Nhập email công việc của bạn.`

## Rule nhỏ
- Nếu người dùng cần hành động, ưu tiên đưa hành động lên sớm.
- Nếu có thể cắt một chữ mà nghĩa không đổi, cắt.
