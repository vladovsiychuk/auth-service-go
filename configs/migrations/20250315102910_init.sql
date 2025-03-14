-- +goose up
CREATE TABLE keys (
    private_key text NOT NULL,
    public_key text NOT NULL,
    secondary_public_key text NOT NULL
);

CREATE TABLE session_tokens (
    id uuid PRIMARY KEY,
    email varchar(100),
    expires_at timestamptz NOT NULL
);