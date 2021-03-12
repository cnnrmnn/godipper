ALTER TABLE triple_dipper
DROP CONSTRAINT fk_td_user;

ALTER TABLE triple_dipper
DROP KEY fk_td_user;

ALTER TABLE triple_dipper
DROP COLUMN user_id;

ALTER TABLE triple_dipper
ADD order_id SMALLINT UNSIGNED NOT NULL;

ALTER TABLE triple_dipper
ADD CONSTRAINT fk_td_order FOREIGN KEY (order_id)
REFERENCES orders (order_id);
