# Windows 파일 모니터링

Windows 시스템의 파일 생성 및 삭제 이벤트를 모니터링하는 Go 애플리케이션입니다.

## 기능

- .exe, .dll 파일 생성 및 삭제 이벤트 감지
- 지정된 드라이브와 디렉토리 재귀적 모니터링
- 새로 생성된 디렉토리 자동 모니터링
- SQLite 데이터베이스를 통한 이벤트 저장 및 조회
- 커스텀 파일 확장자 필터링
- 유연한 저장 간격 설정

## 설치 방법

### 사전 요구사항

- Go 1.16 이상 (개발 시)
- Windows 운영체제 (Windows 10/11 권장)
- SQLite 지원을 위한 CGO 활성화

### 설치

```bash
# 저장소 클론
git clone https://github.com/yhj0901/windowsIOMonitoring.git
cd windowsIOMonitoring

# 의존성 설치
go mod tidy

# 빌드 (Windows에서 직접 빌드 - 권장)
go build -o iomonitor.exe cmd/iomonitor/main.go

# 또는 CGO 활성화하여 빌드 (MinGW 필요)
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o iomonitor.exe cmd/iomonitor/main.go
```

### Windows 사용자를 위한 빠른 설치

1. 최신 릴리스에서 `iomonitor.exe`와 배치 파일들을 다운로드합니다.
2. 모든 파일을 동일한 디렉토리에 저장합니다.
3. 제공된 배치 파일 중 하나를 더블 클릭하여 실행합니다.

## 사용 방법

### 명령줄에서 실행

```bash
# 기본 설정으로 실행 (C:\ 드라이브, 5초 간격)
./iomonitor.exe

# 모니터링 간격 설정 (데이터베이스 저장 간격)
./iomonitor.exe -interval 10s

# 특정 장치 모니터링
./iomonitor.exe -device "C:\"

# 여러 장치 모니터링
./iomonitor.exe -device "C:\,D:\"

# 특정 파일 확장자 모니터링
./iomonitor.exe -filters ".exe,.dll,.sys"

# 데이터베이스 파일 경로 지정
./iomonitor.exe -db "C:\logs\monitor.db"

# 테스트 모드 실행 (더미 파일 생성)
./iomonitor.exe -test

# 버전 정보 확인
./iomonitor.exe -version
```

### 배치 파일로 실행 (Windows)

- `run_monitor_with_db.bat`: 타임스탬프가 포함된 DB 경로로 실행
- `run_monitor_custom_settings.bat`: 사용자가 장치, 필터, 간격을 설정할 수 있는 대화형 배치 파일
- `run_monitor_multi_drives.bat`: 여러 드라이브(C: 및 D:)를 동시에 모니터링

## 프로젝트 구조

```
windowsIOMonitoring/
├── cmd/
│   └── iomonitor/      # 실행 파일 소스 코드
├── pkg/
│   └── monitor/        # 모니터링 기능 패키지
├── main.go             # 디버그용 진입점
├── iomonitor.exe       # Windows용 실행 파일
├── run_monitor_custom_settings.bat # 사용자 정의 실행 배치 파일
├── go.mod              # Go 모듈 정의
└── README.md           # 프로젝트 문서
```

## 라이센스

이 프로젝트는 MIT 라이센스 하에 배포됩니다.

## 데이터베이스 구조

모니터링 데이터는 SQLite 데이터베이스에 저장됩니다:

### 파일 이벤트 테이블 (file_events)

| 필드      | 타입     | 설명                       |
|-----------|----------|----------------------------|
| id        | INTEGER  | 기본 키 (자동 증가)        |
| timestamp | DATETIME | 이벤트 발생 시간           |
| path      | TEXT     | 파일 경로                  |
| operation | TEXT     | 작업 유형 (CREATE/REMOVE)  |
| file_type | TEXT     | 파일 확장자                |

## 데이터 수집 및 저장

- 파일 이벤트(파일 생성, 삭제)는 실시간으로 감지되어 메모리에 저장됩니다.
- 설정된 간격(`interval`)마다 메모리에 있는 이벤트가 데이터베이스에 저장됩니다.
- 프로그램 종료 시 저장되지 않은 모든 데이터가 데이터베이스에 저장됩니다.

## 데이터베이스 확인 방법

저장된 데이터베이스 파일(.db)을 확인하려면 다음 도구 중 하나를 사용할 수 있습니다:

- **DB Browser for SQLite**: 무료 오픈소스 도구 (https://sqlitebrowser.org/dl/)
- **SQLite Studio**: 간편한 GUI 도구 (https://sqlitestudio.pl/)
- **Visual Studio Code**: SQLite 확장 프로그램 설치 후 사용
- **SQLite 명령줄 도구**: 명령줄에서 직접 쿼리 실행

## 디버깅 모드

디버깅을 위해 테스트 파일을 자동으로 생성하는 모드를 제공합니다:

```bash
# 환경 변수 설정으로 디버그 모드 활성화
# Windows
set DEBUG_MONITOR=true
./iomonitor.exe

# 또는 테스트 플래그 사용
./iomonitor.exe -test
```
