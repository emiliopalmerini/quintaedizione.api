CREATE TABLE IF NOT EXISTS incantesimi (
    id                            VARCHAR(255) PRIMARY KEY,
    nome                          VARCHAR(255) NOT NULL,
    livello                       INTEGER NOT NULL CHECK (livello >= 0 AND livello <= 9),
    scuola_di_magia               VARCHAR(50) NOT NULL,
    tempo_di_lancio               VARCHAR(100) NOT NULL,
    gittata                       VARCHAR(100) NOT NULL,
    area                          VARCHAR(100),
    concentrazione                BOOLEAN NOT NULL DEFAULT FALSE,
    sempre_preparato              BOOLEAN NOT NULL DEFAULT FALSE,
    rituale                       BOOLEAN NOT NULL DEFAULT FALSE,
    componenti                    JSONB NOT NULL DEFAULT '[]',
    componenti_materiali          TEXT,
    durata                        VARCHAR(100) NOT NULL,
    descrizione                   TEXT NOT NULL,
    classi                        VARCHAR(255) NOT NULL,
    documentazione_di_riferimento VARCHAR(50) DEFAULT 'DND 2024',
    created_at                    TIMESTAMP DEFAULT NOW(),
    updated_at                    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incantesimi_nome ON incantesimi(nome);
CREATE INDEX IF NOT EXISTS idx_incantesimi_livello ON incantesimi(livello);
CREATE INDEX IF NOT EXISTS idx_incantesimi_scuola ON incantesimi(scuola_di_magia);
CREATE INDEX IF NOT EXISTS idx_incantesimi_concentrazione ON incantesimi(concentrazione);
CREATE INDEX IF NOT EXISTS idx_incantesimi_rituale ON incantesimi(rituale);
CREATE INDEX IF NOT EXISTS idx_incantesimi_documentazione ON incantesimi(documentazione_di_riferimento);
