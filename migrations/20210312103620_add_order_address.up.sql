ALTER TABLE orders
ADD address_id SMALLINT UNSIGNED NOT NULL;

ALTER TABLE orders
ADD CONSTRAINT fk_order_address FOREIGN KEY (address_id)
REFERENCES addresses (address_id);
