package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	num := 0

	for {
		fmt.Printf("./myshell[%02d]> ", num)

		// コマンド入力
		input, err := reader.ReadString('\n')
		if err != nil { // Ctrl+D時
			fmt.Println("シェルを終了します。")
			break
		}
		input = strings.TrimSpace(input)

		// 空の入力は無視
		if input == "" {
			continue
		}

		// "adios" で終了
		if input == "adios" {
			fmt.Println("シェルを終了します。")
			break
		}

		// コマンド解析と分割
		command1, command2, separator := parseCommands(input)

		// パイプがある場合、パイプ処理を優先
		if separator == "|" {
			err := handlePipe(command1, command2)
			if err != nil {
				fmt.Printf("パイプエラー: %v\n", err)
			}
			continue
		}

		// リダイレクトを解析
		command1Args, inputFile1, outputFile1 := handleRedirection(command1)
		command2Args, inputFile2, outputFile2 := handleRedirection(command2)

		// 1つ目のコマンドを実行
		exitStatus, err := executeCommand(command1Args, inputFile1, outputFile1)
		if err != nil {
			fmt.Printf("%s: 実行エラー: %v\n", command1Args[0], err)
		}

		// 2つ目のコマンドを実行する条件
		shouldExecuteSecond := false
		if separator == "&&" && exitStatus == 0 {
			shouldExecuteSecond = true
		} else if separator == "||" && exitStatus != 0 {
			shouldExecuteSecond = true
		}

		// 2つ目のコマンド実行
		if shouldExecuteSecond && len(command2Args) > 0 {
			_, err = executeCommand(command2Args, inputFile2, outputFile2)
			if err != nil {
				fmt.Printf("%s: 実行エラー: %v\n", command2Args[0], err)
			}
		}

		num++
	}
}

// コマンド解析（リダイレクトやパイプ記号を考慮）
func parseCommands(input string) ([]string, []string, string) {
	if strings.Contains(input, "|") {
		parts := strings.Split(input, "|")
		return strings.Fields(strings.TrimSpace(parts[0])), strings.Fields(strings.TrimSpace(parts[1])), "|"
	} else if strings.Contains(input, "&&") {
		parts := strings.Split(input, "&&")
		return strings.Fields(strings.TrimSpace(parts[0])), strings.Fields(strings.TrimSpace(parts[1])), "&&"
	} else if strings.Contains(input, "||") {
		parts := strings.Split(input, "||")
		return strings.Fields(strings.TrimSpace(parts[0])), strings.Fields(strings.TrimSpace(parts[1])), "||"
	}
	return strings.Fields(input), nil, ""
}

// リダイレクト解析
func handleRedirection(command []string) ([]string, string, string) {
	if len(command) == 0 {
		return nil, "", ""
	}

	var inputFile, outputFile string
	var filteredCmd []string

	for i := 0; i < len(command); i++ {
		if command[i] == ">" && i+1 < len(command) {
			outputFile = command[i+1]
			i++
		} else if command[i] == "<" && i+1 < len(command) {
			inputFile = command[i+1]
			i++
		} else {
			filteredCmd = append(filteredCmd, command[i])
		}
	}

	return filteredCmd, inputFile, outputFile
}

// パイプ処理
func handlePipe(command1, command2 []string) error {
	if len(command1) == 0 || len(command2) == 0 {
		return fmt.Errorf("無効なパイプコマンド")
	}

	// パイプの作成
	pipeIn, pipeOut, err := os.Pipe()
	if err != nil {
		return err
	}

	// コマンド1の実行準備
	cmd1 := exec.Command(command1[0], command1[1:]...)
	cmd1.Stdout = pipeOut // 標準出力をパイプの出力に接続
	cmd1.Stderr = os.Stderr

	// コマンド2の実行準備
	cmd2 := exec.Command(command2[0], command2[1:]...)
	cmd2.Stdin = pipeIn   // 標準入力をパイプの入力に接続
	cmd2.Stdout = os.Stdout // 標準出力をシェルの標準出力に接続
	cmd2.Stderr = os.Stderr

	// 必要に応じてパイプを閉じる
	pipeOut.Close() // cmd1 の出力先として使用するので、親プロセスではクローズ
	pipeIn.Close()  // cmd2 の入力元として使用するので、親プロセスではクローズ

	// 両コマンドの実行
	if err := cmd1.Start(); err != nil {
		return fmt.Errorf("コマンド1の実行エラー: %v", err)
	}
	if err := cmd2.Start(); err != nil {
		return fmt.Errorf("コマンド2の実行エラー: %v", err)
	}

	// コマンドの終了待機
	if err := cmd1.Wait(); err != nil {
		return fmt.Errorf("コマンド1の終了待機エラー: %v", err)
	}
	if err := cmd2.Wait(); err != nil {
		return fmt.Errorf("コマンド2の終了待機エラー: %v", err)
	}

	return nil
}

// コマンド実行（リダイレクト対応）
func executeCommand(command []string, inputFile, outputFile string) (int, error) {
	if len(command) == 0 {
		return 0, nil
	}

	cmd := exec.Command(command[0], command[1:]...)

	// リダイレクトの処理
	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			return 1, fmt.Errorf("入力ファイルを開けません: %v", err)
		}
		defer file.Close()
		cmd.Stdin = file
	}
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return 1, fmt.Errorf("出力ファイルを作成できません: %v", err)
		}
		defer file.Close()
		cmd.Stdout = file
	} else {
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = os.Stderr

	// コマンド実行
	err := cmd.Run()
	if err != nil {
		return 1, err
	}

	return 0, nil
}
