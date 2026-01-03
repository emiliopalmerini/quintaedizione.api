CREATE TABLE IF NOT EXISTS oggetti (
    id                            VARCHAR(255) PRIMARY KEY,
    nome                          VARCHAR(255) NOT NULL,
    descrizione                   TEXT,
    documentazione_di_riferimento VARCHAR(50) DEFAULT 'DND 2024',
    created_at                    TIMESTAMP DEFAULT NOW(),
    updated_at                    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_oggetti_nome ON oggetti(nome);
CREATE INDEX IF NOT EXISTS idx_oggetti_documentazione ON oggetti(documentazione_di_riferimento);
