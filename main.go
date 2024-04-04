package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	probing "github.com/prometheus-community/pro-bing"
)

type exitCode int

const (
	ExitCodeOK exitCode = iota
	ExitCodeErrorArgs
	ExitCodeErrorPing
)

type options struct {
	Count     int  `short:"c" long:"count" description:"Number of packets to send."`
	Privilege bool `short:"p" long:"privileged" description:"Execute in privileged mode."`
	Version   bool `short:"v" long:"version" description:"Print version."`
}

var (
	appName        = "nemuping"
	appVersion     = "0.0.1"
	appDescription = "Ping command with custom ASCII art."
	Usage          = "Usage: nemuping [options...] <host>"
)

func main() {
	code, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%v] %v\n", color.New(color.FgRed, color.Bold).Sprint("ERROR"), err)
	}
	os.Exit(int(code))
}

func run(cliArgs []string) (exitCode, error) {
	errorMessage := func(message string) error {
		return fmt.Errorf("%s\n\n%s", message, Usage)
	}
	var options options
	parser := flags.NewParser(&options, flags.Default)
	parser.Name = appName
	parser.Usage = Usage
	parser.ShortDescription = appDescription
	//アスキーアートの行数に合わせてこちらの値を調整すると、アスキーが全部出力された時点で終了するようになります。
	options.Count = 18

	args, err := parser.ParseArgs(cliArgs)
	if err != nil {
		if flags.WroteHelp(err) {
			return ExitCodeOK, nil
		}
		return ExitCodeErrorArgs, errorMessage("パーサエラー" + err.Error())
	}
	if options.Version {
		fmt.Printf("%s version %s\n", appName, appVersion)
		return ExitCodeOK, nil
	}

	if len(args) == 0 {
		return ExitCodeErrorArgs, errors.New("引数エラー")
	}

	if len(args) > 1 {
		return ExitCodeErrorArgs, errors.New("引数エラー")
	}

	asciiArt, err := readASCIIArtFromEnv(".env")
	if err != nil {
		return ExitCodeErrorArgs, err
	}

	pinger, err := initPinger(args[0], options, asciiArt)
	if err != nil {
		return ExitCodeOK, fmt.Errorf("Pingの初期化エラー: %w", err)
	}
	if err := pinger.Run(); err != nil {
		return ExitCodeErrorPing, fmt.Errorf("Pingの実行エラー: %w", err)
	}
	return ExitCodeOK, nil
}

func initPinger(host string, options options, asciiArt []string) (*probing.Pinger, error) {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return nil, fmt.Errorf("pingの初期化エラー: %s", err)
	}

	pinger.Count = options.Count

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		pinger.Stop()
	}()

	fmt.Printf("%s (%s) %d Ctrl+Cにより処理をボイコットできます。\n",
		pinger.Addr(),
		pinger.IPAddr(),
	)

	pinger.OnRecv = pingerOnRecvColor(asciiArt)
	pinger.OnFinish = pingerOnFinishSummary

	if options.Privilege || runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	pinger.OnFinish = func(stats *probing.Statistics) {
		pinger.Stop()
		pingerOnFinishSummary(stats)
	}

	return pinger, nil
}

func pingerOnRecvColor(asciiArt []string) func(pkt *probing.Packet) {
	return func(pkt *probing.Packet) {
		fmt.Printf("%s %d bytes from %s: icmp_seq=%d ttl=%d time=%v\n",
			asciiArt[pkt.Seq%len(asciiArt)],
			pkt.Nbytes,
			color.New(color.FgGreen, color.Bold).Sprint(pkt.IPAddr),
			color.New(color.FgCyan, color.Bold).Sprint(pkt.Seq),
			color.New(color.FgRed, color.Bold).Sprint(pkt.TTL),
			color.New(color.FgBlue, color.Bold).Sprint(pkt.Nbytes),
		)
	}
}

func pingerOnFinishSummary(stats *probing.Statistics) {
	color.New(color.FgGreen, color.Bold).Printf("\n--- %s ねむパケット通信状況 ---\n", stats.Addr)

	fmt.Fprintf(color.Output, "%d パケット転送, %d パケット受信, %v%% パケロス\n",
		color.New(color.FgGreen, color.Bold).Sprint(stats.PacketsSent),
		color.New(color.FgCyan, color.Bold).Sprint(stats.PacketsRecv),
		color.New(color.FgBlue, color.Bold).Sprint(stats.PacketLoss),
	)

	fmt.Fprintf(color.Output, "ラウンドトリップの最小/平均/最大/標準偏差 = %v/%v/%v/%v\n",
		color.New(color.FgGreen, color.Bold).Sprint(stats.MinRtt),
		color.New(color.FgCyan, color.Bold).Sprint(stats.AvgRtt),
		color.New(color.FgBlue, color.Bold).Sprint(stats.MaxRtt),
		color.New(color.FgRed, color.Bold).Sprint(stats.StdDevRtt),
	)
}

func readASCIIArtFromEnv(filename string) ([]string, error) {
	var asciiArt []string

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		asciiArt = append(asciiArt, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return asciiArt, nil
}
