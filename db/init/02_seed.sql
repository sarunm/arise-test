-- Customers
INSERT INTO customers (first_name, last_name, email) VALUES
    ('สมชาย',  'ใจดี',    'somchai@example.com'),
    ('สมหญิง', 'รักเรียน', 'somying@example.com'),
    ('วิชัย',  'มั่งมี',   'wichai@example.com');

-- Accounts
-- balance หน่วยสตางค์: 50000.00 บาท = 5000000
INSERT INTO accounts (customer_id, number, balance, status) VALUES
    (1, '1000000001', 5000000,  'ACTIVE'),   -- สมชาย   500.00 บาท
    (1, '1000000002', 10000000, 'ACTIVE'),   -- สมชาย   1,000.00 บาท (บัญชีที่ 2)
    (2, '1000000003', 25000000, 'ACTIVE'),   -- สมหญิง  2,500.00 บาท
    (3, '1000000004', 0,        'INACTIVE'); -- วิชัย   ปิดบัญชีแล้ว

-- Transactions
INSERT INTO transactions (from_account_id, to_account_id, amount, type) VALUES
    (NULL, 1, 5000000,  'DEPOSIT'),   -- ฝาก 500.00 บาท เข้าบัญชี 1
    (NULL, 3, 25000000, 'DEPOSIT'),   -- ฝาก 2,500.00 บาท เข้าบัญชี 3
    (3,    1, 1000000,  'TRANSFER');  -- โอน 100.00 บาท จากบัญชี 3 → 1
