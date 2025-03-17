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

// IOStats는 IO 통계 정보를 저장하는 구조체입니다.
type IOStats struct {
	Timestamp    time.Time
	ReadBytes    uint64
	WrittenBytes uint64
	ReadOps      uint64
	WriteOps     uint64
	Device       string
}

// FileEvent는 파일 이벤트 정보를 저장하는 구조체입니다.
type FileEvent struct {
	Path      string
	Operation string
	Timestamp time.Time
	FileType  string
}

// Monitor는 IO 모니터링을 담당하는 구조체입니다.
type Monitor struct {
	interval    time.Duration
	devices     []string
	stats       []IOStats
	fileEvents  []FileEvent
	running     bool
	watcher     *fsnotify.Watcher
	watchMutex  sync.Mutex
	fileFilters []string
}

// NewMonitor는 새로운 IO 모니터 인스턴스를 생성합니다.
func NewMonitor(interval time.Duration) *Monitor {
	return &Monitor{
		interval:    interval,
		devices:     []string{},
		stats:       []IOStats{},
		fileEvents:  []FileEvent{},
		running:     false,
		fileFilters: []string{".exe", ".dll"}, // 기본 필터
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

// watchRecursive는 디렉터리를 재귀적으로 watcher에 등록하는 함수입니다.
func (m *Monitor) watchRecursive(path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			// 접근 권한이 없는 폴더는 스킵
			return nil
		}

		if info.IsDir() {
			m.watchMutex.Lock()
			err = m.watcher.Add(walkPath)
			m.watchMutex.Unlock()
			if err == nil {
				log.Printf("감시 시작: %s\n", walkPath)
			}
		}
		return nil
	})
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

	m.running = true
	log.Println("IO 모니터링 시작됨")

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

	return nil
}

// processEvents는 파일 시스템 이벤트를 처리하는 함수입니다.
func (m *Monitor) processEvents() {
	for {
		if !m.running {
			return
		}

		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			// 새로운 폴더가 생성될 때 추가적으로 감시 대상에 포함
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					m.watchMutex.Lock()
					err = m.watcher.Add(event.Name)
					m.watchMutex.Unlock()
					if err == nil {
						log.Printf("새로운 폴더 감시 추가: %s\n", event.Name)
					}
				}
			}

			// 필터에 맞는 파일 이벤트만 처리
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove) != 0 {
				ext := strings.ToLower(filepath.Ext(event.Name))
				for _, filter := range m.fileFilters {
					if ext == filter {
						log.Printf("모니터링 대상 파일 감지: %s (이벤트: %s)\n", event.Name, event.Op)

						// 이벤트 기록
						m.fileEvents = append(m.fileEvents, FileEvent{
							Path:      event.Name,
							Operation: event.Op.String(),
							Timestamp: time.Now(),
							FileType:  ext,
						})

						break
					}
				}
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Println("감시 오류:", err)
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
	if m.watcher != nil {
		m.watcher.Close()
	}
	log.Println("IO 모니터링 중지됨")
}

// GetStats는 현재까지 수집된 통계 정보를 반환합니다.
func (m *Monitor) GetStats() []IOStats {
	return m.stats
}

// GetFileEvents는 현재까지 수집된 파일 이벤트를 반환합니다.
func (m *Monitor) GetFileEvents() []FileEvent {
	return m.fileEvents
}

// PrintStats는 수집된 통계 정보를 출력합니다.
func (m *Monitor) PrintStats() {
	if len(m.stats) == 0 {
		fmt.Println("수집된 IO 통계 정보가 없습니다")
	} else {
		fmt.Println("===== IO 모니터링 통계 =====")
		for _, stat := range m.stats {
			fmt.Printf("[%s] 장치: %s\n", stat.Timestamp.Format("2006-01-02 15:04:05"), stat.Device)
			fmt.Printf("  읽기: %d 바이트 (%d 작업)\n", stat.ReadBytes, stat.ReadOps)
			fmt.Printf("  쓰기: %d 바이트 (%d 작업)\n", stat.WrittenBytes, stat.WriteOps)
			fmt.Println("----------------------------")
		}
	}

	// 파일 이벤트 출력
	if len(m.fileEvents) == 0 {
		fmt.Println("수집된 파일 이벤트가 없습니다")
	} else {
		fmt.Println("\n===== 파일 모니터링 이벤트 =====")
		for _, event := range m.fileEvents {
			fmt.Printf("[%s] %s\n", event.Timestamp.Format("2006-01-02 15:04:05"), event.Path)
			fmt.Printf("  작업: %s, 파일 유형: %s\n", event.Operation, event.FileType)
			fmt.Println("----------------------------")
		}
	}
}
