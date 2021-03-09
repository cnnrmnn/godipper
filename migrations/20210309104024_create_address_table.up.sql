CREATE TABLE address (
    address_id SMALLINT UNSIGNED AUTO_INCREMENT,
    user_id SMALLINT UNSIGNED NOT NULL,
    street VARCHAR(50) NOT NULL,
    unit VARCHAR(15),
    city VARCHAR(25) NOT NULL,
    state CHAR(2) NOT NULL,
    zip CHAR(5) NOT NULL,
    CONSTRAINT pk_address PRIMARY KEY (address_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id)
    REFERENCES user (user_id)
);
