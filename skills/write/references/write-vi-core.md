# write-vi-core

Mục tiêu mặc định cho prose tiếng Việt:
- đúng ý
- ngắn hơn nếu có thể
- rõ, có cấu trúc
- đúng giọng người đọc cần
- ít mùi báo cáo / ít mùi AI / không lan man

## Luật nền

### 1) Giữ nghĩa trước
- Không hi sinh ý nghĩa, dữ kiện, điều kiện hay sắc thái chỉ để câu nghe mượt hơn.
- Nếu câu gốc hơi thô nhưng đúng và rõ, chỉ sửa vừa đủ.

### 2) Ưu tiên gọn
- Mở bài vào thẳng vấn đề.
- Tránh 2 câu làm việc của 1 câu.
- Nếu một đoạn quá dày mà không có lý do, tách hoặc rút.

### 3) Giữ tiếng Việt tự nhiên
- Ưu tiên câu chủ động.
- Tránh dịch sát cấu trúc tiếng Anh.
- Giữ thuật ngữ Anh khi thật sự phổ biến (`token`, `prompt`, `API`, `dashboard`), nhưng gloss ngắn ở lần đầu nếu cần.

### 4) Dấu câu và trình bày
- Dấu câu đi sát chữ trước nó.
- Sau dấu câu là một khoảng trắng.
- Mặc định không đặt dấu phẩy trước `và` hoặc `hoặc` khi nối các thành phần đồng chức.
- Viết hoa theo logic tiếng Việt, không bê nguyên English capitalization nếu không cần.

### 5) Nhất quán xưng hô
- Một văn bản nên có một register chính.
- Nếu đã chọn `bạn`, đừng nhảy sang `anh/chị`, `quý khách`, `người dùng` lung tung.
- Nếu chưa rõ giọng, mặc định nghiêng về trung tính.

### 6) Nhất quán thuật ngữ
- Nếu đã chọn `tệp`, đừng xen `tập tin`.
- Nếu đã chọn `trang web`, đừng lúc khác lại thành `website` trừ khi có lý do.

### 7) Bớt report tone, bớt AI tone
Ưu tiên cắt hoặc thay các cụm như:
- `có thể thấy rằng`
- `điều này cho thấy`
- `nhằm mục đích`
- `đóng vai trò quan trọng`
- `trong bối cảnh hiện nay`
- `một cách hiệu quả`
- `xin thông báo rằng`

Nếu bỏ được mà câu vẫn rõ, bỏ.

## Anti-patterns

1. mở bài dài mới vào ý chính
2. đoạn ôm quá nhiều việc cùng lúc
3. buzzword Anh không giúp hiểu hơn
4. kết luận chỉ lặp lại bài
5. tự ý đổi dàn ý khi user chỉ muốn sửa câu
6. cố làm câu “sang” hơn bằng từ to và trừu tượng
7. lạm dấu chấm than, ngoặc dài, dấu ba chấm

## Default preference
Khi phân vân, chọn:
- rõ hơn thay vì kêu hơn
- ngắn hơn thay vì bóng bẩy hơn
- tự nhiên hơn thay vì “văn mẫu” hơn
