// startup_task_xml_windows.go — 开机自启任务域：XML 定义辅助函数（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/pipeline/tmp/{launcherStartupTaskLogonDelay,launcherStartupTaskWorkingDirectory,
// encodeUTF16LEWithBOM,quoteWindowsTaskActionCommand,currentLauncherStartupTaskUserID,
// buildLauncherStartupTaskDefinitionXML,convertXMLFileToUTF16LE}.asm.txt
//
// 重要修正：本 EXE 无字符串混淆。此前标注 [P 混淆] 的常量经 read_gostring.py
// 正确 va_to_off 复核，全部为明文（此前 pefile.get_offset_from_rva 对 vsize<rawsz
// 的节映射错误，误读为相邻 runtime 字符串片段）。
//
// 本域函数：
//   launcherStartupTaskLogonDelay        (0x1408b24c0) logon 触发延时 → PT%dS
//   launcherStartupTaskWorkingDirectory  (0x1408b2540) 工作目录推断
//   quoteWindowsTaskActionCommand        (0x1408b2960) 命令加引号转义
//   encodeUTF16LEWithBOM                 (0x1408b2600) UTF-16LE + BOM 编码
//   currentLauncherStartupTaskUserID     (0x1408b2a60) 当前用户标识
//   buildLauncherStartupTaskDefinitionXML (0x1408b2080) 任务 XML 定义
//   convertXMLFileToUTF16LE              (0x1408b2740) XML 转 UTF-16LE 临时文件

package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

// launcherStartupTaskLogonDelay 生成 logon 触发延时（Task Scheduler ISO8601 秒格式）。
// [S 汇编 0x1408b24c0, 128B(0x80)] delaySeconds<=0 → 10（@0x1408b24ce/d3）；>100 → 100
// （@0x1408b24e0/e6）；fmt.Sprintf("PT%dS",delaySeconds)（@0x1408b2520，格式串 5B
// @0x140c35d1c）。
func launcherStartupTaskLogonDelay(delaySeconds int) string {
	if delaySeconds <= 0 {
		delaySeconds = 10
	} else if delaySeconds > 100 {
		delaySeconds = 100
	}
	return fmt.Sprintf("PT%dS", delaySeconds)
}

// launcherStartupTaskWorkingDirectory 推断启动任务工作目录。
// [S 汇编 0x1408b2540, 192B(0xc0)] dir=filepath.Dir(exe)（@0x1408b2558）；TrimSpace(dir)
// 空（@0x1408b2567/6f）或 dir=="."（len==1 且首字节 0x2e，@0x1408b2571/84）→
// filepath.Dir(os.Args[0])（@0x1408b2586 全局 []string[0] 推断为 os.Args）；否则返回 dir。
func launcherStartupTaskWorkingDirectory(exe string) string {
	dir := filepath.Dir(exe)
	if strings.TrimSpace(dir) == "" || dir == "." {
		return filepath.Dir(os.Args[0])
	}
	return dir
}

// quoteWindowsTaskActionCommand 给任务动作命令加引号并转义内部引号。
// [S 汇编 0x1408b2960, 256B(0x100)] TrimSpace（@0x1408b2977）空→""（@0x1408b2983/2a1b）；
// 首尾皆 '"'（@0x1408b298a/29a2 memequal）→ 原样返回；否则 `"` + Replace(`"`→`\"`,-1) + `"`
// （@0x1408b29c1/29e1/2a00 concatstring3）。
func quoteWindowsTaskActionCommand(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}
	if command[0] == '"' && command[len(command)-1] == '"' {
		return command
	}
	return `"` + strings.Replace(command, `"`, `\"`, -1) + `"`
}

// encodeUTF16LEWithBOM 把字符串编码为带 BOM 的 UTF-16LE 字节序。
// [S 汇编 0x1408b2600, 320B(0x140)] []rune(content)（@0x1408b2640）→utf16.Encode
// （@0x1408b2645）；make([]byte,2,len(u16)*2+2)（@0x1408b265a/676）；out[0..1]=0xff,0xfe
// （@0x1408b267b mov word 0xfeff 小端）；循环逐 uint16 小端复制（@0x1408b26a2-2708）。
func encodeUTF16LEWithBOM(content string) []byte {
	u16 := utf16.Encode([]rune(content))
	out := make([]byte, 2, len(u16)*2+2)
	out[0], out[1] = 0xff, 0xfe
	for _, u := range u16 {
		out = append(out, byte(u), byte(u>>8))
	}
	return out
}

// currentLauncherStartupTaskUserID 解析当前用户的启动任务身份标识。
// [S 汇编 0x1408b2a60, 192B(0xc0)] os/user.Current()（@0x1408b2a72）err → ("" ,err)；
// TrimSpace(Uid)（@0x1408b2a8b）非空 → (Uid,nil)；否则 TrimSpace(Username)
// （@0x1408b2a9e）非空 → (Username,nil)；都空 → ("" ,errors.New("无法获取当前用户标识"))。
func currentLauncherStartupTaskUserID() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	if id := strings.TrimSpace(u.Uid); id != "" {
		return id, nil
	}
	if name := strings.TrimSpace(u.Username); name != "" {
		return name, nil
	}
	return "", errors.New("无法获取当前用户标识")
}

// buildLauncherStartupTaskDefinitionXML 生成开机自启任务的 XML 定义（UTF-8，带 XML 声明头）。
// [S 汇编 0x1408b2080, 1088B(0x440)] 校验与依赖链 + 模板 duffcopy 已证：
//
//	TrimSpace(exe) 空→errors.New("开机启动程序路径不能为空" 36B @0x140c784e3)
//	  （@0x1408b20d5/e3/2414）；
//	quoteWindowsTaskActionCommand(exe)（@0x1408b20f6）；
//	TrimSpace(userID) 空→errors.New("开机启动任务用户标识不能为空" 42B @0x140c80e94)
//	  （@0x1408b2118/23/23dc）；
//	TrimSpace(args)（@0x1408b2146）；
//	newobject(*int)+mov [rax],7（Settings.Priority=7 @0x1408b2158/6d）；
//	launcherStartupTaskLogonDelay(delaySeconds)（@0x1408b2180）；
//	launcherStartupTaskWorkingDirectory(exe)（@0x1408b21a0）；
//	LogonType：enabled→"S4U"(3B @0x140c33d0d)，否则 "InteractiveToken"(16B @0x140c551fa)
//	  （@0x1408b2215/34 cmovne）；Settings.Hidden=enabled（@0x1408b2248，相对 def+0xda）；
//	duffcopy 静态模板 @0x1411e4ea0（def+0x00 起，@0x1408b21d2）后覆盖动态字段；
//	xml.MarshalIndent(def,"","  " 2B @0x140c3366b)（@0x1408b22d5）；
//	前缀 39B 头 "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"（@0x1408b230e-2345 movabs）。
//
// 模板 @0x1411e4ea0 已解引用（批次 130）：Version="1.2"、Xmlns=2004/02/mit/task、
// LogonTrigger.Enabled=true、Principal.ID/Context="Author"、RunLevel="HighestAvailable"、
// MultipleInstancesPolicy="IgnoreNew"、IdleSettings(Duration="PT10M"/WaitTimeout="PT1H"/
// StopOnIdleEnd=true/RestartOnIdle=false)、AllowStartOnDemand=true、Settings.Enabled=true、
// Hidden 模板 0 但 asm 覆盖为 enabled、ExecutionTimeLimit="PT0S"、Priority 模板 nil 但 asm
// 覆盖为 7。动态字段 UserID/LogonType/Delay/Command/Arguments/WorkingDirectory 由 asm 写入。
func buildLauncherStartupTaskDefinitionXML(exe, args, userID string, enabled bool, delaySeconds int) ([]byte, error) {
	exe = strings.TrimSpace(exe)
	if exe == "" {
		return nil, errors.New("开机启动程序路径不能为空")
	}
	quotedExe := quoteWindowsTaskActionCommand(exe)

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("开机启动任务用户标识不能为空")
	}
	args = strings.TrimSpace(args)

	logonType := "InteractiveToken"
	if enabled {
		logonType = "S4U"
	}

	def := launcherStartupTaskXML{
		Version: "1.2",
		Xmlns:   "http://schemas.microsoft.com/windows/2004/02/mit/task",
		Triggers: launcherStartupTaskTriggersXML{
			LogonTrigger: launcherStartupTaskEnabledXML{
				Enabled: true,
				Delay:   launcherStartupTaskLogonDelay(delaySeconds),
			},
		},
		Principals: launcherStartupTaskPrincipalsXML{
			Principal: launcherStartupTaskPrincipalXML{
				ID:        "Author",
				UserID:    userID,
				LogonType: logonType,
				RunLevel:  "HighestAvailable",
			},
		},
		Settings: launcherStartupTaskSettingsXML{
			MultipleInstancesPolicy:    "IgnoreNew",
			DisallowStartIfOnBatteries: false,
			StopIfGoingOnBatteries:     false,
			AllowHardTerminate:         false,
			StartWhenAvailable:         false,
			RunOnlyIfNetworkAvailable:  false,
			IdleSettings: launcherStartupTaskIdleSettingsXML{
				Duration:      "PT10M",
				WaitTimeout:   "PT1H",
				StopOnIdleEnd: true,
				RestartOnIdle: false,
			},
			AllowStartOnDemand: true,
			Enabled:            true,
			Hidden:             enabled,
			RunOnlyIfIdle:      false,
			WakeToRun:          false,
			ExecutionTimeLimit: "PT0S",
			Priority:           intPtr(7),
		},
		Actions: launcherStartupTaskActionsXML{
			Context: "Author",
			Exec: launcherStartupTaskExecXML{
				Command:          quotedExe,
				Arguments:        args,
				WorkingDirectory: launcherStartupTaskWorkingDirectory(exe),
			},
		},
	}

	body, err := xml.MarshalIndent(def, "", "  ")
	if err != nil {
		return nil, err
	}
	header := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	return append(header, body...), nil
}

// intPtr 返回指向 v 的 *int。
// [S-inline] 对应 newobject（@0x1408b2158）+ mov [rax],7（@0x1408b216d）的 Settings.Priority=7 内联填充。
func intPtr(v int) *int {
	return &v
}

// convertXMLFileToUTF16LE 把 XML 文件转成带 BOM 的 UTF-16LE 临时文件。
// [S 汇编 0x1408b2740, 480B(0x1e0)] os.ReadFile（@0x1408b276a）err→(nil,nil,err)；
// bytes.Replace(内容, 39B UTF-8 声明头, "<?xml version=\"1.0\" encoding=\"UTF-16\"?>\n" 40B
// @0x140c7f142, 1)（@0x1408b27e8）；临时路径 = path + ".utf16.xml" 10B @0x140c441cb
// （@0x1408b2823 concatstring2）；encodeUTF16LEWithBOM（@0x1408b284d）；
// os.WriteFile(临时路径, 数据, 0x180)（@0x1408b286e）err→(nil,nil,err)；
// 返回 (临时路径, func(){os.Remove(临时路径)}, nil)（func3 0x1408b2920）。
func convertXMLFileToUTF16LE(path string) (string, func(), error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	content = bytes.Replace(
		content,
		[]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"),
		[]byte("<?xml version=\"1.0\" encoding=\"UTF-16\"?>\n"),
		1,
	)
	tmpPath := path + ".utf16.xml"
	data := encodeUTF16LEWithBOM(string(content))
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return "", nil, err
	}
	return tmpPath, func() { os.Remove(tmpPath) }, nil
}
