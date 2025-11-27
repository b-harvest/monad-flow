package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"system/journalctl"
	"system/kernel"
	"system/perf"
	"system/scheduler"
	"system/turbostat"

	"system/hook"

	"github.com/joho/godotenv"
	"github.com/zishang520/socket.io/clients/engine/v3/transports"
	"github.com/zishang520/socket.io/clients/socket/v3"
	"github.com/zishang520/socket.io/v3/pkg/types"
)

var hookTargets = []string{
	"_ZN5monad10BlockState6commitERKN4evmc7bytes32ERKNS_11BlockHeaderERKSt6vectorINS_7ReceiptESaIS9_EERKS8_IS8_INS_9CallFrameESaISE_EESaISG_EERKS8_INS1_7addressESaISL_EERKS8_INS_11TransactionESaISQ_EERKS8_IS5_SaIS5_EERKSt8optionalIS8_INS_10WithdrawalESaIS10_EEE",
	// "_ZNK5monad3mpt2Db4findERKNS0_14NodeCursorBaseINS0_4NodeEEENS0_11NibblesViewEm",
	// "_ZNK5monad3mpt2Db4findENS0_11NibblesViewEm",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision0EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision1EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision2EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision3EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision4EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision5EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision6EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision7EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision8EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",
	"_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision9EEEEEN5boost10outcome_v212basic_resultISt6vectorINS_7ReceiptESaIS8_EEN13system_error219errored_status_codeINSB_6detail6erasedIlEEEENS5_12experimental6policy17status_code_throwISA_SG_vEEEERKNS_5ChainERKNS_5BlockESt4spanIKN4evmc7addressELm18446744073709551615EESS_IKS7_ISt8optionalISU_ESaISY_EELm18446744073709551615EERNS_10BlockStateERKNS_15BlockHashBufferERNS_5fiber10FiberGroupERNS_12BlockMetricsESS_ISt10unique_ptrINS_14CallTracerBaseESt14default_deleteIS1E_EELm18446744073709551615EESS_IS1D_ISt7variantIJSt9monostateNS_5trace14PrestateTracerENS1L_15StateDiffTracerENS1L_16AccessListTracerEEES1F_IS1P_EELm18446744073709551615EERKSt8functionIFbRSV_RKNS_11TransactionEmRNS_5StateEEE",

}

type DataPacket struct {
	Source string      `json:"source"`
	Data   interface{} `json:"data"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using OS env vars.")
	}

	rawServices := os.Getenv("TARGET_SERVICES") // 예: "monad-node.service,nginx.service"
	var targetServices []string
	if rawServices != "" {
		parts := strings.Split(rawServices, ",")
		for _, s := range parts {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				targetServices = append(targetServices, trimmed)
			}
		}
	}

	var targetPIDs []string
	if len(targetServices) > 0 {
		var err error
		targetPIDs, err = getPIDsFromServices(targetServices)
		if err != nil {
			log.Printf("[Main] Error extracting PIDs from services: %v", err)
		}
	}
	targetPIDs = uniqueStrings(targetPIDs)

	tracePidStr := os.Getenv("TARGET_TRACE_PID")
	tracePid := 0
	if tracePidStr != "" {
		tracePid, _ = strconv.Atoi(tracePidStr)
	} else if len(targetPIDs) > 0 {
		tracePid, _ = strconv.Atoi(targetPIDs[0])
	}

	socketURL := os.Getenv("BACKEND_URL")
	if socketURL == "" {
		socketURL = "http://127.0.0.1:3000"
	}

	fmt.Println("==========================================")
	fmt.Println("All-in-One System Monitor Starting...")
	fmt.Printf("Target Services: %v\n", targetServices)
	fmt.Printf("Resolved PIDs:   %v\n", targetPIDs)
	fmt.Printf("Hook TargetPID:  %d\n", tracePid) // Hook 대상 확인
	fmt.Printf("Socket Server:   %s\n", socketURL)
	fmt.Println("==========================================")

	mainDataChan := make(chan DataPacket, 1000)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go runWebSocketManager(ctx, &wg, mainDataChan, socketURL)

	offCpuChan := make(chan kernel.OffCPUData, 100)
	schedChan := make(chan scheduler.SchedLog, 100)
	perfChan := make(chan perf.PerfLog, 50)

	go bridge(ctx, offCpuChan, mainDataChan, "OFF_CPU")
	go bridge(ctx, schedChan, mainDataChan, "SCHEDULER")
	go bridge(ctx, perfChan, mainDataChan, "PERF_STAT")

	for _, pid := range targetPIDs {
		wg.Add(1)
		go kernel.Start(ctx, &wg, offCpuChan, kernel.Config{TargetPID: pid})

		wg.Add(1)
		go scheduler.Start(ctx, &wg, schedChan, scheduler.Config{TargetPID: pid})

		wg.Add(1)
		go perf.Start(ctx, &wg, perfChan, perf.Config{TargetPID: pid})
	}

	if tracePid > 0 {
		hookChan := make(chan hook.TraceLog, 100)
		wg.Add(1)
		
		go hook.Start(ctx, &wg, hookChan, hook.Config{
			TargetPID: tracePid,
			Targets:   hookTargets,
		})
		
		go bridge(ctx, hookChan, mainDataChan, "BPF_TRACE")
	} else {
		log.Println("[Main] Skipping Hook Tracer (No valid PID)")
	}

	journalChan := make(chan journalctl.LogEntry, 100)
	wg.Add(1)
	go journalctl.Start(ctx, &wg, journalChan, journalctl.Config{Services: targetServices})
	go bridge(ctx, journalChan, mainDataChan, "SYSTEM_LOG")

	turboChan := make(chan turbostat.TurbostatMetric, 50)
	wg.Add(1)
	go turbostat.Start(ctx, &wg, turboChan)
	go bridge(ctx, turboChan, mainDataChan, "TURBO_STAT")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fmt.Println("\nShutting down system...")

	cancel()
	wg.Wait()
	close(mainDataChan)

	fmt.Println("Monitor system exited cleanly.")
}

func getPIDsFromServices(services []string) ([]string, error) {
	var pids []string
	for _, svc := range services {
		cmd := exec.Command("systemctl", "show", "-p", "MainPID", "--value", svc)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Printf("[Main] Failed to get PID for service %s: %v", svc, err)
			continue
		}
		pidStr := strings.TrimSpace(out.String())
		if pidStr == "" || pidStr == "0" {
			log.Printf("[Main] Service %s is not running or has no PID.", svc)
			continue
		}
		pids = append(pids, pidStr)
		log.Printf("[Main] Found PID %s for service %s", pidStr, svc)
	}
	return pids, nil
}

func uniqueStrings(input []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func bridge[T any](ctx context.Context, input <-chan T, output chan<- DataPacket, sourceName string) {
	for {
		select {
		case data, ok := <-input:
			if !ok {
				return
			}
			output <- DataPacket{
				Source: sourceName,
				Data:   data,
			}
		case <-ctx.Done():
			return
		}
	}
}

func runWebSocketManager(ctx context.Context, wg *sync.WaitGroup, input <-chan DataPacket, url string) {
	defer wg.Done()

	opts := socket.DefaultOptions()
	opts.SetTransports(types.NewSet(
		transports.WebSocket,
	))

	client, err := socket.Connect(url, opts)
	if err != nil {
		log.Printf("Initial Socket Connection Failed: %v (Will try to reconnect...)", err)
	}

	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	fmt.Println("WebSocket Sender Started. Forwarding data...")
	for {
		select {
		case packet, ok := <-input:
			if !ok {
				return
			}
			if client.Connected() {
				client.Emit(packet.Source, packet.Data)
			}
		case <-ctx.Done():
			return
		}
	}
}
