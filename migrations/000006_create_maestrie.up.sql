CREATE TABLE IF NOT EXISTS maestrie (
    id                            VARCHAR(255) PRIMARY KEY,
    nome                          VARCHAR(255) NOT NULL,
    descrizione                   TEXT,
    documentazione_di_riferimento VARCHAR(50) DEFAULT 'DND 2024',
    created_at                    TIMESTAMP DEFAULT NOW(),
    updated_at                    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_maestrie_nome ON maestrie(nome);
CREATE INDEX IF NOT EXISTS idx_maestrie_documentazione ON maestrie(documentazione_di_riferimento);
