package models

import (
	"database/sql"
	"errors"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (int, error)
	Get(id int) (*Snippet, error)
	Latest() ([]*Snippet, error)
	Delete(int) error
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	tx, err := m.DB.Begin()

	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	res, err := tx.Exec(stmt, title, content, expires)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0, nil
	}

	err = tx.Commit()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *SnippetModel) Get(id int) (*Snippet, error) {
	tx, err := m.DB.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	stmt := `SELECT * FROM snippets WHERE expires > UTC_TIMESTAMP() AND id = ?`

	s := &Snippet{}

	err = tx.QueryRow(stmt, id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}

		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (m *SnippetModel) Delete(id int) error {
	tx, err := m.DB.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt := `DELETE FROM snippets WHERE id = ?`

	_, err = tx.Exec(stmt, id)

	if err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (m *SnippetModel) Latest() ([]*Snippet, error) {
	tx, err := m.DB.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	snippets := []*Snippet{}

	for rows.Next() {
		// Create a pointer to a new zeroed Snippet struct.
		s := &Snippet{}
		// Use rows.Scan() to copy the values from each field in the row to the
		// new Snippet object that we created. Again, the arguments to row.Scan()
		// must be pointers to the place you want to copy the data into, and the
		// number of arguments must be exactly the same as the number of
		// columns returned by your statement.
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

		if err != nil {
			return nil, err
		}

		// Append it to the slice of snippets.
		snippets = append(snippets, s)
	}

	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return snippets, nil
}
