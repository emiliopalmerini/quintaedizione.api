# Quinta Edizione API - Classi

API REST per la gestione delle classi di D&D 5a Edizione.

## Requisiti

- Go 1.25+
- Docker e Docker Compose
- PostgreSQL 16+

## Setup

1. Copia il file di configurazione:
```bash
cp .env.example .env
```

2. Avvia PostgreSQL:
```bash
make docker-up
```

3. Avvia l'API:
```bash
make run
```

## Comandi Make

| Comando | Descrizione |
|---------|-------------|
| `make build` | Compila l'applicazione |
| `make run` | Avvia l'API (con PostgreSQL) |
| `make test` | Esegue i test unitari |
| `make test-e2e` | Esegue i test end-to-end |
| `make docker-up` | Avvia PostgreSQL |
| `make docker-down` | Ferma i container |

## Endpoints

Base URL: `http://localhost:8080`

| Metodo | Endpoint | Descrizione |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/swagger` | Documentazione OpenAPI |
| GET | `/v1/classi` | Lista delle classi |
| GET | `/v1/classi/{id}` | Dettaglio classe |
| GET | `/v1/classi/{id}/sotto-classi` | Lista sottoclassi |
| GET | `/v1/classi/{id}/sotto-classi/{id}` | Dettaglio sottoclasse |

### Query Parameters

| Parametro | Tipo | Descrizione |
|-----------|------|-------------|
| `nome` | string | Filtra per nome (max 100 char) |
| `sort` | string | Ordinamento: `asc` o `desc` |
| `$limit` | int | Elementi per pagina (1-100, default: 20) |
| `$offset` | int | Offset paginazione |

## Variabili d'Ambiente

Vedi `.env.example` per la lista completa.
