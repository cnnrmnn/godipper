CREATE TABLE item_extra_combination (
    combination_id SMALLINT UNSIGNED AUTO_INCREMENT,
    item_value_id SMALLINT UNSIGNED NOT NULL,
    extra_value_id SMALLINT UNSIGNED NOT NULL,
    CONSTRAINT pk_item_extra_combination PRIMARY KEY (combination_id),
    CONSTRAINT fk_combination_item FOREIGN KEY (item_value_id)
    REFERENCES item_value (item_value_id),
    CONSTRAINT fk_combination_extra FOREIGN KEY (extra_value_id)
    REFERENCES extra_value (extra_value_id)
);


INSERT INTO item_extra_combination
    (item_value_id, extra_value_id)
VALUES
    (1, 2),
    (1, 6),
    (2, 6),
    (3, 3),
    (3, 6),
    (4, 3),
    (4, 6),
    (5, 3),
    (5, 6),
    (6, 3),
    (6, 6),
    (7, 3),
    (7, 6),
    (8, 1),
    (9, 4),
    (9, 5),
    (9, 6),
    (10, 4),
    (10, 5),
    (10, 6),
    (11, 5),
    (11, 5),
    (11, 6),
    (12, 2),
    (12, 6),
    (13, 3),
    (13, 6),
    (14, 3),
    (14, 6),
    (15, 3),
    (15, 6),
    (16, 4),
    (16, 5),
    (16, 6),
    (17, 2),
    (17, 6);
