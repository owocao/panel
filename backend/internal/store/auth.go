package store

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

func (s *Store) HasUser() (bool, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n > 0, err
}

func (s *Store) CreateUser(username, hash string) error {
	_, err := s.DB.Exec(`INSERT INTO users(username,password_hash,created_at) VALUES(?,?,?)`, username, hash, time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) FindUser(username string) (User, error) {
	var u User
	err := s.DB.QueryRow(`SELECT id,username,password_hash FROM users WHERE username=?`, username).Scan(&u.ID, &u.Username, &u.PasswordHash)
	return u, err
}

func (s *Store) LogLogin(username string, success bool, ip, msg string) {
	val := 0
	if success {
		val = 1
	}
	_, _ = s.DB.Exec(`INSERT INTO login_logs(username,success,ip,message,created_at) VALUES(?,?,?,?,?)`, username, val, ip, msg, time.Now().Format(time.RFC3339))
}

func (s *Store) FailedLoginsSince(username string, since time.Time) (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM login_logs WHERE username=? AND success=0 AND created_at>?`, username, since.Format(time.RFC3339)).Scan(&n)
	return n, err
}

func (s *Store) CleanupExpiredData(logsRetentionDays int) error {
	nowStr := time.Now().Format(time.RFC3339)
	oldLogsStr := time.Now().AddDate(0, 0, -logsRetentionDays).Format(time.RFC3339)

	// Clean expired sessions
	_, err := s.DB.Exec(`DELETE FROM sessions WHERE expires_at != '' AND expires_at < ?`, nowStr)
	if err != nil {
		log.Printf("[Store] Failed to cleanup expired sessions: %v\n", err)
	}

	// Clean old login logs
	_, err = s.DB.Exec(`DELETE FROM login_logs WHERE created_at < ?`, oldLogsStr)
	if err != nil {
		log.Printf("[Store] Failed to cleanup old login_logs: %v\n", err)
	}

	return nil
}

func (s *Store) SaveSession(token string, userID int64, expires *time.Time, remember bool) error {
	rem := 0
	var exp any
	if remember {
		rem = 1
	}
	if expires != nil {
		exp = expires.Format(time.RFC3339)
	}
	_, err := s.DB.Exec(`INSERT INTO sessions(token,user_id,expires_at,remember,created_at) VALUES(?,?,?,?,?)`, token, userID, exp, rem, time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) DeleteSession(token string) {
	_, _ = s.DB.Exec(`DELETE FROM sessions WHERE token=?`, token)
}

func (s *Store) UserBySession(token string) (User, error) {
	var u User
	var exp sql.NullString
	err := s.DB.QueryRow(`SELECT u.id,u.username,u.password_hash,s.expires_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token=?`, token).Scan(&u.ID, &u.Username, &u.PasswordHash, &exp)
	if err != nil {
		return u, err
	}
	if exp.Valid {
		t, _ := time.Parse(time.RFC3339, exp.String)
		if time.Now().After(t) {
			s.DeleteSession(token)
			return u, errors.New("session expired")
		}
	}
	return u, nil
}
