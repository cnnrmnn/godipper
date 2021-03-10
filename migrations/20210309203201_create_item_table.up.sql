CREATE TABLE item (
    item_id SMALLINT UNSIGNED AUTO_INCREMENT,
    triple_dipper_id SMALLINT UNSIGNED NOT NULL,
    item SMALLINT UNSIGNED NOT NULL,
    CONSTRAINT pk_item PRIMARY KEY (item_id),
    CONSTRAINT fk_triple_dipper FOREIGN KEY (triple_dipper_id)
    REFERENCES triple_dipper (triple_dipper_id)
);
