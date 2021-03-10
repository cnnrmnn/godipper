CREATE TABLE item_value (
    item_value_id SMALLINT UNSIGNED AUTO_INCREMENT,
    item_value VARCHAR(50) UNIQUE NOT NULL,
    CONSTRAINT pk_item_value PRIMARY KEY (item_value_id)
);
