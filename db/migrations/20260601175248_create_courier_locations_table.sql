-- migrate:up
CREATE TABLE courier_locations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    courier_id BIGINT NOT NULL,
    latitude DECIMAL(10,7) NOT NULL,
    longitude DECIMAL(10,7) NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_locations_courier FOREIGN KEY (courier_id) REFERENCES couriers(id)
);

-- migrate:down
DROP TABLE courier_locations;
