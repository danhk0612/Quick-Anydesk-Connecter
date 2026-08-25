package main

type messages struct {
	trayConnect       string
	trayStartupAdd    string
	trayStartupRemove string
	trayLanguage      string
	trayExit          string

	connectTitle       string
	connectField       string
	connectDescription string
	connectButton      string
	cancelButton       string
	saveButton         string

	setupTitle       string
	setupPassword    string
	setupDescription string

	confirmConnect string
	inputCheckTitle string
	invalidID      string
	emptyPassword  string

	startupAddSuccess    string
	startupAddFailure    string
	startupRemoveSuccess string
	startupRemoveFailure string

	errExecutablePath string
	errTrayClass      string
	errTrayWindow     string
	errClipboardWatch string
	errAnyDeskStart   string
	errAnyDeskRun     string
	errRegistryOpen   string
	errPathConvert    string
	errRegistrySet    string
	errRegistryDelete string
	errTrayIcon       string
	errConfigSave     string
	errConfigRead     string
	errPasswordEmpty  string
	errPasswordMissing string
	errAnyDeskMissing string
	errDialogCreate   string
}

var languageMessages = map[string]messages{
	"ko": {
		trayConnect: "원격 접속",
		trayStartupAdd: "시작 프로그램 등록",
		trayStartupRemove: "시작 프로그램 제거",
		trayLanguage: "언어",
		trayExit: "종료",

		connectTitle: "AnyDesk 원격 접속",
		connectField: "원격 번호",
		connectDescription: "접속할 AnyDesk 번호를 입력하세요.",
		connectButton: "접속",
		cancelButton: "취소",
		saveButton: "저장",

		setupTitle: "Quick Anydesk Connect 초기 설정",
		setupPassword: "기본 암호",
		setupDescription: "최초 실행입니다. 공통 무인 접속 암호를 입력하세요.",

		confirmConnect: "%s 으로 원격 접속 하시겠습니까?",
		inputCheckTitle: "입력 확인",
		invalidID: "AnyDesk 번호는 9자리 또는 10자리 숫자여야 합니다.",
		emptyPassword: "기본 암호를 입력하세요.",

		startupAddSuccess: "시작 프로그램 등록이 완료되었습니다.",
		startupAddFailure: "시작 프로그램 등록에 실패했습니다.\n\n%s",
		startupRemoveSuccess: "시작 프로그램 제거가 완료되었습니다.",
		startupRemoveFailure: "시작 프로그램 제거에 실패했습니다.\n\n%s",

		errExecutablePath: "실행 파일 경로를 확인할 수 없습니다.",
		errTrayClass: "트레이 창 클래스를 등록할 수 없습니다.",
		errTrayWindow: "트레이 창을 만들 수 없습니다.",
		errClipboardWatch: "클립보드 감시를 시작할 수 없습니다.",
		errAnyDeskStart: "AnyDesk 시작에 실패했습니다.\n\n%s",
		errAnyDeskRun: "AnyDesk 실행에 실패했습니다.\n\n%s",
		errRegistryOpen: "시작 프로그램 레지스트리를 열 수 없습니다. (오류 코드: %d)",
		errPathConvert: "실행 경로 변환 실패: %v",
		errRegistrySet: "시작 프로그램 값을 저장할 수 없습니다. (오류 코드: %d)",
		errRegistryDelete: "시작 프로그램 값을 제거할 수 없습니다. (오류 코드: %d)",
		errTrayIcon: "트레이 아이콘을 추가할 수 없습니다.",
		errConfigSave: "설정 파일을 저장할 수 없습니다.\n\n%s",
		errConfigRead: "config.ini 읽기 오류:\n\n%s",
		errPasswordEmpty: "config.ini의 password 값이 비어 있습니다.",
		errPasswordMissing: "config.ini에서 [anydesk]의 password 값을 찾을 수 없습니다.",
		errAnyDeskMissing: "AnyDesk 실행 파일을 찾을 수 없습니다.\n\n검색 위치:\nC:\\Program Files (x86)\\AnyDesk-*\\AnyDesk-*.exe\nC:\\Program Files\\AnyDesk-*\\AnyDesk-*.exe",
		errDialogCreate: "입력창을 만들 수 없습니다.",
	},
	"en": {
		trayConnect: "Remote Connection",
		trayStartupAdd: "Add to Startup",
		trayStartupRemove: "Remove from Startup",
		trayLanguage: "Language",
		trayExit: "Exit",

		connectTitle: "AnyDesk Remote Connection",
		connectField: "Remote Address",
		connectDescription: "Enter the AnyDesk address to connect to.",
		connectButton: "Connect",
		cancelButton: "Cancel",
		saveButton: "Save",

		setupTitle: "Quick Anydesk Connect Initial Setup",
		setupPassword: "Default Password",
		setupDescription: "Enter the shared unattended access password.",

		confirmConnect: "Connect to %s?",
		inputCheckTitle: "Check Input",
		invalidID: "The AnyDesk address must be a 9- or 10-digit number.",
		emptyPassword: "Enter the default password.",

		startupAddSuccess: "Added Quick Anydesk Connect to startup.",
		startupAddFailure: "Failed to add Quick Anydesk Connect to startup.\n\n%s",
		startupRemoveSuccess: "Removed Quick Anydesk Connect from startup.",
		startupRemoveFailure: "Failed to remove Quick Anydesk Connect from startup.\n\n%s",

		errExecutablePath: "Could not determine the executable path.",
		errTrayClass: "Could not register the tray window class.",
		errTrayWindow: "Could not create the tray window.",
		errClipboardWatch: "Could not start clipboard monitoring.",
		errAnyDeskStart: "Failed to start AnyDesk.\n\n%s",
		errAnyDeskRun: "Failed to launch AnyDesk.\n\n%s",
		errRegistryOpen: "Could not open the startup registry key. (Error code: %d)",
		errPathConvert: "Failed to convert the executable path: %v",
		errRegistrySet: "Could not save the startup registry value. (Error code: %d)",
		errRegistryDelete: "Could not remove the startup registry value. (Error code: %d)",
		errTrayIcon: "Could not add the tray icon.",
		errConfigSave: "Could not save the configuration file.\n\n%s",
		errConfigRead: "Could not read config.ini.\n\n%s",
		errPasswordEmpty: "The password value in config.ini is empty.",
		errPasswordMissing: "Could not find [anydesk] password in config.ini.",
		errAnyDeskMissing: "Could not find the AnyDesk executable.\n\nSearched:\nC:\\Program Files (x86)\\AnyDesk-*\\AnyDesk-*.exe\nC:\\Program Files\\AnyDesk-*\\AnyDesk-*.exe",
		errDialogCreate: "Could not create the input window.",
	},
}

func currentMessages() messages {
	if m, ok := languageMessages[language]; ok {
		return m
	}
	return languageMessages["ko"]
}
