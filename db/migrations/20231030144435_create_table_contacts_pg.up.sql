CREATE TABLE contacts
(
    id         VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name  VARCHAR(100) NULL,
    email      VARCHAR(100) NULL,
    phone      VARCHAR(100) NULL,
    user_id    VARCHAR(100) NOT NULL,
    created_at BIGINT       NOT NULL,
    updated_at BIGINT       NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_contacts_user_id FOREIGN KEY (user_id) REFERENCES users (id)
);
