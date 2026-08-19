CREATE TABLE processed_events (
    order_id UUID PRIMARY KEY,
    saga_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW()
);
