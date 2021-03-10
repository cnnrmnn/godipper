ALTER TABLE triple_dipper
ADD user_id SMALLINT UNSIGNED NOT NULL;

ALTER TABLE triple_dipper
ADD CONSTRAINT fk_td_user FOREIGN KEY (user_id)
REFERENCES user(user_id)
