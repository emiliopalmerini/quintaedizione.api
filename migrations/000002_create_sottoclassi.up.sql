CREATE TABLE IF NOT EXISTS sottoclassi (
    id                            VARCHAR(255) PRIMARY KEY,
    nome                          VARCHAR(255) NOT NULL,
    descrizione                   TEXT,
    documentazione_di_riferimento VARCHAR(50) DEFAULT 'DND 2024',
    id_classe_associata           VARCHAR(255) NOT NULL REFERENCES classi(id) ON DELETE CASCADE,
    proprieta_di_sottoclasse      JSONB,
    created_at                    TIMESTAMP DEFAULT NOW(),
    updated_at                    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sottoclassi_classe ON sottoclassi(id_classe_associata);
CREATE INDEX IF NOT EXISTS idx_sottoclassi_nome ON sottoclassi(nome);
