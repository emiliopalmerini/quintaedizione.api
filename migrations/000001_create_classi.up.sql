CREATE TABLE IF NOT EXISTS classi (
    id                            VARCHAR(255) PRIMARY KEY,
    nome                          VARCHAR(255) NOT NULL,
    descrizione                   TEXT,
    documentazione_di_riferimento VARCHAR(50) DEFAULT 'DND 2024',
    dado_vita                     VARCHAR(10) NOT NULL,
    equipaggiamento_partenza      JSONB,
    proprieta_di_classe           JSONB,
    created_at                    TIMESTAMP DEFAULT NOW(),
    updated_at                    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_classi_nome ON classi(nome);
CREATE INDEX IF NOT EXISTS idx_classi_documentazione ON classi(documentazione_di_riferimento);
