-- migrate:up
ALTER TABLE buyers
    ADD CONSTRAINT uq_buyers_document UNIQUE (document),
    ADD CONSTRAINT uq_buyers_phone UNIQUE (phone),
    ADD CONSTRAINT uq_buyers_email UNIQUE (email);

-- migrate:down
ALTER TABLE buyers
    DROP INDEX uq_buyers_document,
    DROP INDEX uq_buyers_phone,
    DROP INDEX uq_buyers_email;
