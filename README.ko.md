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
- Windows 시작 프로그램 체크 토글
- GitHub Releases 기반 수동 업데이트 확인 및 SHA-256 검증 후 자동 교체/재시작
- 클립보드 이미지 미리보기 후 사용자 승인 시 OpenRouter Vision으로 AnyDesk 번호 분석
- OpenRouter API Key는 Windows Credential Manager에 저장
- 설정 초기화 / 백업 / 복원
- 상황별 OpenRouter/네트워크/분석 오류 안내
- 한국어/영어 UI
- 트레이 메뉴에서 재시작 없이 언어 변경
- 기본 언어 한국어
- 실행 파일 및 트레이 아이콘 내장
- EXE 단독 배포 가능, 사용자 설정은 `%LOCALAPPDATA%` 아래에 저장

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
- 이미지 자동 분석
- 시작 프로그램
- 기본 암호 수정
- OpenRouter 설정
- 업데이트 확인
- 설정 백업
- 설정 복원
- 설정 초기화
- 언어
  - 한국어
  - English
- 종료

## 설정 파일

`config.ini`는 실행 파일 옆이 아니라 현재 Windows 사용자의 다음 경로에 저장됩니다.

```text
%LOCALAPPDATA%\QuickAnydeskConnect\config.ini
```

이전 버전에서 EXE 옆에 사용하던 유효한 `config.ini`가 있고 새 사용자 경로에 설정이 아직 없으면 최초 실행 시 한 번 자동으로 마이그레이션합니다.

```ini
[anydesk]
password=YOUR_PASSWORD

[general]
language=ko
image_analysis=false

[openrouter]
model=google/gemma-4-26b-a4b-it:free
```

지원 언어:

- `ko`
- `en`

`[general]` 또는 `language`가 없으면 한국어를 사용합니다.

> [!WARNING]
> 무인 접속 기본 암호는 `config.ini`에 평문으로 저장됩니다. 파일은 현재 Windows 사용자의 Local AppData 영역에 저장됩니다.

### 설정 백업 / 복원 / 초기화

**설정 백업**은 현재 AnyDesk 기본 암호, 언어, 이미지 분석 설정, OpenRouter 모델 및 Windows Credential Manager에 저장된 OpenRouter API Key를 `.qacbackup` 파일로 저장합니다. **시작 프로그램** 등록 상태는 백업/복원/초기화 대상에 포함하지 않습니다.

**설정 복원**은 백업 파일을 먼저 검증한 뒤 현재 설정과 OpenRouter API Key를 교체합니다. **설정 초기화**는 앱 설정과 OpenRouter API Key를 초기화하고 새 기본 AnyDesk 암호를 다시 입력하도록 합니다.

> [!IMPORTANT]
> `.qacbackup` 파일에는 AnyDesk 기본 암호와 OpenRouter API Key가 읽을 수 있는 형태로 포함됩니다. 다른 사람에게 공유하지 말고 안전한 위치에 보관하십시오.

## 이미지 자동 분석

이미지 자동 분석은 기본적으로 꺼져 있습니다. 트레이 메뉴에서 활성화할 수 있습니다.

최초 활성화 시 OpenRouter API Key를 입력하면 비용이 발생하지 않는 `GET /api/v1/key` 요청으로 키를 검증한 뒤 Windows Credential Manager에 저장합니다. API Key는 `config.ini`에 저장하지 않습니다.

이미지가 클립보드에 복사되면 곧바로 외부로 전송하지 않습니다. 먼저 이미지 미리보기와 **분석 / 무시** 버튼을 표시하며, 사용자가 **분석**을 선택한 경우에만 이미지가 OpenRouter로 전송됩니다. 미리보기는 화면에 맞게 고품질 보간으로 축소하며, 분석용 이미지는 원본이 큰 경우에만 비율을 유지한 채 긴 변 최대 1600px로 축소한 뒤 전송합니다. 작은 이미지는 확대하지 않습니다.

**OpenRouter 설정**에서는 API Key와 모델 ID를 모두 직접 입력할 수 있습니다. 같은 창의 버튼으로 OpenRouter API Key 발급 페이지도 열 수 있습니다. 기본 모델은 `google/gemma-4-26b-a4b-it:free`이지만 모델 입력란은 정해진 목록으로 제한하지 않습니다. 이미지 입력을 지원하는 OpenRouter 모델이면 원하는 모델 ID를 직접 입력할 수 있습니다.

### 추천 이미지 분석 모델

아래 목록은 스크린샷에서 짧은 AnyDesk 번호를 읽는 용도로 검토하기 좋은 예시입니다. OpenRouter의 모델 가용성, 가격, 요청 제한, 모델 ID는 변경될 수 있으므로 실제 사용 전 현재 모델 페이지를 확인하십시오.

| 모델 ID | 비용 | 참고 |
| --- | --- | --- |
| `google/gemma-4-26b-a4b-it:free` | 무료 | 기본값. 멀티모달 지원. 무료 upstream provider가 일시적으로 rate limit에 걸릴 수 있습니다. |
| `google/gemma-4-31b-it:free` | 무료 | 더 큰 무료 멀티모달 대안. 기본 무료 모델이 제한 상태일 때 바꿔볼 수 있습니다. |
| `nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free` | 무료 | 이미지 입력을 지원하는 범용 멀티모달 모델입니다. |
| `google/gemini-2.5-flash-lite` | 유료, 저비용 | 단순 번호 추출 용도에서 안정적인 대안으로 사용할 수 있습니다. |

무료 모델에서 upstream provider 요청 제한 오류가 발생하면 잠시 후 다시 시도하거나 **OpenRouter 설정**에서 다른 모델로 변경하십시오. 프로그램은 이 상황을 일반적인 OpenRouter/계정 요청 한도와 구분하고, 가능한 경우 provider의 실제 오류 응답도 함께 표시합니다.

> [!IMPORTANT]
> 사용자가 분석을 승인한 클립보드 이미지는 OpenRouter와 선택한 모델 제공자에게 전송됩니다. 무료 모델 제공자는 자체적인 로깅, 보관, 가용성 및 요청 제한 정책을 적용할 수 있습니다. 민감한 이미지라면 외부 서비스 전송에 동의할 수 있는 경우에만 분석하십시오.

OpenRouter는 별도 서비스이며 사용량에 따라 API 비용이 발생할 수 있습니다. API Key는 <https://openrouter.ai/settings/keys>에서 생성하거나 관리할 수 있습니다.

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
- `v*` 형식 태그를 만들면 `QuickAnydeskConnect.exe`와 `QuickAnydeskConnect.exe.sha256`을 GitHub Release에 첨부합니다.
- 일반 빌드 artifact는 업로드하지 않아 Actions 저장공간 사용을 최소화합니다.

## 라이선스

MIT License. [LICENSE](LICENSE)를 참고하십시오.

## 면책

Quick Anydesk Connect는 독립적으로 제작된 유틸리티이며 AnyDesk Software GmbH와 제휴, 승인 또는 후원 관계가 없습니다. AnyDesk는 해당 권리자의 상표입니다.
