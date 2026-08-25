# Quick Anydesk Connect

[English](README.md)

Quick Anydesk Connect는 반복적인 AnyDesk 원격 지원 접속 절차를 간소화하기 위한 가벼운 Windows 트레이 유틸리티입니다.

클립보드에 새로 복사된 9자리 또는 10자리 AnyDesk 원격 번호를 감지하고 접속 여부를 확인한 뒤, 설정된 무인 접속 기본 암호를 자동으로 제출합니다. 기본 암호가 원격 PC와 다르면 기존 AnyDesk 암호 입력창에서 실제 암호를 입력할 수 있습니다.

## 주요 기능

- Windows 알림 영역(트레이)에 상주
- 새로 복사된 9자리/10자리 AnyDesk 번호 자동 감지
- 다른 창 위에 표시되는 접속 확인창
- 트레이 메뉴 또는 트레이 아이콘 더블클릭으로 직접 번호 입력
- 설정된 무인 접속 기본 암호 자동 제출
- AnyDesk가 완전히 종료된 경우 먼저 AnyDesk를 시작한 뒤 초기화 후 접속
- Windows 시작 프로그램 등록/제거
- 한국어/영어 UI
- 트레이 메뉴에서 재시작 없이 언어 변경
- 기본 언어 한국어
- 실행 파일 및 트레이 아이콘 내장
- EXE 단독 배포 가능, `config.ini`가 없으면 최초 실행 시 생성

## 요구사항

- Windows
- 지원 경로 중 하나에 AnyDesk 설치
- 소스 빌드 시 Go 1.22 이상

현재 다음 경로를 자동 검색합니다.

```text
C:\Program Files (x86)\AnyDesk-*\AnyDesk-*.exe
C:\Program Files\AnyDesk-*\AnyDesk-*.exe
C:\Program Files (x86)\AnyDesk\AnyDesk.exe
C:\Program Files\AnyDesk\AnyDesk.exe
```

## 사용 방법

1. `QuickAnydeskConnect.exe`를 실행합니다.
2. 최초 실행 시 공통으로 사용할 AnyDesk 무인 접속 암호를 입력합니다.
3. 메신저 등에서 고객의 AnyDesk 원격 번호를 복사합니다.
4. 접속 확인창이 뜨면 **예**를 선택합니다.

직접 번호를 입력하려면:

- 트레이 아이콘을 더블클릭하거나
- 트레이 아이콘 우클릭 → **원격 접속**을 선택합니다.

### 트레이 메뉴

- 원격 접속
- 시작 프로그램 등록
- 시작 프로그램 제거
- 언어
  - 한국어
  - English
- 종료

## 설정 파일

`config.ini`는 실행 파일과 같은 폴더에 저장됩니다.

```ini
[anydesk]
password=YOUR_PASSWORD

[general]
language=ko
```

지원 언어:

- `ko`
- `en`

`[general]` 또는 `language`가 없으면 한국어를 사용합니다.

> [!WARNING]
> 무인 접속 기본 암호는 `config.ini`에 평문으로 저장됩니다. 적절한 사용자만 접근할 수 있는 위치에 보관하십시오.

## AnyDesk 번호 판별

클립보드 텍스트에서 공백, 탭, 줄바꿈, 하이픈을 제거한 뒤 최종 값이 정확히 9자리 또는 10자리 숫자인 경우에만 AnyDesk 번호로 인식합니다.

허용 예:

```text
123456789
123 456 789
123-456-789
```

## 빌드

Windows에서 `build.bat`을 실행합니다.

빌드 과정:

1. 이전에 생성된 Windows 리소스 파일 삭제
2. [`go-winres`](https://github.com/tc-hib/go-winres)로 Windows EXE 아이콘 리소스 생성
3. GUI 실행 파일 `QuickAnydeskConnect.exe` 빌드

`build.bat`은 의도적으로 **EUC-KR + CRLF** 형식으로 저장합니다.

트레이 아이콘은 Go `embed`로도 실행 파일 내부에 포함되므로 빌드 후 `app.ico`는 배포할 필요가 없습니다.

## GitHub Actions

- 일반 push / pull request에서는 Windows 빌드 검증만 수행합니다.
- `v*` 형식 태그를 만들면 `QuickAnydeskConnect.exe`를 빌드해 GitHub Release에 첨부합니다.
- 일반 빌드 artifact는 업로드하지 않아 Actions 저장공간 사용을 최소화합니다.

## 라이선스

MIT License. [LICENSE](LICENSE)를 참고하십시오.

## 면책

Quick Anydesk Connect는 독립적으로 제작된 유틸리티이며 AnyDesk Software GmbH와 제휴, 승인 또는 후원 관계가 없습니다. AnyDesk는 해당 권리자의 상표입니다.
