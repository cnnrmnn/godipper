ALTER TABLE triple_dipper
DROP CONSTRAINT fk_td_order;

ALTER TABLE triple_dipper
DROP KEY fk_td_order;

ALTER TABLE triple_dipper
DROP COLUMN order_id;

ALTER TABLE triple_dipper
ADD user_id SMALLINT UNSIGNED NOT NULL;

ALTER TABLE triple_dipper
ADD CONSTRAINT fk_td_user FOREIGN KEY (user_id)
REFERENCES user (user_id);
