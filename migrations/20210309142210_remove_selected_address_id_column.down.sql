ALTER TABLE user
ADD selected_address_id SMALLINT UNSIGNED;

ALTER TABLE user
ADD CONSTRAINT fk_address FOREIGN KEY (selected_address_id)
REFERENCES address(address_id);
