package talenti

type Talento struct {
	ID                          string `json:"id" db:"id"`
	Nome                        string `json:"nome" db:"nome"`
	Descrizione                 string `json:"descrizione,omitempty" db:"descrizione"`
	DocumentazioneDiRiferimento string `json:"documentazione-di-riferimento" db:"documentazione_di_riferimento"`
}
