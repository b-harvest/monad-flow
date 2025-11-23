package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/frida/frida-go/frida"
)

// --- [Frida 메시지 껍데기 구조체] ---
type FridaEnvelope struct {
	Type    string          `json:"type"`    // "send" 또는 "error"
	Payload json.RawMessage `json:"payload"` // 실제 우리가 보낸 데이터
}

// --- [사용자 정의 데이터 구조체] ---
type TraceLog struct {
	EventType  string          `json:"event_type"`
	FuncName   string          `json:"func_name"`
	PID        string          `json:"pid"`
	TID        string          `json:"tid"`
	DurationNs string          `json:"duration_ns,omitempty"`
	Data       json.RawMessage `json:"data"`
}

type EnterData struct {
	Timestamp      string   `json:"timestamp"`
	CallerFuncName string   `json:"caller_name"`
	Args           []string `json:"args_hex"`
}

type ExitData struct {
	Timestamp      string `json:"timestamp"`
	BackToFuncName string `json:"back_to_name"`
	ReturnValue    string `json:"return_value"`
}

var targetFunctions = []string{
	"_ZN5monad10BlockState6commitERKN4evmc7bytes32ERKNS_11BlockHeaderERKSt6vectorINS_7ReceiptESaIS9_EERKS8_IS8_INS_9CallFrameESaISE_EESaISG_EERKS8_INS1_7addressESaISL_EERKS8_INS_11TransactionESaISQ_EERKS8_IS5_SaIS5_EERKSt8optionalIS8_INS_10WithdrawalESaIS10_EEE",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sudo ./hooker <PID>")
		os.Exit(1)
	}
	pid, _ := strconv.Atoi(os.Args[1])

	manager := frida.NewDeviceManager()
	devices, _ := manager.EnumerateDevices()
	localDevice := devices[0]

	fmt.Printf("[*] Attaching to PID: %d ...\n", pid)
	session, err := localDevice.Attach(pid, nil)
	if err != nil {
		panic(err)
	}
	defer session.Detach()

	jsCode := generateJSPayload(targetFunctions)
	script, _ := session.CreateScript(jsCode)

	// --- [메시지 핸들러 수정됨] ---
	script.On("message", func(message string) {
		// 1. 먼저 껍데기(Envelope)를 깝니다.
		var envelope FridaEnvelope
		if err := json.Unmarshal([]byte(message), &envelope); err != nil {
			// JSON 형식이 아니면 그냥 출력
			fmt.Printf("[Raw] %s\n", message)
			return
		}

		// 2. 메시지 타입이 'send'인 경우만 처리
		if envelope.Type == "send" {
			var logEntry TraceLog
			// Payload를 다시 파싱합니다.
			if err := json.Unmarshal(envelope.Payload, &logEntry); err != nil {
				fmt.Printf("[Error Parsing Payload] %s\n", string(envelope.Payload))
				return
			}

			// 3. 로그 출력 로직
			fmt.Println("---------------------------------------------------")
			fmt.Printf("[%s] Func: %s (TID: %s)\n", strings.ToUpper(logEntry.EventType), shortenName(logEntry.FuncName), logEntry.TID)

			if logEntry.EventType == "enter" {
				var d EnterData
				json.Unmarshal(logEntry.Data, &d)
				fmt.Printf("   ↳ Caller: %s\n", shortenName(d.CallerFuncName))
				fmt.Printf("   ↳ Args:   %v\n", d.Args)
				fmt.Printf("   ↳ Time:   %s\n", d.Timestamp)
			} else if logEntry.EventType == "exit" {
				var d ExitData
				json.Unmarshal(logEntry.Data, &d)
				fmt.Printf("   ↳ Return: %s\n", d.ReturnValue)
				fmt.Printf("   ↳ Dur:    %s ns\n", logEntry.DurationNs)
			}
		} else if envelope.Type == "error" {
			// JS 실행 중 에러 발생 시
			fmt.Printf("[Frida JS Error] %s\n", string(envelope.Payload))
		}
	})

	script.Load()
	fmt.Println("[*] Hooks installed. Press Ctrl+C to stop.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func shortenName(fullName string) string {
	if len(fullName) > 60 {
		return fullName[:25] + "..." + fullName[len(fullName)-25:]
	}
	return fullName
}

func generateJSPayload(symbols []string) string {
	quotedSymbols := make([]string, len(symbols))
	for i, s := range symbols {
		quotedSymbols[i] = fmt.Sprintf("'%s'", s)
	}
	jsArrayStr := strings.Join(quotedSymbols, ", ")

	return fmt.Sprintf(`
		const targetSymbols = [%s];
		const pid = Process.id.toString();

		function hookFunction(symbolName) {
			let targetAddr = DebugSymbol.getFunctionByName(symbolName);
			if (!targetAddr) targetAddr = Module.findExportByName(null, symbolName);
			if (!targetAddr) { 
                // 에러도 send로 보내서 Go에서 찍히게 함
                send({event_type: "error", msg: "Cannot find symbol: " + symbolName});
                return; 
            }

			Interceptor.attach(targetAddr, {
				onEnter: function(args) {
					this.tid = Process.getCurrentThreadId().toString();
					this.startTime = new Date(); 
					const tsString = this.startTime.toISOString();

					let callerSym = DebugSymbol.fromAddress(this.returnAddress);
					let callerName = callerSym.name ? callerSym.name : this.returnAddress.toString();

					let argsList = [];
					for (let i = 0; i < 5; i++) {
						try { argsList.push(args[i].toString()); } catch (e) { argsList.push("0x0"); }
					}

					send({
						event_type: "enter",
						func_name: symbolName,
						pid: pid,
						tid: this.tid,
						duration_ns: "0",
						data: {
							timestamp: tsString,
							caller_name: callerName,
							args_hex: argsList
						}
					});
				},

				onLeave: function(retval) {
					const endTime = new Date();
					const durationMs = endTime.getTime() - this.startTime.getTime();
					const durationNs = (durationMs * 1000000).toString();

					send({
						event_type: "exit",
						func_name: symbolName,
						pid: pid,
						tid: this.tid,
						duration_ns: durationNs,
						data: {
							timestamp: endTime.toISOString(),
							back_to_name: "caller", 
							return_value: retval.toString()
						}
					});
				}
			});
		}
		targetSymbols.forEach(hookFunction);
	`, jsArrayStr)
}