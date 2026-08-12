import glob

sql_to_add = """
CREATE TABLE user_passkeys (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL,
    credential_id BLOB NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type VARCHAR(255) NOT NULL,
    aaguid BLOB NOT NULL,
    sign_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME DEFAULT NULL,
    expires_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE (credential_id)
);
"""

for file in glob.glob("internal/app/dbstart/testdata/original.*.sql"):
    with open(file, "r") as f:
        content = f.read()
    if "user_passkeys" not in content:
        content += sql_to_add
        if "sqlite" in file:
            content = content.replace("AUTO_INCREMENT", "AUTOINCREMENT")
        elif "postgres" in file:
            content = content.replace("INT NOT NULL AUTO_INCREMENT", "SERIAL")
            content = content.replace("BLOB", "BYTEA")
            content = content.replace("DATETIME", "TIMESTAMP")
        with open(file, "w") as f:
            f.write(content)
