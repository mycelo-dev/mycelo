CREATE TABLE EVENTS (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    topic varchar(255),
    event_data jsonb,
    created_at bigint
);