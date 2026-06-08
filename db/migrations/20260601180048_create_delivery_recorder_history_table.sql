-- migrate:up
CREATE TABLE delivery_reorder_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    courier_id BIGINT NOT NULL,
    delivery_id BIGINT NOT NULL,
    sequence_position INT NOT NULL,
    score DECIMAL(10,4),
    reason VARCHAR(255),
    generated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_reorder_history_courier FOREIGN KEY (courier_id) REFERENCES couriers(id),
    CONSTRAINT fk_reorder_history_delivery FOREIGN KEY (delivery_id) REFERENCES deliveries(id) 
);

-- migrate:down
DROP TABLE delivery_reorder_history;
