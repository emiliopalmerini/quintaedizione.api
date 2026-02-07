package classi

type TipoDiDado string

const (
	D3  TipoDiDado = "d3"
	D4  TipoDiDado = "d4"
	D6  TipoDiDado = "d6"
	D8  TipoDiDado = "d8"
	D10 TipoDiDado = "d10"
	D12 TipoDiDado = "d12"
	D20 TipoDiDado = "d20"
)

type TipoAzione string

const (
	Nessuna        TipoAzione = "Nessuna"
	AzioneBonus    TipoAzione = "Azione Bonus"
	Azione         TipoAzione = "Azione"
	Reazione       TipoAzione = "Reazione"
	AzioneGratuita TipoAzione = "Azione Gratuita"
)

type Tratto struct {
	ID             string     `json:"id,omitempty"`
	Nome           string     `json:"nome"`
	Descrizione    string     `json:"descrizione,omitempty"`
	TipoAzione     TipoAzione `json:"tipo-azione,omitempty"`
	TipoDiSorgente string     `json:"tipo-di-sorgente,omitempty"`
}

type SlotIncantesimo struct {
	NumeroSlot             int32 `json:"numero-slot"`
	LivelloSlotIncantesimo int32 `json:"livello-slot-incantesimo"`
}

type IncantesimiClasse struct {
	SlotIncantesimi      []SlotIncantesimo `json:"slot-incantesimi,omitempty"`
	IncantesimiPreparati int32             `json:"incantesimi-preparati,omitempty"`
}

type ProprietaLivello struct {
	LivelloClasse     int32              `json:"livello-classe"`
	TrattoDiClasse    *Tratto            `json:"tratto-di-classe,omitempty"`
	IncantesimiClasse *IncantesimiClasse `json:"incantesimi-di-classe,omitempty"`
}

type RiferimentoSottoclasse struct {
	IDSottoclasse string `json:"id-sottoclasse"`
}

type Valuta string

const (
	MR Valuta = "MR"
	MA Valuta = "MA"
	ME Valuta = "ME"
	MO Valuta = "MO"
	MP Valuta = "MP"
)

type Importo struct {
	Quantita int32  `json:"quantità"`
	Valuta   Valuta `json:"valuta"`
}

type OggettoPartenza struct {
	ID       string `json:"id,omitempty"`
	Nome     string `json:"nome,omitempty"`
	Quantita int32  `json:"quantità,omitempty"`
}

type EquipaggiamentoPartenza struct {
	OpzioneA []OggettoPartenza `json:"opzione-a,omitempty"`
	OpzioneB *Importo          `json:"opzione-b,omitempty"`
}

type Classe struct {
	ID                          string                   `json:"id" db:"id"`
	Nome                        string                   `json:"nome" db:"nome"`
	Descrizione                 string                   `json:"descrizione" db:"descrizione"`
	DocumentazioneDiRiferimento string                   `json:"documentazione-di-riferimento" db:"documentazione_di_riferimento"`
	DadoVita                    TipoDiDado               `json:"dado-vita" db:"dado_vita"`
	ElencoSottoclassi           []RiferimentoSottoclasse `json:"elenco-sottoclassi,omitempty"`
	EquipaggiamentoPartenza     *EquipaggiamentoPartenza `json:"equipaggiamento-id-partenza,omitempty"`
	ProprietaDiClasse           []ProprietaLivello       `json:"proprietà-di-classe,omitempty"`
}

type SottoClasse struct {
	ID                          string             `json:"id" db:"id"`
	Nome                        string             `json:"nome" db:"nome"`
	Descrizione                 string             `json:"descrizione" db:"descrizione"`
	DocumentazioneDiRiferimento string             `json:"documentazione-di-riferimento" db:"documentazione_di_riferimento"`
	IDClasseAssociata           string             `json:"id-classe-associata" db:"id_classe_associata"`
	ProprietaDiSottoclasse      []ProprietaLivello `json:"proprietà-di-sottoclasse,omitempty"`
}

