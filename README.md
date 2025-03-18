# Windows IO 모니터링

Windows 시스템의 입출력(IO) 활동을 모니터링하는 Go 애플리케이션입니다.

## 기능

- 디스크 읽기/쓰기 작업 모니터링
- 장치별 IO 통계 수집
- 실시간 모니터링 및 보고서 생성
- .exe, .dll 파일 생성 및 수정 감지
- 디렉토리 변경 감지 및 자동 모니터링
- SQLite 데이터베이스를 통한 이벤트 저장 및 조회

## 설치 방법

### 사전 요구사항

- Go 1.16 이상 (개발 시)
- Windows 운영체제 (Windows 10/11 권장)

### 설치

```bash
# 저장소 클론
git clone https://github.com/yhj0901/windowsIOMonitoring.git
cd windowsIOMonitoring

# 의존성 설치
go mod tidy

# 빌드
go build -o iomonitor.exe cmd/iomonitor/main.go
```

### Windows 사용자를 위한 빠른 설치

1. 최신 릴리스에서 `iomonitor.exe`, `run_monitor.bat`, `run_monitor_custom.bat` 파일을 다운로드합니다.
2. 세 파일을 동일한 디렉토리에 저장합니다.
3. `run_monitor.bat` 또는 `run_monitor_custom.bat`를 더블 클릭하여 실행합니다.

## 사용 방법

### 명령줄에서 실행

```bash
# 기본 설정으로 실행
./iomonitor.exe

# 모니터링 간격 설정
./iomonitor.exe -interval 10s

# 특정 장치 모니터링
./iomonitor.exe -device "C:\"

# 여러 장치 모니터링
./iomonitor.exe -device "C:\,D:\"

# 특정 파일 확장자 모니터링
./iomonitor.exe -filters ".exe,.dll,.sys"

# 데이터베이스 파일 경로 지정
./iomonitor.exe -db "C:\logs\monitor.db"

# 버전 정보 확인
./iomonitor.exe -version
```

### 배치 파일로 실행 (Windows)

- `run_monitor.bat`: 기본 설정으로 실행
- `run_monitor_custom.bat`: 사용자 정의 설정으로 실행 (10초 간격, C:\ 및 D:\ 드라이브, .exe/.dll/.sys 파일)

## 프로젝트 구조

```
windowsIOMonitoring/
├── cmd/
│   └── iomonitor/      # 실행 파일 소스 코드
├── pkg/
│   └── monitor/        # 모니터링 기능 패키지
├── main.go             # 메인 애플리케이션 진입점
├── iomonitor.exe       # Windows용 실행 파일
├── run_monitor.bat     # 기본 실행 배치 파일
├── run_monitor_custom.bat # 사용자 정의 실행 배치 파일
├── go.mod              # Go 모듈 정의
└── README.md           # 프로젝트 문서
```

## 라이센스

이 프로젝트는 MIT 라이센스 하에 배포됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 데이터베이스 구조

모니터링 데이터는 SQLite 데이터베이스에 저장됩니다:

### 파일 이벤트 테이블 (file_events)

| 필드      | 타입     | 설명                       |
|-----------|----------|----------------------------|
| id        | INTEGER  | 기본 키 (자동 증가)        |
| timestamp | DATETIME | 이벤트 발생 시간           |
| path      | TEXT     | 파일 경로                  |
| operation | TEXT     | 작업 유형 (CREATE/WRITE/REMOVE) |
| file_type | TEXT     | 파일 확장자                |

### IO 통계 테이블 (io_stats)

| 필드          | 타입     | 설명                    |
|---------------|----------|-------------------------|
| id            | INTEGER  | 기본 키 (자동 증가)     |
| timestamp     | DATETIME | 통계 수집 시간          |
| device        | TEXT     | 장치 경로               |
| read_bytes    | INTEGER  | 읽은 바이트 수          |
| written_bytes | INTEGER  | 쓴 바이트 수            |
| read_ops      | INTEGER  | 읽기 작업 수            |
| write_ops     | INTEGER  | 쓰기 작업 수            |

## 데이터 수집 및 저장

- 파일 이벤트(파일 생성, 수정, 삭제)는 실시간으로 감지되어 메모리에 저장됩니다.
- 설정된 간격(`interval`)마다 메모리에 있는 이벤트와 IO 통계가 데이터베이스에 저장됩니다.
- 프로그램 종료 시 저장되지 않은 모든 데이터가 데이터베이스에 저장됩니다.
