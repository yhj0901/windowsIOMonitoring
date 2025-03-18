package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/yhj0901/windowsIOMonitoring/pkg/monitor"
)

func main() {
	// 명령줄 인자 파싱
	intervalFlag := flag.Duration("interval", 5*time.Second, "모니터링 간격 (예: 5s, 1m)")
	deviceFlag := flag.String("device", "", "모니터링할 장치 (쉼표로 구분)")
	filtersFlag := flag.String("filters", ".exe,.dll", "모니터링할 파일 확장자 (쉼표로 구분)")
	versionFlag := flag.Bool("version", false, "버전 정보 출력")
	flag.Parse()

	// 버전 정보 출력
	if *versionFlag {
		fmt.Println("Windows IO 모니터링 도구 v0.1.0")
		os.Exit(0)
	}

	// 로그 설정
	log.SetPrefix("[IO 모니터] ")
	log.SetFlags(log.Ldate | log.Ltime)

	// 모니터 인스턴스 생성
	mon := monitor.NewMonitor(*intervalFlag)

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

	// 모니터링 시작
	err := mon.Start()
	if err != nil {
		log.Fatalf("모니터링 시작 실패: %v", err)
	}

	fmt.Printf("파일 모니터링이 시작되었습니다. 종료하려면 Ctrl+C를 누르세요.\n")
	fmt.Printf("모니터링 간격: %s\n", *intervalFlag)

	// 종료 시그널 처리
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 모니터링 중지
	mon.Stop()

	// 통계 출력
	mon.PrintStats()

	fmt.Println("프로그램이 종료되었습니다.")
}
