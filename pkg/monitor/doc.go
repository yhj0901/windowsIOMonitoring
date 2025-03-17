// Package monitor는 Windows 시스템의 파일 및 IO 활동을 모니터링하는 기능을 제공합니다.
//
// 이 패키지는 파일 시스템 변경 감지, 특정 파일 확장자 필터링, 다중 드라이브 모니터링,
// 통계 수집 및 보고 기능을 제공합니다.
//
// 기본 사용 예시:
//
//	mon := monitor.NewMonitor(5 * time.Second)
//	mon.AddDevice("C:\\")
//	mon.SetFileFilters([]string{".exe", ".dll"})
//	err := mon.Start()
//	// ... 종료 신호 대기 ...
//	mon.Stop()
//	mon.PrintStats()
package monitor
