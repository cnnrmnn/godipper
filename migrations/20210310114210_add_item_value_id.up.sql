ALTER TABLE item
RENAME COLUMN item TO item_value_id;

ALTER TABLE item
ADD CONSTRAINT fk_item_value FOREIGN KEY (item_value_id)
REFERENCES item_value (item_value_id);
