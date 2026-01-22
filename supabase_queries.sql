-- ============================================
-- SUPABASE QUERY COLLECTION
-- ============================================

-- ============================================
-- 1. BOOKINGS TABLE - แก้ไขข้อมูลการจองนัดหมาย
-- ============================================

-- ดูข้อมูลการจองทั้งหมด
SELECT * FROM bookings ORDER BY appointment_date DESC;

-- ดูการจองตามสถานะ
SELECT * FROM bookings WHERE status = 'pending' ORDER BY appointment_date DESC;
SELECT * FROM bookings WHERE status = 'confirmed' ORDER BY appointment_date DESC;
SELECT * FROM bookings WHERE status = 'cancelled' ORDER BY appointment_date DESC;

-- ดูการจองของลูกค้าคนหนึ่ง
SELECT * FROM bookings WHERE customer_id = 'CUSTOMER_ID_HERE';

-- อัปเดตสถานะการจอง
UPDATE bookings SET status = 'confirmed' WHERE id = 'BOOKING_ID_HERE';
UPDATE bookings SET status = 'cancelled' WHERE id = 'BOOKING_ID_HERE';

-- อัปเดตหมายเหตุ
UPDATE bookings SET notes = 'หมายเหตุใหม่' WHERE id = 'BOOKING_ID_HERE';

-- อัปเดตวันที่นัดหมาย
UPDATE bookings SET appointment_date = '2026-02-15' WHERE id = 'BOOKING_ID_HERE';

-- ลบการจองนัดหมาย
DELETE FROM bookings WHERE id = 'BOOKING_ID_HERE';

-- ============================================
-- 2. APPOINTMENTS TABLE - แก้ไขรายละเอียดการตรวจ
-- ============================================

-- ดูการตรวจทั้งหมด
SELECT * FROM appointments ORDER BY created_at DESC;

-- ดูการตรวจตามการจอง
SELECT * FROM appointments WHERE booking_id = 'BOOKING_ID_HERE';

-- ดูการตรวจตามแพทย์
SELECT * FROM appointments WHERE doctor_id = 'DOCTOR_ID_HERE';

-- อัปเดตสถานะการตรวจ
UPDATE appointments SET status = 'completed' WHERE id = 'APPOINTMENT_ID_HERE';
UPDATE appointments SET status = 'cancelled' WHERE id = 'APPOINTMENT_ID_HERE';

-- อัปเดตสถานที่ตรวจ
UPDATE appointments SET location = 'ห้อง 101' WHERE id = 'APPOINTMENT_ID_HERE';

-- อัปเดตแพทย์ที่ตรวจ
UPDATE appointments SET doctor_id = 'NEW_DOCTOR_ID' WHERE id = 'APPOINTMENT_ID_HERE';

-- ลบการตรวจ
DELETE FROM appointments WHERE id = 'APPOINTMENT_ID_HERE';

-- ============================================
-- 3. USERS TABLE - แก้ไขข้อมูลผู้ใช้
-- ============================================

-- ดูข้อมูลผู้ใช้ทั้งหมด
SELECT * FROM users ORDER BY created_at DESC;

-- ดูผู้ใช้ตามบทบาท
SELECT * FROM users WHERE role = 'nurse';
SELECT * FROM users WHERE role = 'customer';
SELECT * FROM users WHERE role = 'admin';

-- ดูผู้ใช้ที่ใช้งาน
SELECT * FROM users WHERE is_active = true;

-- อัปเดตชื่อผู้ใช้
UPDATE users SET full_name = 'ชื่อใหม่' WHERE id = 'USER_ID_HERE';

-- อัปเดตเบอร์โทร
UPDATE users SET phone = '0812345678' WHERE id = 'USER_ID_HERE';

-- อัปเดตอีเมล
UPDATE users SET email = 'newemail@example.com' WHERE id = 'USER_ID_HERE';

-- อัปเดตบริษัท
UPDATE users SET company_name = 'ชื่อบริษัทใหม่' WHERE id = 'USER_ID_HERE';

-- อัปเดตสถานะการใช้งาน
UPDATE users SET is_active = false WHERE id = 'USER_ID_HERE';
UPDATE users SET is_active = true WHERE id = 'USER_ID_HERE';

-- อัปเดตข้อมูลส่วนตัว
UPDATE users SET 
  birth_date = '1990-01-15',
  gender = 'ชาย',
  blood_type = 'O+',
  age = 34
WHERE id = 'USER_ID_HERE';

-- ลบผู้ใช้
DELETE FROM users WHERE id = 'USER_ID_HERE';

-- ============================================
-- 4. DOCTORS TABLE - แก้ไขข้อมูลแพทย์
-- ============================================

-- ดูข้อมูลแพทย์ทั้งหมด
SELECT * FROM doctors ORDER BY created_at DESC;

-- ดูแพทย์ตามความเชี่ยวชาญ
SELECT * FROM doctors WHERE specialty = 'ตาแพทย์';
SELECT * FROM doctors WHERE specialty = 'ศัลยแพทย์';

-- ดูแพทย์ที่ใช้งาน
SELECT * FROM doctors WHERE is_active = true;

-- อัปเดตชื่อแพทย์
UPDATE doctors SET full_name = 'นพ.ชื่อใหม่' WHERE id = 'DOCTOR_ID_HERE';

-- อัปเดตความเชี่ยวชาญ
UPDATE doctors SET specialty = 'ตาแพทย์' WHERE id = 'DOCTOR_ID_HERE';

-- อัปเดตเบอร์โทร
UPDATE doctors SET phone = '0812345678' WHERE id = 'DOCTOR_ID_HERE';

-- อัปเดตอีเมล
UPDATE doctors SET email = 'doctor@example.com' WHERE id = 'DOCTOR_ID_HERE';

-- อัปเดตสถานะการใช้งาน
UPDATE doctors SET is_active = false WHERE id = 'DOCTOR_ID_HERE';
UPDATE doctors SET is_active = true WHERE id = 'DOCTOR_ID_HERE';

-- ลบแพทย์
DELETE FROM doctors WHERE id = 'DOCTOR_ID_HERE';

-- ============================================
-- 5. TIME_SLOTS TABLE - แก้ไขช่วงเวลาว่าง
-- ============================================

-- ดูช่วงเวลาทั้งหมด
SELECT * FROM time_slots ORDER BY created_at DESC;

-- ดูช่วงเวลาตามแพทย์
SELECT ts.* FROM time_slots ts
JOIN doctor_schedules ds ON ts.doctor_schedule_id = ds.id
WHERE ds.doctor_id = 'DOCTOR_ID_HERE';

-- ดูช่วงเวลาว่าง
SELECT * FROM time_slots WHERE status = 'available';

-- ดูช่วงเวลาที่ถูกจอง
SELECT * FROM time_slots WHERE status = 'booked';

-- อัปเดตสถานะช่วงเวลา
UPDATE time_slots SET status = 'available' WHERE id = 'SLOT_ID_HERE';
UPDATE time_slots SET status = 'booked' WHERE id = 'SLOT_ID_HERE';

-- อัปเดตจำนวนที่จองได้
UPDATE time_slots SET booked_count = 0 WHERE id = 'SLOT_ID_HERE';
UPDATE time_slots SET max_capacity = 5 WHERE id = 'SLOT_ID_HERE';

-- ลบช่วงเวลา
DELETE FROM time_slots WHERE id = 'SLOT_ID_HERE';

-- ============================================
-- 6. DOCTOR_SCHEDULES TABLE - แก้ไขตารางแพทย์
-- ============================================

-- ดูตารางแพทย์ทั้งหมด
SELECT * FROM doctor_schedules ORDER BY schedule_date DESC;

-- ดูตารางแพทย์คนหนึ่ง
SELECT * FROM doctor_schedules WHERE doctor_id = 'DOCTOR_ID_HERE' ORDER BY schedule_date DESC;

-- ดูตารางแพทย์ตามวันที่
SELECT * FROM doctor_schedules WHERE schedule_date = '2026-02-15';

-- อัปเดตสถานะความพร้อม
UPDATE doctor_schedules SET is_available = true WHERE id = 'SCHEDULE_ID_HERE';
UPDATE doctor_schedules SET is_available = false WHERE id = 'SCHEDULE_ID_HERE';

-- ลบตารางแพทย์
DELETE FROM doctor_schedules WHERE id = 'SCHEDULE_ID_HERE';

-- ============================================
-- 7. QUERY ที่มีประโยชน์ - JOIN ข้อมูลหลายตาราง
-- ============================================

-- ดูการจองพร้อมข้อมูลลูกค้า
SELECT 
  b.id,
  b.booking_number,
  b.appointment_date,
  b.status,
  u.full_name as customer_name,
  u.phone as customer_phone,
  u.company_name
FROM bookings b
LEFT JOIN users u ON b.customer_id = u.id
ORDER BY b.appointment_date DESC;

-- ดูการตรวจพร้อมข้อมูลแพทย์
SELECT 
  a.id,
  a.booking_id,
  a.service_type,
  a.status,
  d.full_name as doctor_name,
  d.specialty,
  ts.start_time,
  ts.end_time
FROM appointments a
LEFT JOIN doctors d ON a.doctor_id = d.id
LEFT JOIN time_slots ts ON a.time_slot_id = ts.id
ORDER BY a.created_at DESC;

-- ดูการจองพร้อมรายละเอียดการตรวจ
SELECT 
  b.id as booking_id,
  b.booking_number,
  b.appointment_date,
  b.status as booking_status,
  u.full_name as customer_name,
  u.phone,
  a.id as appointment_id,
  a.service_type,
  a.status as appointment_status,
  d.full_name as doctor_name,
  ts.start_time,
  ts.end_time
FROM bookings b
LEFT JOIN users u ON b.customer_id = u.id
LEFT JOIN appointments a ON b.id = a.booking_id
LEFT JOIN doctors d ON a.doctor_id = d.id
LEFT JOIN time_slots ts ON a.time_slot_id = ts.id
ORDER BY b.appointment_date DESC;

-- ============================================
-- 8. BULK UPDATE - อัปเดตหลายรายการพร้อมกัน
-- ============================================

-- เปลี่ยนสถานะการจองทั้งหมดที่รอยืนยันเป็นยืนยัน
UPDATE bookings SET status = 'confirmed' WHERE status = 'pending' AND appointment_date < NOW();

-- เปลี่ยนสถานะการตรวจทั้งหมดที่เกินวันที่เป็น completed
UPDATE appointments SET status = 'completed' 
WHERE status = 'pending' 
AND booking_id IN (
  SELECT id FROM bookings WHERE appointment_date < NOW()
);

-- ปิดใช้งานแพทย์ทั้งหมดที่ไม่มีการจอง
UPDATE doctors SET is_active = false 
WHERE id NOT IN (
  SELECT DISTINCT doctor_id FROM appointments
);

-- ============================================
-- 9. STATISTICS & REPORTS
-- ============================================

-- นับจำนวนการจองตามสถานะ
SELECT status, COUNT(*) as count FROM bookings GROUP BY status;

-- นับจำนวนการจองตามแพทย์
SELECT d.full_name, COUNT(a.id) as appointment_count
FROM doctors d
LEFT JOIN appointments a ON d.id = a.doctor_id
GROUP BY d.id, d.full_name
ORDER BY appointment_count DESC;

-- นับจำนวนการจองตามบริษัท
SELECT u.company_name, COUNT(b.id) as booking_count
FROM users u
LEFT JOIN bookings b ON u.id = b.customer_id
WHERE u.role = 'customer'
GROUP BY u.company_name
ORDER BY booking_count DESC;

-- ============================================
-- 10. CLEANUP - ลบข้อมูลเก่า
-- ============================================

-- ลบการจองที่ยกเลิกแล้วเกิน 90 วัน
DELETE FROM bookings 
WHERE status = 'cancelled' 
AND updated_at < NOW() - INTERVAL '90 days';

-- ลบการตรวจที่เสร็จสิ้นแล้วเกิน 1 ปี
DELETE FROM appointments 
WHERE status = 'completed' 
AND updated_at < NOW() - INTERVAL '1 year';
