-- migrate:up
ALTER TABLE couriers
    ADD CONSTRAINT uq_couriers_phone UNIQUE (phone);

-- migrate:down
ALTER TABLE couriers
    DROP INDEX uq_couriers_phone;
