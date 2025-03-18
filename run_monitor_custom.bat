@echo off
echo Windows IO 모니터링 도구 v0.1.0
echo Copyright © 2023
echo.

REM 사용자 정의 설정으로 실행
echo 사용자 정의 설정으로 실행 중...
echo 모니터링 간격: 10초
echo 모니터링 대상: C:\ 및 D:\
echo 파일 필터: .exe, .dll, .sys
echo 종료하려면 Ctrl+C를 누르세요.
echo.

REM 실행 파일 실행 (사용자 정의 옵션)
iomonitor.exe -interval 10s -device "C:\,D:\" -filters ".exe,.dll,.sys" -db "monitor.db"

pause 