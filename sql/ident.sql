DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
        user_id VARCHAR(256) NOT NULL,
        password VARCHAR(512) NOT NULL,
        totp_secret VARCHAR(512) NOT NULL,
        email VARCHAR(256) NOT NULL,
        PRIMARY KEY (user_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
