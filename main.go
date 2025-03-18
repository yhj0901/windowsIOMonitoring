package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	log.Println("Windows IO 모니터링 시작")

	// 프로그램 버전 정보
	fmt.Println("Windows 파일 모니터링 도구 v0.1.3")
	fmt.Println("Copyright © 2025 yangheejune")
	fmt.Println()

	// 디버깅 모드 설정 (환경 변수)
	os.Setenv("DEBUG_MONITOR", "true")

	// 현재 디렉토리 확인
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("작업 디렉토리를 가져오는 중 오류 발생: %v", err)
	}

	// cmd/iomonitor/main.go 실행
	cmdPath := filepath.Join(dir, "cmd", "iomonitor", "main.go")

	// go run 명령 실행
	cmd := exec.Command("go", "run", cmdPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("cmd/iomonitor/main.go 실행 중...")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("cmd/iomonitor/main.go 실행 중 오류 발생: %v", err)
	}
}
