CREATE TABLE extra_value (
    extra_value_id SMALLINT UNSIGNED AUTO_INCREMENT,
    extra_value VARCHAR(50) UNIQUE NOT NULL,
    CONSTRAINT pk_extra_value PRIMARY KEY (extra_value_id)
);

INSERT INTO extra_value
    (extra_value)
VALUES
    ('Ancho-Chile Ranch Dressing'),
    ('Avocado-Ranch Dressing'),
    ('Bleu Cheese Dressing'),
    ('Honey-Mustard Dressing'),
    ('Original BBQ Sauce'),
    ('Ranch Dressing');
