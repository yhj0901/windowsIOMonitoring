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
	fmt.Println("Windows IO 모니터링 도구 v0.1.0")
	fmt.Println("Copyright © 2023")
	fmt.Println()

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
