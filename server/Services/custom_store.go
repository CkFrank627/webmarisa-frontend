package Services

import (
	"fmt"
	"strings"
	"time"
)

type Persona struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Prompt    string `json:"prompt"`
	CreatedAt int64  `json:"created_at"`
}

type TeachItem struct {
	ID        int64  `json:"id"`
	Q         string `json:"q"`
	A         string `json:"a"`
	CreatedAt int64  `json:"created_at"`
}

func ListPersonas(userID int64) ([]Persona, error) {
	if err := appInit(); err != nil {
		return nil, err
	}
	rows, err := ragDB.Query(`SELECT id, name, prompt, created_at FROM app_personas WHERE user_id=? ORDER BY id DESC`, userID)
	if err != nil { return nil, err }
	defer rows.Close()

	out := []Persona{}
	for rows.Next() {
		var p Persona
		if err := rows.Scan(&p.ID, &p.Name, &p.Prompt, &p.CreatedAt); err != nil { return nil, err }
		out = append(out, p)
	}
	return out, nil
}

func CreatePersona(userID int64, name, prompt string) (int64, error) {
	if err := appInit(); err != nil { return 0, err }
	name = strings.TrimSpace(name)
	prompt = strings.TrimSpace(prompt)
	if name == "" || prompt == "" { return 0, fmt.Errorf("empty name/prompt") }

	res, err := ragDB.Exec(
		`INSERT INTO app_personas(user_id, name, prompt, created_at) VALUES(?,?,?,?)`,
		userID, name, prompt, time.Now().Unix(),
	)
	if err != nil { return 0, err }
	return res.LastInsertId()
}

// 只允许删除自己的 persona
func DeletePersona(userID, personaID int64) error {
	if err := appInit(); err != nil { return err }

	// 校验归属
	var owner int64
	if err := ragDB.QueryRow(`SELECT user_id FROM app_personas WHERE id=?`, personaID).Scan(&owner); err != nil {
		return err
	}
	if owner != userID { return fmt.Errorf("forbidden") }

	// 先删 teach
	_, _ = ragDB.Exec(`DELETE FROM app_persona_teach WHERE persona_id=?`, personaID)
	_, err := ragDB.Exec(`DELETE FROM app_personas WHERE id=?`, personaID)
	return err
}

func ListTeach(userID, personaID int64, limit int) ([]TeachItem, error) {
	if err := appInit(); err != nil { return nil, err }
	if limit <= 0 || limit > 200 { limit = 100 }

	// 校验归属
	var owner int64
	if err := ragDB.QueryRow(`SELECT user_id FROM app_personas WHERE id=?`, personaID).Scan(&owner); err != nil {
		return nil, err
	}
	if owner != userID { return nil, fmt.Errorf("forbidden") }

	rows, err := ragDB.Query(
		`SELECT id, q, a, created_at FROM app_persona_teach WHERE persona_id=? ORDER BY id DESC LIMIT ?`,
		personaID, limit,
	)
	if err != nil { return nil, err }
	defer rows.Close()

	out := []TeachItem{}
	for rows.Next() {
		var t TeachItem
		if err := rows.Scan(&t.ID, &t.Q, &t.A, &t.CreatedAt); err != nil { return nil, err }
		out = append(out, t)
	}
	return out, nil
}

func AddTeach(userID, personaID int64, q, a string) error {
	if err := appInit(); err != nil { return err }

	// 校验归属
	var owner int64
	if err := ragDB.QueryRow(`SELECT user_id FROM app_personas WHERE id=?`, personaID).Scan(&owner); err != nil {
		return err
	}
	if owner != userID { return fmt.Errorf("forbidden") }

	q = strings.TrimSpace(q)
	a = strings.TrimSpace(a)
	if q == "" || a == "" { return fmt.Errorf("empty q/a") }

	_, err := ragDB.Exec(
		`INSERT INTO app_persona_teach(persona_id, q, a, created_at) VALUES(?,?,?,?)`,
		personaID, q, a, time.Now().Unix(),
	)
	return err
}

func DeleteTeachLast(userID, personaID int64) error {
	if err := appInit(); err != nil { return err }

	// 校验归属
	var owner int64
	if err := ragDB.QueryRow(`SELECT user_id FROM app_personas WHERE id=?`, personaID).Scan(&owner); err != nil {
		return err
	}
	if owner != userID { return fmt.Errorf("forbidden") }

	_, err := ragDB.Exec(`
DELETE FROM app_persona_teach
WHERE id = (SELECT id FROM app_persona_teach WHERE persona_id=? ORDER BY id DESC LIMIT 1)`,
		personaID,
	)
	return err
}

func GetPersonaPrompt(userID, personaID int64) (string, error) {
	if err := appInit(); err != nil { return "", err }
	var owner int64
	var prompt string
	err := ragDB.QueryRow(`SELECT user_id, prompt FROM app_personas WHERE id=?`, personaID).Scan(&owner, &prompt)
	if err != nil { return "", err }
	if owner != userID { return "", fmt.Errorf("forbidden") }
	return prompt, nil
}
