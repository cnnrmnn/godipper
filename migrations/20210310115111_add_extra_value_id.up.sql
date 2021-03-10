ALTER TABLE extra
RENAME COLUMN extra TO extra_value_id;

ALTER TABLE extra
ADD CONSTRAINT fk_extra_value FOREIGN KEY (extra_value_id)
REFERENCES extra_value (extra_value_id);
