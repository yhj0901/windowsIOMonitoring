package monitor

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database는 모니터링 데이터를 저장하기 위한 데이터베이스 연결을 관리합니다.
type Database struct {
	db         *sql.DB
	insertStmt *sql.Stmt
}

// NewDatabase는 새로운 SQLite 데이터베이스 연결을 생성합니다.
func NewDatabase(dbPath string) (*Database, error) {
	// 디렉토리가 없으면 생성
	dir := filepath.Dir(dbPath)
	if err := createDirIfNotExists(dir); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 테이블 생성
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS file_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			path TEXT NOT NULL UNIQUE,
			operation TEXT NOT NULL,
			file_type TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	// 파일 이벤트 삽입 또는 업데이트 준비문 생성
	insertFileStmt, err := db.Prepare(`
		INSERT OR REPLACE INTO file_events (timestamp, path, operation, file_type)
		VALUES (?, ?, ?, ?);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Database{
		db:         db,
		insertStmt: insertFileStmt,
	}, nil
}

// Close는 데이터베이스 연결을 닫습니다.
func (d *Database) Close() error {
	if d.insertStmt != nil {
		d.insertStmt.Close()
	}
	return d.db.Close()
}

// SaveFileEvent는 파일 이벤트를 데이터베이스에 저장합니다.
// 같은 경로의 파일이 이미 존재하면 덮어씁니다.
func (d *Database) SaveFileEvent(event FileEvent) error {
	log.Printf("SaveFileEvent: %v", event)
	_, err := d.insertStmt.Exec(
		event.Timestamp.Format("2006-01-02 15:04:05"),
		event.Path,
		event.Operation,
		event.FileType,
	)
	return err
}

// SaveBatchFileEvents는 여러 파일 이벤트를 일괄적으로 저장합니다.
// 같은 경로의 파일이 이미 존재하면 덮어씁니다.
func (d *Database) SaveBatchFileEvents(events []FileEvent) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO file_events (timestamp, path, operation, file_type)
		VALUES (?, ?, ?, ?);
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		_, err := stmt.Exec(
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Path,
			event.Operation,
			event.FileType,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetFileEventsByTimeRange는 지정된 시간 범위 내의 파일 이벤트를 조회합니다.
func (d *Database) GetFileEventsByTimeRange(start, end time.Time) ([]FileEvent, error) {
	rows, err := d.db.Query(`
		SELECT timestamp, path, operation, file_type 
		FROM file_events 
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC;
	`,
		start.Format("2006-01-02 15:04:05"),
		end.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []FileEvent
	for rows.Next() {
		var event FileEvent
		var timeStr string
		err := rows.Scan(&timeStr, &event.Path, &event.Operation, &event.FileType)
		if err != nil {
			return nil, err
		}
		event.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		events = append(events, event)
	}

	return events, nil
}

// GetFileEvents는 저장된 모든 파일 이벤트를 조회합니다.
func (d *Database) GetFileEvents() ([]FileEvent, error) {
	rows, err := d.db.Query(`
		SELECT timestamp, path, operation, file_type 
		FROM file_events 
		ORDER BY timestamp DESC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []FileEvent
	for rows.Next() {
		var event FileEvent
		var timeStr string
		err := rows.Scan(&timeStr, &event.Path, &event.Operation, &event.FileType)
		if err != nil {
			return nil, err
		}
		event.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		events = append(events, event)
	}

	return events, nil
}

// 디렉토리가 없으면 생성하는 유틸리티 함수
func createDirIfNotExists(dir string) error {
	if dir == "" {
		return nil // 현재 디렉토리인 경우
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("디렉토리 생성 실패: %v", err)
		}
		log.Printf("디렉토리 생성됨: %s", dir)
	}

	return nil
}
