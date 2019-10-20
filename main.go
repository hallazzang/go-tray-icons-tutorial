package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const notifyIconMsg = WM_APP + 1

func wndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case notifyIconMsg:
		switch nmsg := LOWORD(uint32(lParam)); nmsg {
		case NIN_BALLOONUSERCLICK:
			fmt.Println("balloon clicked")
		case WM_LBUTTONDOWN:
			fmt.Println("tray icon clicked")
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
		windows.StringToUTF16Ptr("NotifyIcon Example"),
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

func main() {
	hwnd, err := createMainWindow()
	if err != nil {
		panic(err)
	}

	var data NOTIFYICONDATA

	data.CbSize = uint32(unsafe.Sizeof(data))
	data.UFlags = NIF_ICON | NIF_MESSAGE | NIF_TIP | NIF_INFO
	data.HWnd = hwnd
	data.UCallbackMessage = notifyIconMsg
	copy(data.SzTip[:], windows.StringToUTF16("Tray Icon"))
	copy(data.SzInfo[:], windows.StringToUTF16("Hello from Tay Icon!"))

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
	data.HIcon = icon

	if _, err := Shell_NotifyIcon(NIM_ADD, &data); err != nil {
		panic(err)
	}

	ShowWindow(hwnd, SW_SHOW)

	var msg MSG
	for {
		r, err := GetMessage(&msg, 0, 0, 0)
		if err != nil {
			panic(err)
		}
		if r == 0 {
			break
		}

		TranslateMessage(&msg)
		DispatchMessage(&msg)
	}
}
