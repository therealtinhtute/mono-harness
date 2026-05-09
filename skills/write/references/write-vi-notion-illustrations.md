# write-vi-notion-illustrations

> **Trong file này:** Mục tiêu · Khi nào nên thêm ảnh · Số lượng · Loại ảnh · Style rule · Layout cho Notion · Quy tắc mật độ · Pairing với prose · Caption pattern · Workflow · Mặc định gu

File này dùng khi bài **Notion report / research note / decision memo** cần thêm **ảnh minh hoạ, sơ đồ, hoặc visual so sánh**.

## Mục tiêu
- giúp người đọc hiểu nhanh hơn, không phải để trang trí
- nhìn gọn trên mobile
- ăn khớp với nhịp đọc của bài
- giảm cảm giác wall-of-text

## Khi nào nên thêm ảnh
Chỉ thêm khi ảnh làm tốt hơn text ở ít nhất 1 việc:
- so sánh 2–4 lựa chọn
- cho thấy layer / flow / vị trí trong hệ thống
- tóm tắt tradeoff nhanh
- chèn một điểm nghỉ thị giác sau đoạn kỹ thuật dày

Nếu chỉ là ý đơn giản có thể nói bằng 1–2 câu, **không cần vẽ**.

## Số lượng gợi ý
- bài ngắn: **0–1 ảnh**
- bài research / technical note cỡ vừa: **1–2 ảnh**
- mặc định đừng vượt quá **2 ảnh** trừ khi user yêu cầu rõ

## Loại ảnh nên ưu tiên
### 1) So sánh tối giản
Dùng khi cần trả lời kiểu:
- A vs B
- option nào tiện hơn / kiểm soát hơn
- tool nào hợp use case nào

Ví dụ tốt:
- quadrant tối giản
- 2–3 điểm đặt trên trục
- bảng visual cực gọn nếu thật sự cần

### 2) Layer / stack / vị trí trong hệ thống
Dùng khi cần cho người đọc thấy:
- thành phần này nằm ở đâu
- trách nhiệm thuộc layer nào
- đâu là core path

Ví dụ tốt:
- layer stack
- architecture strip rất gọn
- một flow ngắn 3–5 bước

## Style rule
- **tối giản trước**: bỏ mọi thứ không giúp hiểu nhanh hơn
- **chỉ trả về diagram itself** nếu user muốn ảnh minh hoạ cho bài; không thêm card, caption giả, khung giải thích dư
- **padding mỏng**: ưu tiên viền thở vừa đủ; nếu đang thấy “rộng tay”, thường có thể giảm khoảng **1/3 đến 1/2**
- **không trang trí kiểu slide deck**
- **không nhồi quá nhiều node** chỉ vì còn chỗ trống
- mỗi ảnh chỉ nên có **1 ý chính**

## Layout rule cho Notion
- ảnh phải đọc ổn khi nhúng vào Notion và scan trên mobile
- text trong ảnh phải ngắn
- tránh caption trong ảnh quá dài; để caption thật nằm ở block caption của Notion nếu cần
- nếu một dòng label quá dài và sát mép, lùi node vào thay vì nới canvas vô tội vạ

## Quy tắc mật độ
Checklist nhanh trước khi chốt ảnh:
- có đúng **1 thông điệp chính** chưa?
- có chữ nào thừa không?
- có khoảng trắng thừa ở ngoài rìa không?
- có phần nào sát mép hoặc bị cramped không?
- nếu giảm canvas/padding thêm nhẹ nữa thì ảnh còn ổn không?

## Pairing với prose
Ảnh nên khớp với section ngay trước hoặc ngay sau nó.

Pattern tốt:
1. đoạn mở rất ngắn
2. ảnh minh hoạ
3. 2–4 bullet giải thích takeaway

Không nên:
- quăng ảnh to vào giữa bài mà không có câu dẫn
- để ảnh nói một đằng, text nói một nẻo
- lặp lại toàn bộ nội dung ảnh trong 1 đoạn dài bên dưới

## Caption pattern gợi ý
- `Sơ đồ 1 · So sánh tối giản: ...`
- `Sơ đồ 2 · Vị trí của X trong ...`
- `Minh hoạ nhanh · Tradeoff giữa ...`

Caption tốt nên:
- gọi tên đúng job của ảnh
- nói rõ người đọc sẽ học được gì
- không dài như một đoạn văn

## Workflow gợi ý cho agent
1. xác định ảnh này để làm gì: so sánh, layer, flow, hay tradeoff
2. chốt **1 câu takeaway** trước khi vẽ
3. chọn loại ảnh tối giản nhất đủ dùng
4. render bản đầu
5. tự soi lại: padding, mật độ chữ, mép phải/mép dưới, scanability trên mobile
6. nếu thấy “còn rộng”, ưu tiên **giảm padding/canvas trước khi thêm nội dung**
7. update vào Notion rồi check lại trong context bài

## Mặc định gu nên giữ
- gọn
- kỹ thuật
- editorial nhẹ
- không phô diễn
- nhìn là hiểu gần ngay
