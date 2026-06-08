-- migrate:up
CREATE TABLE addresses (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    zip_code VARCHAR(10) NOT NULL,
    street VARCHAR(255) NOT NULL,
    number VARCHAR(20) NOT NULL,
    complement VARCHAR(255),
    district VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    state VARCHAR(10) NOT NULL,
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- migrate:down
DROP TABLE addresses;
