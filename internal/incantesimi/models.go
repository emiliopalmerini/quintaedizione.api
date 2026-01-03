package incantesimi

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type ScuolaDiMagia string

const (
	Abiurazione    ScuolaDiMagia = "Abiurazione"
	Divinazione    ScuolaDiMagia = "Divinazione"
	Evocazione     ScuolaDiMagia = "Evocazione"
	Invocazione    ScuolaDiMagia = "Invocazione"
	Necromanzia    ScuolaDiMagia = "Necromanzia"
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

type Incantesimo struct {
	ID                          string            `json:"id" db:"id"`
	Nome                        string            `json:"nome" db:"nome"`
	Livello                     int               `json:"livello" db:"livello"`
	ScuolaDiMagia               ScuolaDiMagia     `json:"scuola-di-magia" db:"scuola_di_magia"`
	TempoDiLancio               string            `json:"tempo-di-lancio" db:"tempo_di_lancio"`
	Gittata                     string            `json:"gittata" db:"gittata"`
	Area                        string            `json:"area,omitempty" db:"area"`
	Concentrazione              bool              `json:"concentrazione" db:"concentrazione"`
	SemprePreparato             bool              `json:"sempre-preparato,omitempty" db:"sempre_preparato"`
	Rituale                     bool              `json:"rituale" db:"rituale"`
	Componenti                  ComponentiSlice   `json:"componenti" db:"componenti"`
	ComponentiMateriali         string            `json:"componenti-materiali,omitempty" db:"componenti_materiali"`
	Durata                      string            `json:"durata" db:"durata"`
	Descrizione                 string            `json:"descrizione" db:"descrizione"`
	Classi                      string            `json:"classi" db:"classi"`
	DocumentazioneDiRiferimento string            `json:"documentazione-di-riferimento" db:"documentazione_di_riferimento"`
}

type ComponentiSlice []Componente

func (c *ComponentiSlice) Scan(src any) error {
	if src == nil {
		*c = nil
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion failed for ComponentiSlice")
	}
	return json.Unmarshal(source, c)
}

func (c ComponentiSlice) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
