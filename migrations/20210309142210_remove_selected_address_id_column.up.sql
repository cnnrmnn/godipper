ALTER TABLE user
DROP CONSTRAINT fk_address;

ALTER TABLE user
DROP KEY fk_address;

ALTER TABLE user
DROP COLUMN selected_address_id;
