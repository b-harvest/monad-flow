package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/frida/frida-go/frida"
)

type Config struct {
	TargetPID int
	Targets   []string // 후킹할 함수 목록
}

type TraceLog struct {
	EventType  string          `json:"event_type"` // "enter" | "exit"
	FuncName   string          `json:"func_name"`
	PID        string          `json:"pid"`
	Timestamp  string          `json:"timestamp"` // 여기로 이동
	DurationNs string          `json:"duration_ns,omitempty"`
	Data       json.RawMessage `json:"data"`
}

type EnterData struct {
	CallerFuncName string   `json:"caller_name"`
	Args           []string `json:"args_hex"`
}

type ExitData struct {
	BackToFuncName string `json:"back_to_name"`
	ReturnValue    string `json:"return_value"`
}

type fridaEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func Start(ctx context.Context, wg *sync.WaitGroup, outChan chan<- TraceLog, cfg Config) {
	defer wg.Done()

	if cfg.TargetPID <= 0 {
		log.Println("[Hook] Invalid PID, skipping hooker.")
		return
	}

	manager := frida.NewDeviceManager()
	devices, err := manager.EnumerateDevices()
	if err != nil || len(devices) == 0 {
		log.Printf("[Hook] Failed to enumerate devices: %v", err)
		return
	}
	localDevice := devices[0]

	log.Printf("[Hook] Attaching to PID: %d ...", cfg.TargetPID)
	session, err := localDevice.Attach(cfg.TargetPID, nil)
	if err != nil {
		log.Printf("[Hook] Failed to attach: %v", err)
		return
	}
	
	defer func() {
		session.Detach()
		log.Println("[Hook] Detached from process.")
	}()

	jsCode := generateJSPayload(cfg.Targets)
	script, err := session.CreateScript(jsCode)
	if err != nil {
		log.Printf("[Hook] Failed to create script: %v", err)
		return
	}

	script.On("message", func(message string) {
		var envelope fridaEnvelope
		if err := json.Unmarshal([]byte(message), &envelope); err != nil {
			return
		}

		if envelope.Type == "send" {
			var logEntry TraceLog
			if err := json.Unmarshal(envelope.Payload, &logEntry); err != nil {
				return
			}

			if len(logEntry.FuncName) > 70 {
				logEntry.FuncName = logEntry.FuncName[:67] + "..."
			}
			select {
			case outChan <- logEntry:
			case <-ctx.Done():
			}
		}
	})

	if err := script.Load(); err != nil {
		log.Printf("[Hook] Failed to load script: %v", err)
		return
	}

	log.Println("[Hook] Scripts loaded. Monitoring started.")

	<-ctx.Done()
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

        let clock_gettime_addr = Module.findExportByName(null, 'clock_gettime');
        if (!clock_gettime_addr) clock_gettime_addr = Module.findExportByName("libc.so.6", "clock_gettime");
        
        let clock_gettime = null;
        if (clock_gettime_addr) {
            clock_gettime = new NativeFunction(clock_gettime_addr, 'int', ['int', 'pointer']);
        }

        const CLOCK_MONOTONIC = 1;
        const timespecSize = 16;

        function hookFunction(symbolName) {
            let targetAddr = DebugSymbol.getFunctionByName(symbolName);
            if (!targetAddr) targetAddr = Module.findExportByName(null, symbolName);
            if (!targetAddr) return; 

            Interceptor.attach(targetAddr, {
                onEnter: function(args) {
                    if (clock_gettime) {
                        this.startTs = Memory.alloc(timespecSize);
                        clock_gettime(CLOCK_MONOTONIC, this.startTs);
                    } else {
                        this.startTimeDate = new Date();
                    }

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
                        timestamp: new Date().toISOString(), 
                        duration_ns: "0",
                        data: {
                            caller_name: callerName,
                            args_hex: argsList
                        }
                    });
                },

                onLeave: function(retval) {
                    let durationBn = "0";

                    if (clock_gettime && this.startTs) {
                        const endTs = Memory.alloc(timespecSize);
                        clock_gettime(CLOCK_MONOTONIC, endTs);
                        
                        const startSec = BigInt(this.startTs.readU64().toString());
                        const startNsec = BigInt(this.startTs.add(8).readU64().toString());
                        
                        const endSec = BigInt(endTs.readU64().toString());
                        const endNsec = BigInt(endTs.add(8).readU64().toString());

                        const billion = BigInt(1000000000);
                        const startTotal = (startSec * billion) + startNsec;
                        const endTotal = (endSec * billion) + endNsec;
                        
                        durationBn = (endTotal - startTotal).toString();

                    } else if (this.startTimeDate) {
                        durationBn = ((new Date().getTime() - this.startTimeDate.getTime()) * 1000000).toString();
                    }

                    send({
                        event_type: "exit",
                        func_name: symbolName,
                        pid: pid,
                        timestamp: new Date().toISOString(),
                        duration_ns: durationBn,
                        data: {
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