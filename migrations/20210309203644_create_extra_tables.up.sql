CREATE TABLE extra (
    extra_id SMALLINT UNSIGNED AUTO_INCREMENT,
    item_id SMALLINT UNSIGNED NOT NULL,
    extra SMALLINT UNSIGNED NOT NULL,
    CONSTRAINT pk_extra PRIMARY KEY (extra_id),
    CONSTRAINT fk_item FOREIGN KEY (item_id)
    REFERENCES item (item_id)
);
