ALTER TABLE orders
DROP CONSTRAINT fk_order_address;

ALTER TABLE orders
DROP KEY fk_order_address;

ALTER TABLE orders
DROP COLUMN address_id;
