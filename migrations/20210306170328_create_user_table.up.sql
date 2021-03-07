CREATE TABLE user (
    user_id SMALLINT UNSIGNED AUTO_INCREMENT,
    first_name VARCHAR(35),
    last_name VARCHAR(35),
    phone CHAR(10),
    email VARCHAR(254),
    CONSTRAINT pk_user PRIMARY KEY (user_id)
);
