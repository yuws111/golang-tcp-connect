package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	RUNLINEHEAD = "::"
	COMMANDHEAD = ":>"
)

var conn net.Conn = nil

/*
命令使用帮助
? 显示所有选项
send 发送文件
test 连接测试
connect 建立连接
*/

func Run_help() {
	fmt.Print(COMMANDHEAD)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
	 	inputs := strings.Split(strings.TrimLeft(input.Text(), " "), " ")
		command  := Command {
			name:strings.Trim(inputs[0]," "),
			description:"",
		}
		switch command.name {
		case "test":
			command.test()
		case "connect":
			command.connect()
		case "disconnect":
			command.disconnect()
		case "send":
			command.send(inputs)
		case "version":
			command.version()
		case "set":
			command.set(inputs)
		case "target":
			command.target()
		case "?":
			command.help()
		case "help":
			command.help()
		case "quit":
			command.quit()
		default:
			fmt.Println(RUNLINEHEAD, "无此命令,请输入help或者?查看使用方法")
		}
		fmt.Print(COMMANDHEAD)
	}
}

