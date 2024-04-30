package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"golang.org/x/time/rate"
)

type Config struct {
	PrivateKey  string  `json:"private_key"`
	RpcUrl      string  `json:"rpc_url"`
	WsUrl       string  `json:"ws_url"`
	SendRpcUrl  string  `json:"send_rpc_url"`
	RateLimit   uint64  `json:"rate_limit"`
	TxCount     uint64  `json:"tx_count"`
	PrioFee     float64 `json:"prio_fee"`
	NodeRetries uint    `json:"node_retries"`
}

func (c *Config) GetWsUrl() string {
	if c.WsUrl != "" {
		return c.WsUrl
	}

	// replace http:// with ws:// and https:// with wss://
	return strings.ReplaceAll(strings.ReplaceAll(c.RpcUrl, "http://", "ws://"), "https://", "wss://")
}

func (c *Config) GetSendUrl() string {
	if c.SendRpcUrl != "" {
		return c.SendRpcUrl
	}

	return c.RpcUrl
}

type WebsocketListener struct {
	Subscription *ws.LogSubscription
	Listening    bool
}

func (l *WebsocketListener) Start() {
	wsClient, err := ws.Connect(context.TODO(), GlobalConfig.GetWsUrl())
	if err != nil {
		log.Fatalf("error connecting to websocket: %v", err)
	}

	defer wg.Done()

	l.Subscription, err = wsClient.LogsSubscribeMentions(TestAccount.PublicKey(), rpc.CommitmentProcessed)
	if err != nil {
		log.Fatalf("error subscribing to logs: %v", err)
	}
	l.Listening = true

	log.Info("Listening for transactions...")

	// start sending transactions now that the websocket is ready
	SendTransactions()

	for l.Listening {
		got, err := l.Subscription.Recv()
		if err != nil {
			log.Error(err.Error())
		}

		if got == nil || got.Value.Err != nil {
			continue
		}

		re := regexp.MustCompile(`memobench:.*?(\d+).*\[(.*?)\]`)
		for _, line := range got.Value.Logs {
			matches := re.FindStringSubmatch(line)
			if len(matches) != 3 {
				continue
			}
			testNum, id := matches[1], matches[2]

			if id != TestID {
				log.Warn(
					"Received unexpected test ID",
					"num", testNum,
					"id", id,
					"sig", got.Value.Signature.String(),
				)
				continue
			}

			mu.Lock()
			txSendTime, found := TxTimes[got.Value.Signature]
			if found {
				ProcessedTransactions += 1
			}
			mu.Unlock()

			// skip this tx if it's not in the TxTimes map
			// this could happen if the test was restarted and a tx from a previous test landed
			if !found {
				continue
			}

			delta := time.Since(txSendTime)
			log.Info(
				"Tx Processed",
				"num", testNum,
				"sig", got.Value.Signature.String(),
				"delta", delta.Truncate(time.Millisecond),
				"landed", fmt.Sprintf("%d/%d", ProcessedTransactions, SentTransactions),
			)

			if ProcessedTransactions >= SentTransactions {
				l.Stop()
			}
			break
		}
	}

	log.Info("Stopping listening for log events...")
}

func (l *WebsocketListener) Stop() {
	l.Listening = false
	l.Subscription.Unsubscribe()
}

var (
	DEFAULT_CONFIG = Config{
		RpcUrl:    "http://node.foo.cc",
		RateLimit: 200,
		TxCount:   100,
		PrioFee:   0,
	}

	TestID string

	// variable for the log file; set to benchmark.log as a fallback
	LogFileName string = "benchmark.log"

	GlobalConfig *Config
	TestAccount  *solana.PrivateKey

	wg sync.WaitGroup
	mu sync.RWMutex

	// the rate limiter
	Limiter = rate.NewLimiter(rate.Limit(200), 200)

	// the time the test should end
	StopTime time.Time

	// the number of transactions sent and transactions that landed
	SentTransactions      uint64
	ProcessedTransactions uint64

	// transaction send times
	TxTimes = make(map[solana.Signature]time.Time)

	WsListener *WebsocketListener
)

func SetupLogger() {
	LogFileName = fmt.Sprintf("memobench_%d_%s.log", time.Now().UnixMilli(), TestID)
	logFile, err := os.OpenFile(LogFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	logger := log.NewWithOptions(io.MultiWriter(os.Stdout, logFile), log.Options{
		Prefix:          TestID,
		ReportTimestamp: true,
		TimeFunction:    func(time.Time) time.Time { return time.Now().UTC() },
		TimeFormat:      "15:04:05.0000",
	})

	log.SetDefault(logger)
}

func ReadConfig() *Config {
	data, err := os.ReadFile("config.json")
	if err != nil {
		// if the error is that the file doesn't exist, create it, and exit
		if os.IsNotExist(err) {
			if err := WriteConfig(&DEFAULT_CONFIG); err != nil {
				log.Fatalf("error creating config file: %v", err)
			}

			log.Info("config file saved, edit the config and restart")
			os.Exit(0)
		}

		log.Fatalf("error opening config file: %v", err)
	}

	var out Config

	err = json.Unmarshal(data, &out)
	if err != nil {
		log.Fatalf("error parsing config file: %v", err)
	}

	return &out
}

func WriteConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("error saving config file: %v", err)
	}

	return os.WriteFile("config.json", data, 0644)
}

func VerifyPrivateKey(base58key string) {
	account, err := solana.PrivateKeyFromBase58(base58key)
	if err != nil {
		log.Fatalf("error parsing private key: %v", err)
	}
	TestAccount = &account
}

func SendTransactions() {
	// Create a new RPC client:
	rpcClient := rpc.New(GlobalConfig.RpcUrl)

	// create the send client
	sendClient := rpc.New(GlobalConfig.GetSendUrl())

	// fetch the latest blockhash
	recent, err := rpcClient.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("error getting recent blockhash: %v", err)
	}

	// save current time and set the experiment end time
	// hash expire after 150 blocks, each block is about 400ms
	// we use 160 blocks just out of abundance of caution
	StopTime = time.Now().Add(160 * 400 * time.Millisecond)
	time.AfterFunc(time.Until(StopTime), WsListener.Stop)

	for i := uint64(0); i < GlobalConfig.TxCount; i++ {
		go func(id uint64) {
			instructions := []solana.Instruction{}

			if GlobalConfig.PrioFee > 0 {
				instructions = append(instructions, computebudget.NewSetComputeUnitPriceInstruction(uint64(GlobalConfig.PrioFee*1e6)).Build())
			}

			instructions = append(instructions, solana.NewInstruction(
				solana.MemoProgramID,
				solana.AccountMetaSlice{
					solana.NewAccountMeta(TestAccount.PublicKey(), false, true),
				},
				[]byte(fmt.Sprintf("memobench: Test %d [%s]", id, TestID)),
			))

			tx, err := solana.NewTransaction(
				instructions,
				recent.Value.Blockhash,
				solana.TransactionPayer(TestAccount.PublicKey()),
			)
			if err != nil {
				log.Fatalf("error creating new transaction: %v", err)
			}

			_, err = tx.Sign(
				func(key solana.PublicKey) *solana.PrivateKey {
					if TestAccount.PublicKey().Equals(key) {
						return TestAccount
					}
					return nil
				},
			)
			if err != nil {
				log.Fatalf("error signing new transaction: %v", err)
			}

			// sleep until the next xx:xx:10s; then start spamming the transactions
			startTime := time.Now().Truncate(5 * time.Second).Add(10 * time.Second)
			sleepTime := time.Until(startTime)
			log.Info("Thread sleeping until starting spam", "thread", id, "delay", sleepTime.Truncate(time.Millisecond))
			time.Sleep(sleepTime)

			t0 := time.Now()
			if err := Limiter.Wait(context.TODO()); err != nil {
				log.Error(err.Error())
				return
			}

			// log if the thread had to throttle to keep under the rate limit
			throttleTime := time.Since(t0).Truncate(time.Millisecond)
			if throttleTime > 0 {
				log.Info("Thread throttled to respect rate-limit, Sending now", "thread", id, "delay", throttleTime)
			}

			log.Infof("Sending Tx [%s]", tx.Signatures[0])

			sig, err := sendClient.SendTransactionWithOpts(
				context.TODO(),
				tx,
				rpc.TransactionOpts{
					Encoding:      solana.EncodingBase64,
					SkipPreflight: true,
					MaxRetries:    &GlobalConfig.NodeRetries,
				},
			)
			if err != nil {
				if val, ok := err.(*jsonrpc.RPCError); ok {
					log.Errorf("Error sending tx: Received RPC error: %s", val.Message)
					return
				}

				log.Errorf("Error sending tx: %v", err)
				return
			}

			// save the tx send time for later comparison
			mu.Lock()
			TxTimes[sig] = time.Now()
			SentTransactions += 1
			mu.Unlock()
		}(i + 1)
	}
}

func main() {
	fmt.Println("                                                                                   ")
	fmt.Println(" ███╗   ███╗███████╗███╗   ███╗ ██████╗ ██████╗ ███████╗███╗   ██╗ ██████╗██╗  ██╗ ")
	fmt.Println(" ████╗ ████║██╔════╝████╗ ████║██╔═══██╗██╔══██╗██╔════╝████╗  ██║██╔════╝██║  ██║ ")
	fmt.Println(" ██╔████╔██║█████╗  ██╔████╔██║██║   ██║██████╔╝█████╗  ██╔██╗ ██║██║     ███████║ ")
	fmt.Println(" ██║╚██╔╝██║██╔══╝  ██║╚██╔╝██║██║   ██║██╔══██╗██╔══╝  ██║╚██╗██║██║     ██╔══██║ ")
	fmt.Println(" ██║ ╚═╝ ██║███████╗██║ ╚═╝ ██║╚██████╔╝██████╔╝███████╗██║ ╚████║╚██████╗██║  ██║ ")
	fmt.Println(" ╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝ ")
	fmt.Println("                                                                                   ")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		fmt.Println()
		log.Info("CTRL+C detected, Force stopping the test")
		fmt.Println()

		WsListener.Stop()
	}()

	// generate the test id
	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	TestID = hex.EncodeToString(randomBytes)

	// set up logger
	SetupLogger()

	// read the config file
	GlobalConfig = ReadConfig()

	// verify the private key is valid
	VerifyPrivateKey(GlobalConfig.PrivateKey)

	// set the rate limit
	Limiter.SetLimit(rate.Limit(GlobalConfig.RateLimit))
	Limiter.SetBurst(int(GlobalConfig.RateLimit))

	fmt.Printf("Starting Test ID    : %s\n", TestID)
	fmt.Printf("Test Account        : %s\n", TestAccount.PublicKey())
	fmt.Printf("RPC URL             : %s\n", GlobalConfig.RpcUrl)
	fmt.Printf("WS URL              : %s\n", GlobalConfig.GetWsUrl())
	fmt.Printf("RPC Send URL        : %s\n", GlobalConfig.GetSendUrl())
	fmt.Printf("Transaction Count   : %d\n", GlobalConfig.TxCount)
	fmt.Printf("Priority Fee        : %f Lamports (%.9f SOL)\n", GlobalConfig.PrioFee, (GlobalConfig.PrioFee*25000.0)/1e9)
	fmt.Printf("Node Retries        : %d\n", GlobalConfig.NodeRetries)
	fmt.Println()

	// start the websocket listener
	wg.Add(1)
	WsListener = new(WebsocketListener)
	go WsListener.Start()
	wg.Wait()

	fmt.Println()
	fmt.Printf("Finished Test ID    : %s\n", TestID)
	fmt.Printf("RPC URL             : %s\n", GlobalConfig.RpcUrl)
	fmt.Printf("WS URL              : %s\n", GlobalConfig.GetWsUrl())
	fmt.Printf("RPC Send URL        : %s\n", GlobalConfig.GetSendUrl())
	fmt.Printf("Transaction Count   : %d\n", GlobalConfig.TxCount)
	fmt.Printf("Priority Fee        : %f Lamports (%.9f SOL)\n", GlobalConfig.PrioFee, (GlobalConfig.PrioFee*25000.0)/1e9)
	fmt.Printf("Node Retries        : %d\n", GlobalConfig.NodeRetries)
	fmt.Printf("Transactions Landed : %d/%d (%.1f%%)\n", ProcessedTransactions, SentTransactions, float64(ProcessedTransactions)/float64(SentTransactions)*100.0)
	fmt.Println()
	fmt.Printf("Benchmark results saved to %s\n", LogFileName)
}
