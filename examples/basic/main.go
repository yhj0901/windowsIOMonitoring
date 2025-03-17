package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yangheejun/windowsIOMonitoring/pkg/monitor"
)

func main() {
	// 로그 설정
	log.SetPrefix("[예제] ")
	log.SetFlags(log.Ldate | log.Ltime)

	// 모니터 인스턴스 생성
	mon := monitor.NewMonitor(5 * time.Second)

	// 장치 추가 (Windows 기준)
	mon.AddDevice("C:\\")

	// 파일 필터 설정
	mon.SetFileFilters([]string{".exe", ".dll"})

	// 모니터링 시작
	err := mon.Start()
	if err != nil {
		log.Fatalf("모니터링 시작 실패: %v", err)
	}

	fmt.Println("파일 모니터링이 시작되었습니다. 종료하려면 Ctrl+C를 누르세요.")
	fmt.Println("모니터링 중인 장치: C:\\")
	fmt.Println("모니터링 중인 파일 타입: .exe, .dll")

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
