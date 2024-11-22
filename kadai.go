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

		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}
		if input == "adios" {
			break
		}

		command1, command2, separator := parseCommands(input)

		// パイプ処理
		if strings.Contains(input, "|") {
			err := handlePipe(command1, command2)
			if err != nil {
				fmt.Printf("Pipe error: %v\n", err)
			}
			num++
			continue
		}

		// リダイレクト処理
		cmd1, inputFile1, outputFile1 := handleRedirection(command1)
		cmd2, inputFile2, outputFile2 := handleRedirection(command2)

		// 1つ目のコマンドを実行
		exitStatus, err := executeCommand(cmd1, inputFile1, outputFile1)
		if err != nil {
			fmt.Printf("%s: %v\n", cmd1[0], err)
		}

		// 2つ目のコマンドの実行条件
		shouldExecuteSecond := false
		if separator == "&&" && exitStatus == 0 {
			shouldExecuteSecond = true
		} else if separator == "||" && exitStatus != 0 {
			shouldExecuteSecond = true
		}

		if shouldExecuteSecond && len(cmd2) > 0 {
			_, err := executeCommand(cmd2, inputFile2, outputFile2)
			if err != nil {
				fmt.Printf("%s: %v\n", cmd2[0], err)
			}
		}

		num++
	}
}

// コマンド分割
func parseCommands(input string) ([]string, []string, string) {
	if strings.Contains(input, "&&") {
		parts := strings.Split(input, "&&")
		return strings.Fields(strings.TrimSpace(parts[0])), strings.Fields(strings.TrimSpace(parts[1])), "&&"
	} else if strings.Contains(input, "||") {
		parts := strings.Split(input, "||")
		return strings.Fields(strings.TrimSpace(parts[0])), strings.Fields(strings.TrimSpace(parts[1])), "||"
	}
	return strings.Fields(input), nil, ""
}

// リダイレクト処理
func handleRedirection(command []string) ([]string, string, string) {
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
	pipeIn, pipeOut, err := os.Pipe()
	if err != nil {
		return err
	}

	cmd1 := exec.Command(command1[0], command1[1:]...)
	cmd1.Stdout = pipeOut
	cmd1.Stderr = os.Stderr

	cmd2 := exec.Command(command2[0], command2[1:]...)
	cmd2.Stdin = pipeIn
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr

	if err := cmd1.Start(); err != nil {
		return fmt.Errorf("command1 failed: %v", err)
	}
	if err := cmd2.Start(); err != nil {
		return fmt.Errorf("command2 failed: %v", err)
	}

	pipeOut.Close()
	pipeIn.Close()

	if err := cmd1.Wait(); err != nil {
		return fmt.Errorf("command1 wait failed: %v", err)
	}
	if err := cmd2.Wait(); err != nil {
		return fmt.Errorf("command2 wait failed: %v", err)
	}
	return nil
}

// コマンド実行
func executeCommand(command []string, inputFile, outputFile string) (int, error) {
	cmd := exec.Command(command[0], command[1:]...)

	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			return 1, err
		}
		defer file.Close()
		cmd.Stdin = file
	}
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return 1, err
		}
		defer file.Close()
		cmd.Stdout = file
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return 1, err
	}
	return 0, nil
}
