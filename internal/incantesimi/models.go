package incantesimi

import "encoding/json"

type ScuolaDiMagia string

const (
	Abiurazione    ScuolaDiMagia = "Abiurazione"
	Divinazione    ScuolaDiMagia = "Divinazione"
	Evocazione     ScuolaDiMagia = "Evocazione"
	Invocazione    ScuolaDiMagia = "Invocazione"
	Necromamzia    ScuolaDiMagia = "Necromamzia"
	Illusione      ScuolaDiMagia = "Illusione"
	Transmutazione ScuolaDiMagia = "Transmutazione"
	Incantamento   ScuolaDiMagia = "Incantamento"
)

type Componente string

const (
	Verbale   Componente = "V"
	Somatica  Componente = "S"
	Materiale Componente = "M"
)

type EffettoIncantesimo struct {
	RipetizioneEffetto *int             `json:"ripetizione-effetto,omitempty"`
	Effetto            *json.RawMessage `json:"effetto,omitempty"`
}

type Incantesimo struct {
	ID                          string              `json:"id" db:"id"`
	Nome                        string              `json:"nome" db:"nome"`
	Livello                     int32               `json:"livello" db:"livello"`
	ScuolaDiMagia               ScuolaDiMagia       `json:"scuola-di-magia" db:"scuola_di_magia"`
	TempoDiLancio               string              `json:"tempo-di-lancio" db:"tempo_di_lancio"`
	Gittata                     string              `json:"gittata" db:"gittata"`
	Area                        string              `json:"area,omitempty" db:"area"`
	Concentrazione              bool                `json:"concentrazione" db:"concentrazione"`
	SemprePreparato             bool                `json:"sempre-preparato" db:"sempre_preparato"`
	Rituale                     bool                `json:"rituale" db:"rituale"`
	EffettoIncantesimo          *EffettoIncantesimo `json:"effetto-incantesimo,omitempty"`
	Componenti                  []Componente        `json:"componenti" db:"componenti"`
	ComponentiMateriali         string              `json:"componenti-materiali,omitempty" db:"componenti_materiali"`
	Durata                      string              `json:"durata" db:"durata"`
	Descrizione                 string              `json:"descrizione" db:"descrizione"`
	EffettoLivelloMaggiore      *EffettoIncantesimo `json:"effetto-livello-maggiore,omitempty"`
	Classi                      string              `json:"classi" db:"classi"`
	DocumentazioneDiRiferimento string              `json:"documentazione-di-riferimento" db:"documentazione_di_riferimento"`
}
