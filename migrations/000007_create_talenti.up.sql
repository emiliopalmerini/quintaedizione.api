CREATE TABLE IF NOT EXISTS talenti (
    id                            VARCHAR(255) PRIMARY KEY,
    nome                          VARCHAR(255) NOT NULL,
    descrizione                   TEXT,
    documentazione_di_riferimento VARCHAR(50) DEFAULT 'DND 2024',
    created_at                    TIMESTAMP DEFAULT NOW(),
    updated_at                    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_talenti_nome ON talenti(nome);
CREATE INDEX IF NOT EXISTS idx_talenti_documentazione ON talenti(documentazione_di_riferimento);
