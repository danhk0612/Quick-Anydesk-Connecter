package main

type messages struct {
	trayConnect        string
	trayImageAnalysis  string
	trayChangePassword string
	trayOpenRouter     string
	trayStartupAdd     string
	trayStartupRemove  string
	trayLanguage       string
	trayExit           string

	connectTitle       string
	connectField       string
	connectDescription string
	connectButton      string
	cancelButton       string
	saveButton         string
	verifyButton       string

	changePasswordTitle       string
	changePasswordDescription string
	openRouterTitle           string
	openRouterKeyLabel        string
	openRouterKeyDescription  string
	openRouterModelLabel      string
	openRouterKeysLink        string
	errOpenRouterModelEmpty   string
	openRouterSaved           string
	imagePreviewTitle         string
	imagePreviewQuestion      string
	analyzeButton             string
	ignoreButton              string
	imagePrivacyNotice        string
	anyDeskNotFoundInImage    string
	multipleAnyDeskIDs        string

	setupTitle       string
	setupPassword    string
	setupDescription string

	confirmConnect  string
	inputCheckTitle string
	invalidID       string
	emptyPassword   string

	startupAddSuccess    string
	startupAddFailure    string
	startupRemoveSuccess string
	startupRemoveFailure string

	errExecutablePath              string
	errTrayClass                   string
	errTrayWindow                  string
	errClipboardWatch              string
	errAnyDeskStart                string
	errAnyDeskRun                  string
	errRegistryOpen                string
	errPathConvert                 string
	errRegistrySet                 string
	errRegistryDelete              string
	errTrayIcon                    string
	errConfigSave                  string
	errConfigRead                  string
	errPasswordEmpty               string
	errPasswordMissing             string
	errAnyDeskMissing              string
	errDialogCreate                string
	errOpenRouterKeyEmpty          string
	errCredentialSave              string
	errCredentialRead              string
	errOpenRouterUnauthorized      string
	errOpenRouterForbidden         string
	errOpenRouterRateLimit         string
	errOpenRouterUpstreamRateLimit string
	errOpenRouterServer            string
	errOpenRouterTimeout           string
	errOpenRouterNetwork           string
	errOpenRouterPayment           string
	errOpenRouterInvalidResponse   string
	errClipboardImageRead          string
	errImageEncode                 string
	errImageAnalysisTimeout        string
	errOpenRouterBadRequest        string
	errOpenRouterModelUnavailable  string
	openRouterResponseDetail       string
	openRouterModelDetail          string
}

var languageMessages = map[string]messages{
	"ko": {
		trayConnect:        "원격 접속",
		trayImageAnalysis:  "이미지 자동 분석",
		trayChangePassword: "기본 암호 수정",
		trayOpenRouter:     "OpenRouter 설정",
		trayStartupAdd:     "시작 프로그램 등록",
		trayStartupRemove:  "시작 프로그램 제거",
		trayLanguage:       "언어",
		trayExit:           "종료",

		connectTitle:       "AnyDesk 원격 접속",
		connectField:       "원격 번호",
		connectDescription: "접속할 AnyDesk 번호를 입력하세요.",
		connectButton:      "접속",
		cancelButton:       "취소",
		saveButton:         "저장",
		verifyButton:       "연결 및 저장",

		changePasswordTitle:       "기본 암호 수정",
		changePasswordDescription: "새 기본 무인 접속 암호를 입력하세요.",
		openRouterTitle:           "OpenRouter 설정",
		openRouterKeyLabel:        "API Key",
		openRouterKeyDescription:  "OpenRouter API Key와 사용할 모델 ID를 입력하세요. API Key는 Windows 자격 증명에 저장합니다.",
		openRouterModelLabel:      "모델 ID",
		openRouterKeysLink:        "OpenRouter API Key 발급 페이지 열기",
		errOpenRouterModelEmpty:   "사용할 OpenRouter 모델 ID를 입력하세요.",
		openRouterSaved:           "OpenRouter 연결을 확인했고 API Key와 모델 설정을 저장했습니다.",
		imagePreviewTitle:         "클립보드 이미지 확인",
		imagePreviewQuestion:      "이 이미지에서 AnyDesk 번호를 분석하시겠습니까?",
		analyzeButton:             "분석",
		ignoreButton:              "무시",
		imagePrivacyNotice:        "이미지 자동 분석을 사용하면 사용자가 분석을 승인한 클립보드 이미지가 AnyDesk 번호 인식을 위해 OpenRouter로 전송됩니다.\n\n계속하시겠습니까?",
		anyDeskNotFoundInImage:    "이미지에서 AnyDesk 번호를 찾지 못했습니다.",
		multipleAnyDeskIDs:        "이미지에서 여러 개의 AnyDesk 번호가 감지되었습니다.\n\n자동 접속을 진행하지 않습니다.",

		setupTitle:       "Quick Anydesk Connect 초기 설정",
		setupPassword:    "기본 암호",
		setupDescription: "최초 실행입니다. 공통 무인 접속 암호를 입력하세요.",

		confirmConnect:  "%s 으로 원격 접속 하시겠습니까?",
		inputCheckTitle: "입력 확인",
		invalidID:       "AnyDesk 번호는 9자리 또는 10자리 숫자여야 합니다.",
		emptyPassword:   "기본 암호를 입력하세요.",

		startupAddSuccess:    "시작 프로그램 등록이 완료되었습니다.",
		startupAddFailure:    "시작 프로그램 등록에 실패했습니다.\n\n%s",
		startupRemoveSuccess: "시작 프로그램 제거가 완료되었습니다.",
		startupRemoveFailure: "시작 프로그램 제거에 실패했습니다.\n\n%s",

		errExecutablePath:              "실행 파일 경로를 확인할 수 없습니다.",
		errTrayClass:                   "트레이 창 클래스를 등록할 수 없습니다.",
		errTrayWindow:                  "트레이 창을 만들 수 없습니다.",
		errClipboardWatch:              "클립보드 감시를 시작할 수 없습니다.",
		errAnyDeskStart:                "AnyDesk 시작에 실패했습니다.\n\n%s",
		errAnyDeskRun:                  "AnyDesk 실행에 실패했습니다.\n\n%s",
		errRegistryOpen:                "시작 프로그램 레지스트리를 열 수 없습니다. (오류 코드: %d)",
		errPathConvert:                 "실행 경로 변환 실패: %v",
		errRegistrySet:                 "시작 프로그램 값을 저장할 수 없습니다. (오류 코드: %d)",
		errRegistryDelete:              "시작 프로그램 값을 제거할 수 없습니다. (오류 코드: %d)",
		errTrayIcon:                    "트레이 아이콘을 추가할 수 없습니다.",
		errConfigSave:                  "설정 파일을 저장할 수 없습니다.\n\n%s",
		errConfigRead:                  "config.ini 읽기 오류:\n\n%s",
		errPasswordEmpty:               "config.ini의 password 값이 비어 있습니다.",
		errPasswordMissing:             "config.ini에서 [anydesk]의 password 값을 찾을 수 없습니다.",
		errAnyDeskMissing:              "AnyDesk 실행 파일을 찾을 수 없습니다.\n\n검색 위치:\nC:\\Program Files (x86)\\AnyDesk-*\\AnyDesk-*.exe\nC:\\Program Files\\AnyDesk-*\\AnyDesk-*.exe",
		errDialogCreate:                "입력창을 만들 수 없습니다.",
		errOpenRouterKeyEmpty:          "OpenRouter API Key를 입력하세요.",
		errCredentialSave:              "OpenRouter API Key를 Windows 자격 증명에 저장하지 못했습니다.\n\n%v",
		errCredentialRead:              "Windows 자격 증명에서 OpenRouter API Key를 읽지 못했습니다.\n\n%v",
		errOpenRouterUnauthorized:      "OpenRouter 인증에 실패했습니다.\n\nAPI Key가 올바른지 확인해주세요.",
		errOpenRouterForbidden:         "OpenRouter 요청 권한이 없습니다.\n\nAPI Key 또는 계정 설정을 확인해주세요.",
		errOpenRouterRateLimit:         "OpenRouter 요청 제한에 도달했습니다.\n\n무료 모델을 사용 중이라면 무료 모델의 계정 한도 또는 요청 제한일 수 있습니다. 잠시 후 다시 시도하거나 다른 모델을 설정해주세요.",
		errOpenRouterUpstreamRateLimit: "선택한 모델의 제공자가 일시적으로 요청 제한 상태입니다.\n\n잠시 후 다시 시도하거나 OpenRouter 설정에서 다른 모델로 변경해주세요.",
		errOpenRouterServer:            "OpenRouter 서버에서 일시적인 오류가 발생했습니다.\n\n잠시 후 다시 시도해주세요.",
		errOpenRouterTimeout:           "OpenRouter 연결 확인 시간이 초과되었습니다.\n\n네트워크 상태를 확인해주세요.",
		errOpenRouterNetwork:           "OpenRouter에 연결할 수 없습니다.\n\n인터넷 연결 상태를 확인해주세요.",
		errOpenRouterPayment:           "OpenRouter 잔액이 부족하거나 결제 제한에 도달했습니다.\n\nOpenRouter 계정의 크레딧 및 사용 한도를 확인해주세요.",
		errOpenRouterInvalidResponse:   "이미지 분석 결과에서 유효한 AnyDesk 번호를 확인하지 못했습니다.",
		errClipboardImageRead:          "클립보드 이미지를 읽을 수 없습니다.",
		errImageEncode:                 "이미지를 분석 가능한 형식으로 변환하지 못했습니다.\n\n%v",
		errImageAnalysisTimeout:        "이미지 분석 요청 시간이 초과되었습니다.\n\n네트워크 상태를 확인하고 다시 시도해주세요.",
		errOpenRouterBadRequest:        "OpenRouter가 이미지 분석 요청을 처리하지 못했습니다.\n\n요청 형식 또는 계정의 모델 사용 설정을 확인해주세요.",
		errOpenRouterModelUnavailable:  "설정된 OpenRouter 이미지 분석 모델을 사용할 수 없습니다.\n\n잠시 후 다시 시도해주세요.",
		openRouterResponseDetail:       "OpenRouter 응답:",
		openRouterModelDetail:          "모델:",
	},
	"en": {
		trayConnect:        "Remote Connection",
		trayImageAnalysis:  "Image Auto Analysis",
		trayChangePassword: "Change Default Password",
		trayOpenRouter:     "OpenRouter Settings",
		trayStartupAdd:     "Add to Startup",
		trayStartupRemove:  "Remove from Startup",
		trayLanguage:       "Language",
		trayExit:           "Exit",

		connectTitle:       "AnyDesk Remote Connection",
		connectField:       "Remote Address",
		connectDescription: "Enter the AnyDesk address to connect to.",
		connectButton:      "Connect",
		cancelButton:       "Cancel",
		saveButton:         "Save",
		verifyButton:       "Verify & Save",

		changePasswordTitle:       "Change Default Password",
		changePasswordDescription: "Enter the new default unattended access password.",
		openRouterTitle:           "OpenRouter Settings",
		openRouterKeyLabel:        "API Key",
		openRouterKeyDescription:  "Enter your OpenRouter API Key and the model ID to use. The API Key is stored in Windows Credential Manager.",
		openRouterModelLabel:      "Model ID",
		openRouterKeysLink:        "Open OpenRouter API Keys page",
		errOpenRouterModelEmpty:   "Enter an OpenRouter model ID to use.",
		openRouterSaved:           "OpenRouter connection verified. API Key and model settings were saved.",
		imagePreviewTitle:         "Clipboard Image",
		imagePreviewQuestion:      "Analyze this image for an AnyDesk address?",
		analyzeButton:             "Analyze",
		ignoreButton:              "Ignore",
		imagePrivacyNotice:        "When Image Auto Analysis is enabled, clipboard images you explicitly approve for analysis are sent to OpenRouter to identify an AnyDesk address.\n\nContinue?",
		anyDeskNotFoundInImage:    "No AnyDesk address was found in the image.",
		multipleAnyDeskIDs:        "Multiple AnyDesk addresses were detected in the image.\n\nAutomatic connection will not continue.",

		setupTitle:       "Quick Anydesk Connect Initial Setup",
		setupPassword:    "Default Password",
		setupDescription: "Enter the shared unattended access password.",

		confirmConnect:  "Connect to %s?",
		inputCheckTitle: "Check Input",
		invalidID:       "The AnyDesk address must be a 9- or 10-digit number.",
		emptyPassword:   "Enter the default password.",

		startupAddSuccess:    "Added Quick Anydesk Connect to startup.",
		startupAddFailure:    "Failed to add Quick Anydesk Connect to startup.\n\n%s",
		startupRemoveSuccess: "Removed Quick Anydesk Connect from startup.",
		startupRemoveFailure: "Failed to remove Quick Anydesk Connect from startup.\n\n%s",

		errExecutablePath:              "Could not determine the executable path.",
		errTrayClass:                   "Could not register the tray window class.",
		errTrayWindow:                  "Could not create the tray window.",
		errClipboardWatch:              "Could not start clipboard monitoring.",
		errAnyDeskStart:                "Failed to start AnyDesk.\n\n%s",
		errAnyDeskRun:                  "Failed to launch AnyDesk.\n\n%s",
		errRegistryOpen:                "Could not open the startup registry key. (Error code: %d)",
		errPathConvert:                 "Failed to convert the executable path: %v",
		errRegistrySet:                 "Could not save the startup registry value. (Error code: %d)",
		errRegistryDelete:              "Could not remove the startup registry value. (Error code: %d)",
		errTrayIcon:                    "Could not add the tray icon.",
		errConfigSave:                  "Could not save the configuration file.\n\n%s",
		errConfigRead:                  "Could not read config.ini.\n\n%s",
		errPasswordEmpty:               "The password value in config.ini is empty.",
		errPasswordMissing:             "Could not find [anydesk] password in config.ini.",
		errAnyDeskMissing:              "Could not find the AnyDesk executable.\n\nSearched:\nC:\\Program Files (x86)\\AnyDesk-*\\AnyDesk-*.exe\nC:\\Program Files\\AnyDesk-*\\AnyDesk-*.exe",
		errDialogCreate:                "Could not create the input window.",
		errOpenRouterKeyEmpty:          "Enter an OpenRouter API Key.",
		errCredentialSave:              "Could not save the OpenRouter API Key to Windows Credential Manager.\n\n%v",
		errCredentialRead:              "Could not read the OpenRouter API Key from Windows Credential Manager.\n\n%v",
		errOpenRouterUnauthorized:      "OpenRouter authentication failed.\n\nCheck that the API Key is correct.",
		errOpenRouterForbidden:         "OpenRouter denied the request.\n\nCheck the API Key or account settings.",
		errOpenRouterRateLimit:         "The OpenRouter request was rate limited.\n\nFor free models, this can be an account-level free-tier limit or another request limit. Try again later or configure another model.",
		errOpenRouterUpstreamRateLimit: "The provider for the selected model is temporarily rate limited.\n\nTry again shortly or choose another model in OpenRouter Settings.",
		errOpenRouterServer:            "OpenRouter returned a temporary server error.\n\nTry again later.",
		errOpenRouterTimeout:           "The OpenRouter connection check timed out.\n\nCheck your network connection.",
		errOpenRouterNetwork:           "Could not connect to OpenRouter.\n\nCheck your internet connection.",
		errOpenRouterPayment:           "OpenRouter has insufficient credit or a spending limit was reached.\n\nCheck your OpenRouter credits and usage limits.",
		errOpenRouterInvalidResponse:   "The image analysis result did not contain a valid AnyDesk address.",
		errClipboardImageRead:          "Could not read the clipboard image.",
		errImageEncode:                 "Could not convert the image into an analyzable format.\n\n%v",
		errImageAnalysisTimeout:        "The image analysis request timed out.\n\nCheck the network connection and try again.",
		errOpenRouterBadRequest:        "OpenRouter could not process the image analysis request.\n\nCheck the request or model permissions on the account.",
		errOpenRouterModelUnavailable:  "The configured OpenRouter image analysis model is unavailable.\n\nTry again later.",
		openRouterResponseDetail:       "OpenRouter response:",
		openRouterModelDetail:          "Model:",
	},
}

func currentMessages() messages {
	if m, ok := languageMessages[language]; ok {
		return m
	}
	return languageMessages["ko"]
}
