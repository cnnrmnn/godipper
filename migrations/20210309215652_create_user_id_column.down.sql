ALTER TABLE triple_dipper
DROP CONSTRAINT fk_td_user;

ALTER TABLE triple_dipper
DROP KEY fk_td_user;

ALTER TABLE triple_dipper 
DROP COLUMN user_id;
