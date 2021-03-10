CREATE TABLE item_value (
    item_value_id SMALLINT UNSIGNED AUTO_INCREMENT,
    item_value VARCHAR(50) UNIQUE NOT NULL,
    CONSTRAINT pk_item_value PRIMARY KEY (item_value_id)
);

INSERT INTO item_value 
    (item_value)
VALUES
    ('Awesome Blossom Petals'),
    ('Big Mouth® Bites'),
    ('Boneless Buffalo Wings'),
    ('Boneless Honey-Chipotle Wings'),
    ('Boneless House BBQ Wings'),
    ('Boneless Mango-Habanero Wings'),
    ('Buffalo Wings'),
    ('Crispy Cheddar Bites'),
    ('Crispy Chicken Crispers'),
    ('Crispy Honey-Chipotle Chicken Crispers®'),
    ('Crispy Mango-Habanero Crispers®'),
    ('Fried Pickles'),
    ('Honey-Chipotle Wings'),
    ('House BBQ Wings'),
    ('Mango-Habanero Wings'),
    ('Original Chicken Crispers®'),
    ('Southwestern Eggrolls');
