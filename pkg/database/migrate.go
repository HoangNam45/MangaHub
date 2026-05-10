package database

func Migrate() {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE,
		password_hash TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	mangaTable := `
	CREATE TABLE IF NOT EXISTS mangas (
		id TEXT PRIMARY KEY,
		title TEXT,
		description TEXT,
		author TEXT,
		genres TEXT,
		status TEXT,
		chapters INTEGER,
		rating REAL
	);`

	userLibraryTable := `
	CREATE TABLE IF NOT EXISTS user_library (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		manga_id TEXT NOT NULL,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
		UNIQUE(user_id, manga_id)
	);`

	readingProgressTable := `
	CREATE TABLE IF NOT EXISTS reading_progress (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		manga_id TEXT NOT NULL,
		current_chapter INTEGER DEFAULT 0,
		progress INTEGER DEFAULT 0,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
		UNIQUE(user_id, manga_id)
	);`

	_, err := DB.Exec(userTable)

	if err != nil {
		panic(err)
	}

	_, err = DB.Exec(mangaTable)
	if err != nil {
		panic(err)
	}

	_, err = DB.Exec(userLibraryTable)
	if err != nil {
		panic(err)
	}

	_, err = DB.Exec(readingProgressTable)
	if err != nil {
		panic(err)
	}
}