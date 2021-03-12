CREATE TABLE orders (
    order_id SMALLINT UNSIGNED AUTO_INCREMENT,
    user_id SMALLINT UNSIGNED NOT NULL,
    session_id SMALLINT UNSIGNED,
    completed BOOLEAN,
    subtotal FLOAT,
    tax FLOAT,
    delivery_fee FLOAT,
    service_fee FLOAT,
    delivery_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT pk_order PRIMARY KEY (order_id),
    CONSTRAINT fk_order_user FOREIGN KEY (user_id)
    REFERENCES user (user_id) 
);
