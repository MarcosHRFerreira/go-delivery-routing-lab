-- migrate:up
CREATE TABLE deliveries (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_id BIGINT NOT NULL UNIQUE,
    courier_id BIGINT NOT NULL,
    status VARCHAR(30) NOT NULL,
    current_sequence INT,
    last_reordered_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_deliveries_order FOREIGN KEY (order_id) REFERENCES orders(id),
    CONSTRAINT fk_deliveries_courier FOREIGN KEY (courier_id) REFERENCES couriers(id)
);

-- migrate:down
DROP TABLE deliveries;
