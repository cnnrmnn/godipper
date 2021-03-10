ALTER TABLE extra
DROP CONSTRAINT fk_extra_value;

ALTER TABLE extra
DROP KEY fk_extra_value;

ALTER TABLE extra
RENAME COLUMN extra_value_id TO extra;
