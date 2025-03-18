package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileEvent는 파일 이벤트 정보를 저장하는 구조체입니다.
type FileEvent struct {
	Path      string
	Operation string
	Timestamp time.Time
	FileType  string
}

// Monitor는 파일 모니터링을 담당하는 구조체입니다.
type Monitor struct {
	interval    time.Duration
	devices     []string
	fileEvents  []FileEvent
	running     bool
	watcher     *fsnotify.Watcher
	watchMutex  sync.Mutex
	fileFilters []string
	db          *Database
	dbPath      string
	saveTimer   *time.Ticker
	eventsMutex sync.Mutex
}

// NewMonitor는 새로운 모니터 인스턴스를 생성합니다.
func NewMonitor(interval time.Duration) *Monitor {
	return &Monitor{
		interval:    interval,
		devices:     []string{},
		fileEvents:  []FileEvent{},
		running:     false,
		fileFilters: []string{".exe", ".dll"}, // 기본 필터
		dbPath:      "monitor.db",             // 기본 데이터베이스 경로
	}
}

// AddDevice는 모니터링할 장치를 추가합니다.
func (m *Monitor) AddDevice(device string) {
	m.devices = append(m.devices, device)
	log.Printf("장치 추가됨: %s", device)
}

// SetFileFilters는 모니터링할 파일 확장자 필터를 설정합니다.
func (m *Monitor) SetFileFilters(filters []string) {
	m.fileFilters = filters
	log.Printf("파일 필터 설정됨: %v", filters)
}

// SetDatabasePath는 데이터베이스 파일 경로를 설정합니다.
func (m *Monitor) SetDatabasePath(path string) {
	m.dbPath = path
}

// watchRecursive는 디렉터리를 재귀적으로 watcher에 등록하는 함수입니다.
func (m *Monitor) watchRecursive(path string) error {
	log.Printf("재귀적 감시 시작: %s\n", path)
	count := 0

	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("접근 권한 오류: %s - %v\n", walkPath, err)
			// 접근 권한이 없는 폴더는 스킵
			return nil
		}

		if info.IsDir() {
			m.watchMutex.Lock()
			err = m.watcher.Add(walkPath)
			m.watchMutex.Unlock()
			if err == nil {
				count++
				// 디렉토리 100개마다 로그 출력 (너무 많은 로그 방지)
				if count%100 == 0 {
					log.Printf("감시 디렉토리 %d개 추가됨 (최근: %s)\n", count, walkPath)
				}
			} else {
				log.Printf("디렉토리 감시 추가 실패: %s - %v\n", walkPath, err)
			}
		}
		return nil
	})

	log.Printf("재귀적 감시 설정 완료: %s (총 %d개 디렉토리)\n", path, count)
	return err
}

// Start는 모니터링을 시작합니다.
func (m *Monitor) Start() error {
	if m.running {
		return fmt.Errorf("모니터링이 이미 실행 중입니다")
	}

	if len(m.devices) == 0 {
		return fmt.Errorf("모니터링할 장치가 없습니다")
	}

	// fsnotify 워처 초기화
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("파일 시스템 감시자 생성 실패: %v", err)
	}
	m.watcher = watcher

	// 각 장치에 대해 재귀적 감시 설정
	for _, device := range m.devices {
		go func(dev string) {
			log.Printf("%s 장치 모니터링 시작...\n", dev)
			err := m.watchRecursive(dev)
			if err != nil {
				log.Printf("장치 %s 감시 설정 중 오류 발생: %v\n", dev, err)
			}
		}(device)
	}

	// 이벤트 처리 고루틴
	go m.processEvents()

	// 데이터베이스 초기화
	db, err := NewDatabase(m.dbPath)
	if err != nil {
		if m.watcher != nil {
			m.watcher.Close()
		}
		return fmt.Errorf("데이터베이스 초기화 실패: %v", err)
	}
	m.db = db

	m.running = true
	log.Println("파일 모니터링 시작됨")

	// 타이머를 사용하여 주기적으로 데이터베이스에 저장
	m.saveTimer = time.NewTicker(m.interval)
	go m.periodicSave()

	return nil
}

// periodicSave는 주기적으로 수집된 이벤트를 데이터베이스에 저장합니다.
func (m *Monitor) periodicSave() {
	for {
		select {
		case <-m.saveTimer.C:
			if !m.running {
				return
			}

			// 새로운 이벤트 저장
			m.saveEventsToDatabase()

			log.Printf("데이터베이스에 저장 완료 (interval: %s)\n", m.interval)
		}
	}
}

// saveEventsToDatabase는 수집된 이벤트를 데이터베이스에 저장합니다.
func (m *Monitor) saveEventsToDatabase() {
	if m.db == nil {
		return
	}

	m.eventsMutex.Lock()
	events := m.fileEvents
	m.fileEvents = []FileEvent{} // 저장 후 이벤트 목록 초기화
	m.eventsMutex.Unlock()

	if len(events) == 0 {
		return
	}

	// 일괄 저장
	err := m.db.SaveBatchFileEvents(events)
	if err != nil {
		log.Printf("이벤트 저장 중 오류 발생: %v\n", err)

		// 오류 발생 시 이벤트 복원
		m.eventsMutex.Lock()
		m.fileEvents = append(events, m.fileEvents...)
		m.eventsMutex.Unlock()
	} else {
		log.Printf("%d개의 이벤트가 데이터베이스에 저장되었습니다.\n", len(events))
	}
}

// isDirectory는 주어진 경로가 디렉토리인지 확인합니다.
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// processEvents는 파일 시스템 이벤트를 처리합니다.
func (m *Monitor) processEvents() {
	log.Println("파일 이벤트 처리 고루틴 시작")
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			// 이벤트 로깅 (디버깅)
			log.Printf("원시 이벤트 감지됨: %s, 작업: %s", event.Name, event.Op.String())

			// 파일 확장자 확인
			ext := strings.ToLower(filepath.Ext(event.Name))

			// 필터 로깅 (디버깅)
			log.Printf("파일: %s, 확장자: %s, 필터: %v", event.Name, ext, m.fileFilters)

			// 이벤트 필터링 (확장자 기준)
			matched := false
			for _, filter := range m.fileFilters {
				if ext == filter {
					matched = true
					break
				}
			}

			if !matched {
				log.Printf("필터와 일치하지 않아 무시됨: %s (확장자: %s)", event.Name, ext)
				continue
			}

			var operation string
			var createEvent, removeEvent bool

			// fsnotify 작업 처리
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				operation = "CREATE"
				createEvent = true
				log.Printf("[중요] 파일 생성 감지됨: %s", event.Name)

			case event.Op&fsnotify.Remove == fsnotify.Remove:
				operation = "REMOVE"
				removeEvent = true
				log.Printf("[중요] 파일 삭제 감지됨: %s", event.Name)

			case event.Op&fsnotify.Rename == fsnotify.Rename:
				// Rename은 일반적으로 REMOVE와 CREATE로 처리됨
				log.Printf("파일 이름 변경 감지됨: %s", event.Name)
				operation = "RENAME"

			case event.Op&fsnotify.Write == fsnotify.Write:
				log.Printf("파일 쓰기 감지됨: %s", event.Name)
				operation = "WRITE"

			case event.Op&fsnotify.Chmod == fsnotify.Chmod:
				log.Printf("파일 권한 변경 감지됨: %s", event.Name)
				operation = "CHMOD"

			default:
				log.Printf("알 수 없는 작업 감지됨: %s (%s)", event.Name, event.Op.String())
				continue
			}

			// CREATE 또는 REMOVE 이벤트만 처리
			if createEvent || removeEvent {
				log.Printf("파일 %s: %s (타입: %s)\n", operation, event.Name, ext)

				// 이벤트 기록
				m.eventsMutex.Lock()
				m.fileEvents = append(m.fileEvents, FileEvent{
					Path:      event.Name,
					Operation: operation,
					Timestamp: time.Now(),
					FileType:  ext,
				})
				m.eventsMutex.Unlock()

				// 새 디렉터리가 생성된 경우 감시 대상에 추가
				if createEvent && isDirectory(event.Name) {
					log.Printf("새 디렉터리 감지됨, 감시 대상에 추가: %s", event.Name)
					m.watchMutex.Lock()
					err := m.watchRecursive(event.Name)
					if err != nil {
						log.Printf("새 디렉터리 감시 설정 실패: %v", err)
					}
					m.watchMutex.Unlock()
				}
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("감시자 오류: %v", err)
		}
	}
}

// Stop은 모니터링을 중지합니다.
func (m *Monitor) Stop() {
	if !m.running {
		log.Println("모니터링이 실행 중이지 않습니다")
		return
	}

	m.running = false

	// 저장 타이머 종료
	if m.saveTimer != nil {
		m.saveTimer.Stop()
	}

	// 마지막으로 데이터베이스에 저장
	m.saveEventsToDatabase()

	// 리소스 정리
	if m.watcher != nil {
		m.watcher.Close()
	}

	// 데이터베이스 연결 종료
	if m.db != nil {
		m.db.Close()
	}

	log.Println("파일 모니터링 중지됨")
}

// GetFileEvents는 현재까지 수집된 파일 이벤트를 반환합니다.
func (m *Monitor) GetFileEvents() []FileEvent {
	return m.fileEvents
}

// GetAllFileEvents는 데이터베이스에 저장된 모든 파일 이벤트를 반환합니다.
func (m *Monitor) GetAllFileEvents() ([]FileEvent, error) {
	if m.db == nil {
		return nil, fmt.Errorf("데이터베이스가 초기화되지 않았습니다")
	}
	return m.db.GetFileEvents()
}

// PrintStats는 수집된 파일 이벤트를 출력합니다.
func (m *Monitor) PrintStats() {
	// 메모리에 있는 이벤트 출력
	if len(m.fileEvents) == 0 {
		fmt.Println("메모리에 수집된 파일 이벤트가 없습니다")
	} else {
		fmt.Println("\n===== 메모리 내 파일 이벤트 =====")
		for _, event := range m.fileEvents {
			fmt.Printf("[%s] %s\n", event.Timestamp.Format("2006-01-02 15:04:05"), event.Path)
			fmt.Printf("  작업: %s, 파일 유형: %s\n", event.Operation, event.FileType)
			fmt.Println("----------------------------")
		}
	}

	// 데이터베이스에서 이벤트 불러와 출력
	if m.db != nil {
		events, err := m.db.GetFileEvents()
		if err != nil {
			fmt.Printf("데이터베이스 조회 오류: %v\n", err)
			return
		}

		if len(events) == 0 {
			fmt.Println("\n데이터베이스에 저장된 파일 이벤트가 없습니다")
		} else {
			fmt.Printf("\n===== 데이터베이스 저장 파일 이벤트 (%d개) =====\n", len(events))
			// 일정 개수만 표시 (너무 많으면 화면이 복잡해짐)
			maxDisplay := 10
			displayCount := len(events)
			if displayCount > maxDisplay {
				displayCount = maxDisplay
				fmt.Printf("(최근 %d개만 표시)\n", maxDisplay)
			}

			for i := 0; i < displayCount; i++ {
				event := events[i]
				fmt.Printf("[%s] %s\n", event.Timestamp.Format("2006-01-02 15:04:05"), event.Path)
				fmt.Printf("  작업: %s, 파일 유형: %s\n", event.Operation, event.FileType)
				fmt.Println("----------------------------")
			}
		}
	}
}

// GetDevices는 현재 모니터링 중인 장치 목록을 반환합니다.
func (m *Monitor) GetDevices() []string {
	return m.devices
}

// GetFileFilters는 현재 설정된 파일 확장자 필터 목록을 반환합니다.
func (m *Monitor) GetFileFilters() []string {
	return m.fileFilters
}
