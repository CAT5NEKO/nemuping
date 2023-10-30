package main

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	probing "github.com/prometheus-community/pro-bing"
	"os"
	"os/signal"
	"runtime"
	"strings"
)

// ねむだるの部分を変更すれば、好きなアイコンが出力できるようになる。
// 　でも、ねむダルに併せて色を2色に絞っちゃってるから、色が変わらない。なんか色々足してください。
var nemudaru = []string{

	`* ...................................................`,
	`* ...................................................`,
	`* ...................................::+#%##*-.......`,
	`* ..........-=***+=::................=%%%: =#%=......`,
	`* ........=%%%#:.=##-...............-%%%#  .##%:.....`,
	`* .......*%%%%+  .##%-..............*%%%%:-*%#%=.....`,
	`* ......=%%%%%*..=%%##..............#%%%%#%%%+#+.....`,
	`* ......#%#%%%#*%%%#+%..............**#%%%%%*=%-.....`,
	`* ......+%*#%%%%%%%*+%..............:#+####*+**......`,
	`* .......##+*#####*+#=...............-*++=-=#*.......`,
	`* ........*#++++==+#=..................=+**+:........`,
	`* .........:+*#***=:......................... .......`,
	`* ...............................:-+-................`,
	`* .........  ..........BBBBBBBBBBBCC.................`,
	`* ......................CBBBBBBB.BBN.................`,
	`* ........................CBBBBBBC...................`,
	`* ...................................................`,
	`* ...................................................`,
}

var appName = "nemudaru"
var appVersion = "0.0.1"
var appDescription = "ping command with nemudaru"
var Usage = "Usage: nemudaru [options...] <host>"

type exitCode int

const (
	ExitCodeOK exitCode = iota
	ExitCodeErrorArgs
	ExitCodeErrorPing
)

// 現段階ではバージョニングが上手く再現できず、今後修正する。
type options struct {
	Count     int  `short:"c" long:"count" description:"パケットの送信回数を指定します。" default:"4"`
	Privilege bool `short:"p" long:"privileged" description:"特権モードで実行します。"`
	Version   bool `short:"v" long:"version" description:"バージョンを表示します。"`
}

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
	//アスキーアートに合わせてデフォルトの出力数弄れる。本家ではデフォルトを開放しても止まることなく挙動してたから修正必要かも。
	options.Count = 17

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
	//　!0処理では一概にして因数不備として処理するエラーが出たので対応。
	if len(args) == 0 {
		return ExitCodeErrorArgs, errors.New("引数エラー")
	}

	if len(args) > 1 {
		return ExitCodeErrorArgs, errors.New("引数エラー")
	}

	pinger, err := initPinger(args[0], options)
	if err != nil {
		return ExitCodeOK, fmt.Errorf("Pingの初期化エラー: %w", err)
	}
	if err := pinger.Run(); err != nil {
		return ExitCodeErrorPing, fmt.Errorf("Pingの実行エラー: %w", err)
	}
	return ExitCodeOK, nil
}

func initPinger(host string, options options) (*probing.Pinger, error) {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return nil, fmt.Errorf("pingの初期化エラー: %s", err)
	}

	pinger.Count = options.Count

	// キャンセル処理
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

	pinger.OnRecv = pingerOnRecvColor
	pinger.OnFinish = pingerOnFinishSummary

	if options.Privilege || runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	return pinger, nil
}

func pingerOnRecvColor(pkt *probing.Packet) {
	fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%v\n",
		compileASCIICode(pkt.Seq),
		color.New(color.FgGreen, color.Bold).Sprint(pkt.IPAddr),
		color.New(color.FgCyan, color.Bold).Sprint(pkt.Seq),
		color.New(color.FgRed, color.Bold).Sprint(pkt.TTL),
		color.New(color.FgGreen, color.Bold).Sprint(pkt.IPAddr),
		color.New(color.FgBlue, color.Bold).Sprint(pkt.Nbytes),
	)
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

func compileASCIICode(idx int) string {
	if len(nemudaru) <= idx {
		idx %= len(nemudaru)
	}

	line := nemudaru[idx]

	line = colorize(line, ".", color.New(color.FgWhite, color.Bold))
	line = colorize(line, "-,+,*,+,=,:,#", color.New(color.FgGreen, color.Bold))
	line = colorize(line, "B,C,N", color.New(color.FgRed, color.Bold))

	return line
}

// 上の文字列はruneで扱うと死んでしまうので文字列縛りで処理
func colorize(text string, target string, color *color.Color) string {
	return strings.ReplaceAll(
		text,
		string(target),
		color.Sprint("#"),
	)
}
