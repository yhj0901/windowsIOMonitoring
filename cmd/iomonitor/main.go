package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/yhj0901/windowsIOMonitoring/pkg/monitor"
)

func main() {
	// 디버그 모드 확인
	debugMode := os.Getenv("DEBUG_MONITOR") == "true"
	if debugMode {
		log.Println("디버그 모드 활성화됨")
	}

	// 명령줄 인자 파싱
	intervalFlag := flag.Duration("interval", 5*time.Second, "모니터링 간격 (예: 5s, 1m)")
	deviceFlag := flag.String("device", "", "모니터링할 장치 (쉼표로 구분)")
	filtersFlag := flag.String("filters", ".exe,.dll", "모니터링할 파일 확장자 (쉼표로 구분)")
	dbPathFlag := flag.String("db", "monitor.db", "데이터베이스 파일 경로")
	testFlag := flag.Bool("test", false, "테스트 모드 (더미 파일 생성)")
	versionFlag := flag.Bool("version", false, "버전 정보 출력")
	flag.Parse()

	// 버전 정보 출력
	if *versionFlag {
		fmt.Println("Windows 파일 모니터링 도구 v0.1.2")

		os.Exit(0)
	}

	// 로그 설정
	log.SetPrefix("[파일 모니터] ")
	log.SetFlags(log.Ldate | log.Ltime)

	// 모니터 인스턴스 생성
	mon := monitor.NewMonitor(*intervalFlag)

	// 데이터베이스 경로 설정
	mon.SetDatabasePath(*dbPathFlag)

	// 장치 추가
	if *deviceFlag != "" {
		devices := strings.Split(*deviceFlag, ",")
		for _, device := range devices {
			mon.AddDevice(strings.TrimSpace(device))
		}
	} else {
		// 기본 장치 추가 (Windows 기준)
		mon.AddDevice("C:\\")
	}

	// 파일 필터 설정
	if *filtersFlag != "" {
		filters := strings.Split(*filtersFlag, ",")
		// 확장자 형식 확인 및 수정
		for i, filter := range filters {
			filter = strings.TrimSpace(filter)
			if !strings.HasPrefix(filter, ".") {
				filters[i] = "." + filter
			} else {
				filters[i] = filter
			}
		}
		mon.SetFileFilters(filters)
	}

	// 테스트 모드
	if *testFlag || debugMode {
		go generateTestFiles()
	}

	// 모니터링 시작
	err := mon.Start()
	if err != nil {
		log.Fatalf("모니터링 시작 실패: %v", err)
	}

	fmt.Printf("파일 모니터링이 시작되었습니다. 종료하려면 Ctrl+C를 누르세요.\n")
	fmt.Printf("모니터링 대상: %s\n", strings.Join(mon.GetDevices(), ", "))
	fmt.Printf("파일 필터: %s\n", strings.Join(mon.GetFileFilters(), ", "))
	fmt.Printf("데이터베이스: %s\n", *dbPathFlag)
	fmt.Printf("저장 간격: %s\n", *intervalFlag)

	// 종료 시그널 처리
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 모니터링 중지
	mon.Stop()

	// 이벤트 출력
	mon.PrintStats()

	fmt.Println("프로그램이 종료되었습니다.")
}

// generateTestFiles는 테스트용 더미 파일을 생성합니다.
func generateTestFiles() {
	log.Println("테스트 파일 생성 모드 시작")
	testDir := "test_files"

	// 테스트 디렉토리 생성
	if err := os.MkdirAll(testDir, 0755); err != nil {
		log.Printf("테스트 디렉토리 생성 실패: %v", err)
		return
	}

	// 파일 생성 및 삭제 반복
	for i := 0; i < 5; i++ {
		// 1초 대기
		time.Sleep(2 * time.Second)

		// 테스트 파일 생성
		exeFile := filepath.Join(testDir, fmt.Sprintf("test_%d.exe", i))
		dllFile := filepath.Join(testDir, fmt.Sprintf("test_%d.dll", i))

		// EXE 파일 생성
		if err := os.WriteFile(exeFile, []byte("test exe file"), 0644); err != nil {
			log.Printf("테스트 EXE 파일 생성 실패: %v", err)
		} else {
			log.Printf("테스트 파일 생성됨: %s", exeFile)
		}

		// DLL 파일 생성
		if err := os.WriteFile(dllFile, []byte("test dll file"), 0644); err != nil {
			log.Printf("테스트 DLL 파일 생성 실패: %v", err)
		} else {
			log.Printf("테스트 파일 생성됨: %s", dllFile)
		}

		// 3초 대기
		time.Sleep(3 * time.Second)

		// 파일 삭제
		if err := os.Remove(exeFile); err != nil {
			log.Printf("테스트 EXE 파일 삭제 실패: %v", err)
		} else {
			log.Printf("테스트 파일 삭제됨: %s", exeFile)
		}

		if err := os.Remove(dllFile); err != nil {
			log.Printf("테스트 DLL 파일 삭제 실패: %v", err)
		} else {
			log.Printf("테스트 파일 삭제됨: %s", dllFile)
		}
	}

	log.Println("테스트 파일 생성 완료")
}
