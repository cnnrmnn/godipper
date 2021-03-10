ALTER TABLE item
DROP CONSTRAINT fk_item_value;

ALTER TABLE item
DROP KEY fk_item_value;

ALTER TABLE item
RENAME COLUMN item_value_id TO item;
