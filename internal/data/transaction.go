package data

import (
	"context"
	"database/sql"
)

func (r *BookRepositoryImpl) InitPreparedStatements() error {
	var err error

	// Prepared insert statement for books
	r.insertBookStmt, err = r.DB.Prepare(`
		INSERT INTO books (title, subtitle, description, language, page_count, publish_date, image_link, notes, tags, created_at, last_updated, isbn_10, isbn_13)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetBookByID
	r.getBookByIDStmt, err = r.DB.Prepare(`
		WITH book_data AS (
				SELECT id, title, subtitle, description, language, page_count, publish_date,
							image_link, notes, tags, created_at, last_updated, isbn_10, isbn_13
				FROM books
				WHERE id = $1
		),
		authors_data AS (
				SELECT a.name, ba.book_id
				FROM authors a
				JOIN book_authors ba ON a.id = ba.author_id
				WHERE ba.book_id = $1
		),
		genres_data AS (
				SELECT g.name, bg.book_id
				FROM genres g
				JOIN book_genres bg ON g.id = bg.genre_id
				WHERE bg.book_id = $1
		),
		formats_data AS (
				SELECT f.format_type, bf.book_id
				FROM formats f
				JOIN book_formats bf ON f.id = bf.format_id
				WHERE bf.book_id = $1
		)
		SELECT
				b.id, b.title, b.subtitle, b.description, b.language, b.page_count, b.publish_date,
				b.image_link, b.notes, b.tags, b.created_at, b.last_updated, b.isbn_10, b.isbn_13,
				COALESCE(a.name, '') AS author_name, COALESCE(g.name, '') AS genre_name, COALESCE(f.format_type, '') AS format_type
		FROM book_data b
		LEFT JOIN authors_data a ON b.id = a.book_id
		LEFT JOIN genres_data g ON b.id = g.book_id
		LEFT JOIN formats_data f ON b.id = f.book_id
		`)
	if err != nil {
		return err
	}

	// Prepared insert statement for adding books to user
	r.addBookToUserStmt, err = r.DB.Prepare(`INSERT INTO user_books (user_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}

	// Prepared select statement for getting book ID by title
	r.getBookIdByTitleStmt, err = r.DB.Prepare(`SELECT id FROM books WHERE title = $1`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetAuthorsForBooks
	r.getAuthorsForBooksStmt, err = r.DB.Prepare(`
		SELECT ba.book_id, a.name
		FROM authors a
		JOIN book_authors ba ON a.id = ba.author_id
		WHERE ba.book_id = ANY($1)`)
	if err != nil {
		return err
	}

	// Prepared select statement for IsUserBookOwner
	r.isUserBookOwnerStmt, err = r.DB.Prepare(`
		SELECT EXISTS(SELECT 1 FROM user_books WHERE user_id = $1 AND book_id = $2)`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetAllBooksByUserID
	r.getAllBooksByUserIDStmt, err = r.DB.Prepare(`
		SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
					 r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13
		FROM books b
		INNER JOIN user_books ub ON r.id = ub.book_id
		WHERE ub.user_id = $1`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetAllBooksByAuthors
	r.getAllBooksByAuthorsStmt, err = r.DB.Prepare(`
	SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
				 r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13,
				 a.name AS author_name,
				 json_agg(DISTINCT g.name) AS genres,
				 json_agg(DISTINCT f.format_type) AS formats,
				 r.tags
	FROM books b
	INNER JOIN book_authors ba ON r.id = ba.book_id
	INNER JOIN authors a ON ba.author_id = a.id
	INNER JOIN user_books ub ON r.id = ub.book_id
	LEFT JOIN book_genres bg ON r.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	LEFT JOIN book_formats bf ON r.id = bf.book_id
	LEFT JOIN formats f ON bf.format_id = f.id
	WHERE ub.user_id = $1::integer  -- Explicitly cast the user_id to integer
	GROUP BY r.id, a.name`)
	if err != nil {
		return err
	}

	// Prepared statment for GetAllBooksByGenres
	r.getAllBooksByGenresStmt, err = r.DB.Prepare(`
	SELECT r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
								 r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13,
								 json_agg(DISTINCT g.name) AS genres,
								 json_agg(DISTINCT a.name) AS authors,
								 json_agg(DISTINCT f.format_type) AS formats,
								 r.tags
					FROM books b
					INNER JOIN user_books ub ON r.id = ub.book_id
					LEFT JOIN book_genres bg ON r.id = bg.book_id
					LEFT JOIN genres g ON bg.genre_id = g.id
					LEFT JOIN book_authors ba ON r.id = ba.book_id
					LEFT JOIN authors a ON ba.author_id = a.id
					LEFT JOIN book_formats bf ON r.id = bf.book_id
					LEFT JOIN formats f ON bf.format_id = f.id
					WHERE ub.user_id = $1
					GROUP BY r.id`)
	if err != nil {
		return err
	}

	// Prepared statement for GetAllBooksByFormat
	r.getAllBooksByFormatStmt, err = r.DB.Prepare(`
	SELECT
		r.id, r.title, r.subtitle, r.description, r.language, r.page_count, r.publish_date,
		r.image_link, r.notes, r.created_at, r.last_updated, r.isbn_10, r.isbn_13,
		f.format_type,
		array_to_json(array_agg(DISTINCT a.name)) as authors,
		array_to_json(array_agg(DISTINCT g.name)) as genres,
		r.tags
	FROM books b
	INNER JOIN book_formats bf ON r.id = bf.book_id
	INNER JOIN formats f ON bf.format_id = f.id
	INNER JOIN user_books ub ON r.id = ub.book_id
	LEFT JOIN book_authors ba ON r.id = ba.book_id
	LEFT JOIN authors a ON ba.author_id = a.id
	LEFT JOIN book_genres bg ON r.id = bg.book_id
	LEFT JOIN genres g ON bg.genre_id = g.id
	WHERE ub.user_id = $1
	GROUP BY r.id, f.format_type`)
	if err != nil {
	return err
	}

	// Prepared select statement for GetFormats
	r.getFormatsStmt, err = r.DB.Prepare(`
		SELECT f.format_type
		FROM formats f
		JOIN book_formats bf ON f.id = bf.format_id
		WHERE bf.book_id = $1`)
	if err != nil {
		return err
	}

	// Prepared select statement for GetGenres
	r.getGenresStmt, err = r.DB.Prepare(`
		SELECT g.name
		FROM genres g
		JOIN book_genres bg ON g.id = bg.genre_id
		WHERE bg.book_id = $1`)
	if err != nil {
		return err
	}

	// Prepared statment for GetLanguages
	r.getAllLangStmt, err = r.DB.Prepare(`
	SELECT language, COUNT(*) AS total
		FROM books
		INNER JOIN user_books ub ON books.id = ub.book_id
		WHERE ub.user_id = $1
		GROUP BY language
		ORDER BY total DESC`)
	if err != nil {
		return err
	}

	r.getBookListByGenreStmt, err = r.DB.Prepare(`
	SELECT g.name, COUNT(DISTINCT r.id) AS total_books
		FROM books b
		INNER JOIN book_genres bg ON r.id = bg.book_id
		INNER JOIN genres g ON bg.genre_id = g.id
		INNER JOIN user_books ub ON r.id = ub.book_id
		WHERE ub.user_id = $1
		GROUP BY g.name
		ORDER BY total_books DESC`)
	if err != nil {
		return err
	}

	r.getUserTagsStmt, err = r.DB.Prepare(`
	SELECT r.tags
		FROM books b
		INNER JOIN user_books ub ON r.id = ub.book_id
		WHERE ub.user_id = $1
	`)
	if err != nil {
		return err
	}

	return nil
}

func (r *BookRepositoryImpl) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		r.Logger.Error("Error beginning transaction", "error", err)
		return nil, err
	}
	return tx, nil
}