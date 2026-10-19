package main

import (
	"fmt"
	"math"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// ============================================================
//  SETTINGS / НАСТРОЙКИ: EDIT THE COMMANDS HERE
//  ЗДЕСЬ ХЕЛПДЕСК РЕДАКТИРУЕТ КОМАНДЫ
// ============================================================
//
//  Each line = one terminal window: {"Window title", "command"}
//  Каждая строка = одно окно терминала: {"Название окна", "команда"}
//
//  EVERY line must end with  },   (curly bracket + comma)
//  КАЖДАЯ строка должна заканчиваться на  },  (скобка и запятая)
//
//  {GATEWAY} is replaced automatically with the default gateway IP.
//  {GATEWAY} автоматически заменяется на IP шлюза по умолчанию.
//
//  Window title: do not use the characters & | "
//  Название окна: не используйте символы & | "
//  In a command, escape & and | like this: ^&  ^|
//  В команде символы & и | экранируйте так: ^&  ^|
// ============================================================

type Terminal struct {
	Title   string
	Command string
}

var terminals = []Terminal{
	{"Ping Google DNS", "ping 8.8.8.8 -t"},
	{"Ping Gateway", "ping {GATEWAY} -t"},
	{"Tracert Google", "tracert 8.8.8.8"},
	{"External IP", "curl ipinfo.io"},

	// --- EXAMPLES TO COPY / ПРИМЕРЫ ДЛЯ КОПИРОВАНИЯ ---
	// Remove the leading // and put your own command or IP.
	// Уберите // в начале строки и впишите свою команду или IP.
	//
	// {"Ping by name", "ping google.com -t"},
	// {"Ping server", "ping 10.0.0.1 -t"},            // put your IP / впишите нужный IP
	// {"Tracert server", "tracert 10.0.0.1"},         // put your IP / впишите нужный IP
	// {"DNS check", "nslookup google.com"},
	// {"Network config", "ipconfig /all"},
	// {"Ping large packet", "ping 8.8.8.8 -t -l 1400"},
	// {"My template", "PUT_YOUR_COMMAND_HERE"},
}

// ============================================================
//  DO NOT EDIT BELOW / НИЖЕ ТРОГАТЬ НЕ НУЖНО
// ============================================================

type rect struct{ Left, Top, Right, Bottom int32 }

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	procEnumWindows          = user32.NewProc("EnumWindows")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procIsWindowVisible      = user32.NewProc("IsWindowVisible")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procSystemParametersInfo = user32.NewProc("SystemParametersInfoW")
	procSetProcessDPIAware   = user32.NewProc("SetProcessDPIAware")

	searchTitle string
	foundHwnd   uintptr
	enumCb      = syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		if v, _, _ := procIsWindowVisible.Call(hwnd); v == 0 {
			return 1
		}
		buf := make([]uint16, 256)
		procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
		if strings.Contains(syscall.UTF16ToString(buf), searchTitle) {
			foundHwnd = hwnd
			return 0 // found, stop enumeration
		}
		return 1
	})
)

func findWindow(tag string) uintptr {
	searchTitle, foundHwnd = tag, 0
	procEnumWindows.Call(enumCb, 0)
	return foundHwnd
}

// Work area of the screen (without the taskbar)
func workArea() rect {
	var r rect
	procSystemParametersInfo.Call(0x0030, 0, uintptr(unsafe.Pointer(&r)), 0) // SPI_GETWORKAREA
	return r
}

// Default gateway IP (parsed from `route print`, independent of Windows language)
func defaultGateway() string {
	cmd := exec.Command("route", "print", "-4", "0.0.0.0")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	best, bestMetric := "", math.MaxInt32
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)
		if len(f) >= 5 && f[0] == "0.0.0.0" && f[1] == "0.0.0.0" && net.ParseIP(f[2]) != nil {
			if m, err := strconv.Atoi(f[4]); err == nil && m < bestMetric {
				best, bestMetric = f[2], m
			}
		}
	}
	return best
}

func main() {
	procSetProcessDPIAware.Call() // real pixels on HiDPI screens

	gw := defaultGateway()
	if gw == "" {
		gw = "GATEWAY_NOT_FOUND" // visible right in the ping window: no cable / no DHCP
	}

	n := len(terminals)
	if n == 0 {
		return
	}

	// 1. Launch windows via conhost.exe so they are classic console windows
	//    and do not merge into Windows Terminal tabs (Windows 11).
	tags := make([]string, n)
	for i, t := range terminals {
		tags[i] = fmt.Sprintf("HD%d:", i+1) // unique tag used to find the window
		command := strings.ReplaceAll(t.Command, "{GATEWAY}", gw)
		cmdline := fmt.Sprintf(`conhost.exe cmd.exe /k title %s %s&&%s`, tags[i], t.Title, command)

		c := exec.Command("conhost.exe")
		c.SysProcAttr = &syscall.SysProcAttr{CmdLine: cmdline}
		if err := c.Start(); err != nil {
			fmt.Println("Failed to start:", t.Title, err)
		}
	}

	// 2. Arrange windows in a grid over the work area
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	rows := int(math.Ceil(float64(n) / float64(cols)))
	area := workArea()
	w := int(area.Right-area.Left) / cols
	h := int(area.Bottom-area.Top) / rows

	for i := range terminals {
		var hwnd uintptr
		for try := 0; try < 50 && hwnd == 0; try++ { // wait up to 5 seconds
			hwnd = findWindow(tags[i])
			if hwnd == 0 {
				time.Sleep(100 * time.Millisecond)
			}
		}
		if hwnd == 0 {
			continue
		}
		x := int(area.Left) + (i%cols)*w
		y := int(area.Top) + (i/cols)*h
		procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), uintptr(w), uintptr(h), 1)
	}
}
