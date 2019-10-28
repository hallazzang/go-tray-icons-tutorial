package main

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const TrayIconMsg = WM_APP + 1

func wndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case TrayIconMsg:
		switch nmsg := LOWORD(uint32(lParam)); nmsg {
		case NIN_BALLOONUSERCLICK:
			fmt.Println("user clicked the balloon notification")
		case WM_LBUTTONDOWN:
			fmt.Println("user clicked the tray icon")
		}
	case WM_DESTROY:
		PostQuitMessage(0)
	default:
		r, _ := DefWindowProc(hWnd, msg, wParam, lParam)
		return r
	}
	return 0
}

func createMainWindow() (uintptr, error) {
	hInstance, err := GetModuleHandle(nil)
	if err != nil {
		return 0, err
	}

	wndClass := windows.StringToUTF16Ptr("MyWindow")

	var wcex WNDCLASSEX

	wcex.CbSize = uint32(unsafe.Sizeof(wcex))
	wcex.LpfnWndProc = windows.NewCallback(wndProc)
	wcex.HInstance = hInstance
	wcex.LpszClassName = wndClass
	if _, err := RegisterClassEx(&wcex); err != nil {
		return 0, err
	}

	hwnd, err := CreateWindowEx(
		0,
		wndClass,
		windows.StringToUTF16Ptr("Tray Icons Example"),
		WS_OVERLAPPEDWINDOW,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		400,
		300,
		0,
		0,
		hInstance,
		nil)
	if err != nil {
		return 0, err
	}

	return hwnd, nil
}

func newGUID() GUID {
	var buf [16]byte
	rand.Read(buf[:])
	return *(*GUID)(unsafe.Pointer(&buf[0]))
}

type TrayIcon struct {
	hwnd uintptr
	guid GUID
}

func NewTrayIcon(hwnd uintptr) (*TrayIcon, error) {
	ti := &TrayIcon{hwnd: hwnd, guid: newGUID()}
	data := ti.initData()
	data.UFlags |= NIF_MESSAGE
	data.UCallbackMessage = TrayIconMsg
	if _, err := Shell_NotifyIcon(NIM_ADD, data); err != nil {
		return nil, err
	}
	return ti, nil
}

func (ti *TrayIcon) initData() *NOTIFYICONDATA {
	var data NOTIFYICONDATA
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.UFlags = NIF_GUID
	data.HWnd = ti.hwnd
	data.GUIDItem = ti.guid
	return &data
}

func (ti *TrayIcon) Dispose() error {
	_, err := Shell_NotifyIcon(NIM_DELETE, ti.initData())
	return err
}

func (ti *TrayIcon) SetIcon(icon uintptr) error {
	data := ti.initData()
	data.UFlags |= NIF_ICON
	data.HIcon = icon
	_, err := Shell_NotifyIcon(NIM_MODIFY, data)
	return err
}

func (ti *TrayIcon) SetTooltip(tooltip string) error {
	data := ti.initData()
	data.UFlags |= NIF_TIP
	copy(data.SzTip[:], windows.StringToUTF16(tooltip))
	_, err := Shell_NotifyIcon(NIM_MODIFY, data)
	return err
}

func (ti *TrayIcon) ShowBalloonNotification(title, text string) error {
	data := ti.initData()
	data.UFlags |= NIF_INFO
	if title != "" {
		copy(data.SzInfoTitle[:], windows.StringToUTF16(title))
	}
	copy(data.SzInfo[:], windows.StringToUTF16(text))
	_, err := Shell_NotifyIcon(NIM_MODIFY, data)
	return err
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	hwnd, err := createMainWindow()
	if err != nil {
		panic(err)
	}

	icon, err := LoadImage(
		0,
		windows.StringToUTF16Ptr("icon.ico"),
		IMAGE_ICON,
		0,
		0,
		LR_DEFAULTSIZE|LR_LOADFROMFILE)
	if err != nil {
		panic(err)
	}

	ti, err := NewTrayIcon(hwnd)
	if err != nil {
		panic(err)
	}
	defer ti.Dispose()

	ti.SetIcon(icon)
	ti.SetTooltip("Tray Icon!")

	go func() {
		for i := 1; i <= 3; i++ {
			time.Sleep(3 * time.Second)
			ti.ShowBalloonNotification(
				fmt.Sprintf("Message %d", i),
				"This is a balloon message",
			)
		}
	}()

	ShowWindow(hwnd, SW_SHOW)

	var msg MSG
	for {
		if r, err := GetMessage(&msg, 0, 0, 0); err != nil {
			panic(err)
		} else if r == 0 {
			break
		}
		TranslateMessage(&msg)
		DispatchMessage(&msg)
	}
}
