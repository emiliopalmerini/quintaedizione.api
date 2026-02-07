# Quinta Edizione API

API REST per D&D 5a Edizione, scritta in Go.

## 1. Setup e sviluppo locale

### Requisiti

- Go 1.25+
- Docker e Docker Compose

### Avvio rapido

```bash
# 1. Copia il file di configurazione
cp .env.example .env

# 2. Avvia PostgreSQL e pgAdmin
make docker-up

# 3. Avvia l'API
make run
```

L'API sarà disponibile su `http://localhost:8080`, pgAdmin su `http://localhost:5050`.

### Comandi Make

| Comando            | Descrizione                        |
| ------------------ | ---------------------------------- |
| `make build`       | Compila l'applicazione             |
| `make run`         | Avvia l'API (con PostgreSQL)       |
| `make test`        | Esegue i test unitari              |
| `make test-e2e`    | Esegue i test end-to-end           |
| `make bench`       | Esegue i benchmark                 |
| `make loadtest`    | Esegue il load test                |
| `make docker-up`   | Avvia i container                  |
| `make docker-down` | Ferma i container                  |
| `make fmt`         | Formatta il codice                 |
| `make vet`         | Analisi statica                    |
| `make clean`       | Rimuove la directory bin           |

### Configurazione

La configurazione avviene tramite variabili d'ambiente. Vedi `.env.example` per la lista completa.

| Variabile | Default | Descrizione |
| --- | --- | --- |
| `API_PORT` | `8080` | Porta del server |
| `DATABASE_URL` | — | URL di connessione PostgreSQL (obbligatorio) |
| `API_KEY` | — | Chiave API (obbligatoria in produzione) |
| `RATE_LIMIT_RPM` | `60` | Richieste al minuto per IP |
| `RATE_LIMIT_ENABLED` | `true` | Abilita/disabilita rate limiting |
| `CORS_ALLOWED_ORIGINS` | `*` | Origini CORS consentite |

### Autenticazione

Gli endpoint `/v1/*` richiedono l'header `X-API-Key` se `API_KEY` è configurata.

```bash
curl -H "X-API-Key: your-secret-key" http://localhost:8080/v1/classi
```

Per disabilitare l'autenticazione in sviluppo, lasciare `API_KEY` vuoto in `.env`.

## 2. Modulo classi (implementazione di riferimento)

Il modulo `classi` è il primo modulo di dominio e funge da esempio per i futuri moduli. Segue l'architettura esagonale:

```
internal/classi/
├── models.go          # tipi di dominio (Classe, SottoClasse, enum)
├── interfaces.go      # porta Repository (interfaccia)
├── service.go         # logica di business
├── errors.go          # errori di dominio
├── responses.go       # DTO di risposta API
├── persistence/
│   └── postgres.go    # adapter Repository (PostgreSQL + JSONB)
└── transports/
    └── http.go        # adapter HTTP (handler chi)
```

Il dominio definisce l'interfaccia `Repository`; la persistenza la implementa. Il service e gli handler HTTP dipendono da interfacce, non da implementazioni concrete.

### Endpoint

| Metodo | Endpoint                            | Descrizione           |
| ------ | ----------------------------------- | --------------------- |
| GET    | `/health`                           | Health check          |
| GET    | `/swagger`                          | Documentazione OpenAPI|
| GET    | `/v1/classi`                        | Lista classi          |
| GET    | `/v1/classi/{id}`                   | Dettaglio classe      |
| GET    | `/v1/classi/{id}/sotto-classi`      | Lista sottoclassi     |
| GET    | `/v1/classi/{id}/sotto-classi/{id}` | Dettaglio sottoclasse |

### Query Parameters

| Parametro | Tipo   | Descrizione                              |
| --------- | ------ | ---------------------------------------- |
| `nome`    | string | Filtra per nome (max 100 char)           |
| `sort`    | string | Ordinamento: `asc` o `desc`              |
| `$limit`  | int    | Elementi per pagina (1-100, default: 20) |
| `$offset` | int    | Offset paginazione                       |

### Test

```bash
# Test unitari
make test

# Test end-to-end (richiede Docker)
make test-e2e

# Benchmark
make bench

# Load test
make loadtest
```
