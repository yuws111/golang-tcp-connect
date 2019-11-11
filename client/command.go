package main

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/goconfig"
	"io"
	"io/ioutil"
	"net"
	"os"
)

type Command struct {
	name string
	description string
}

const (
	CONNECT_CMD    = 0x00 // 连接命令,发送到目标机的状态
	SEND_CMD       = 0x01 // 发送命令
	DISCONNECT_CMD = 0xFF // 关闭命令
	TEST_CMD       = 0x02 // 测试
	TRANSFER       = 0xEE // 持续传输
	TITLE          = 0x03 // 先传名称
	FILEEND        = 0x04 // 传输文件结束
)

// 帮助命令提示
func (command Command) help() {
	command_list := []Command{
		Command{name:"test",description:"连接测试"},
		Command{name:"connect",description:"与目标机建立连接"},
		Command{name:"disconnect",description:"关闭当前连接"},
		Command{name:"send",description:"发送数据"},
		Command{name:"version",description:"版本"},
		Command{name:"target",description:"目标主机地址"},
		Command{name:"set",description:"设置ip,port,version"},
		Command{name:"quit",description:"退出"}}
	for i,v := range command_list {
		fmt.Printf("%d%-4v%-12s%s\n", i,RUNLINEHEAD,v.name,v.description)
	}
}

// 发送数据
func (command Command) send(inputs []string) {
	if conn == nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "请先与目标机建立连接")
		return
	}
	if len(inputs) > 1 {
		packet := Packet{Type: SEND_CMD, Content: []byte("")}
		agreePackets, err := json.Marshal(packet)
		if err != nil {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendError1"+err.Error())
			return
		}
		_, err = conn.Write(agreementData(agreePackets))
		if err != nil {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendError2"+err.Error())
			return
		}
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, inputs[1])
		fileInfo, err := os.Stat(inputs[1])
		if err != nil {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendError3"+err.Error())
			return
		}
		filename := fileInfo.Name()
		//发送文件名到接收端
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendError4"+err.Error())
			return
		}
		if string(buf[:n]) == "title" {
			packet := Packet{Type: TITLE, Content: []byte(filename)}
			agreePackets, err := json.Marshal(packet)
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "目标机正在准备接收文件")
			if err != nil {
				fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendError5"+err.Error())
				return
			}
			_, err = conn.Write(agreementData(agreePackets))
			if err != nil {
				fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendError6"+err.Error())
				return
			}
		}
		n, err = conn.Read(buf)
		if err != nil {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendError7"+err.Error())
			return
		}
		// 返回OK 文件已经创建，可以接收内容
		if string(buf[:n]) == "ok" {
			sendFile(conn, inputs[1])
		}
	} else {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "缺少filepath，命令send使用方法为:send filepath")
	}
	return
}

// 与目标机建立连接
func (command Command) connect() {
	if conn != nil {
		conn.Close()
	}
	tip, err := command.getTarget()
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "connectError1"+err.Error())
		return
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp4", tip)
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "connectError2"+err.Error())
		conn = nil
		return
	}
	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "connectError3"+err.Error())
		conn = nil
		return
	}
	fmt.Printf("%-4v%s\n", RUNLINEHEAD, "已经建立了连接，正在执行相关操作")
}

func (command Command) disconnect() {
	if conn == nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "还未建立连接")
		return
	}
	conn.Close()
	conn = nil
	fmt.Printf("%-4v%s\n", RUNLINEHEAD, "连接已经关闭")
}

func (command Command) getTarget() (string, error) {
	conf, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "getTargetError1"+err.Error())
		return "", err
	}
	section, err := conf.GetSection("target")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "getTargetError2"+err.Error())
		return "", err
	}
	return section["ip"] + ":" + section["port"], nil
}

// 测试与目标机连接情况
func (command Command) test() {

	tip, err := command.getTarget()
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "testError1"+err.Error())
		return
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp4", tip)
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "testError2"+err.Error())
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "testError3"+err.Error())
		return
	}
	packet := Packet{Type: TEST_CMD, Content: []byte("")}
	agreePackets, err := json.Marshal(packet)
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "testError4"+err.Error())
		return
	}
	_, err = conn.Write(agreementData(agreePackets))
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "testError5"+err.Error())
		return
	}
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "testError6"+err.Error())
		return
	}
	fmt.Printf("%-4v%v\n", RUNLINEHEAD, string(result))
	conn.Close()
	conn = nil
	return
}

// 查看版本号
func (command Command) version() error {
	conf, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "versionError1"+err.Error())
		return err
	}
	section, err := conf.GetSection("client")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "versionError2"+err.Error())
		return err
	}
	fmt.Printf("%-4v%-10s%-20s\n", RUNLINEHEAD, "version", section["version"])
	return nil
}

// 设置命令
func (command Command) set(commands []string) {
	if len(commands) > 2 {
		switch commands[1] {
		case "ip":
			setip(commands[2])
		case "port":
			setport(commands[2])
		case "version":
			setversion(commands[2])
		default:
			fmt.Println(RUNLINEHEAD, "命令格式错误")
		}
	} else {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "set命令使用方式 set ip addr | set port 90 | set version 1.0")
	}
}

// 获取目标主机
func (command Command) target() error {
	conf, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "targetError1"+err.Error())
		return err
	}
	section, err := conf.GetSection("target")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "targetError2"+err.Error())
		return err
	}
	fmt.Printf("%-4vIP:%-10s:%-8s\n", RUNLINEHEAD, section["ip"], section["port"])
	return nil
}

func (command Command) quit() {
	os.Exit(1)
}

// 设置ip
func setip(ip string) {
	// 检测ip地址是否合法
	conf, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "setIpError1"+err.Error())
		return
	}
	conf.SetValue("target", "ip", ip)
	err = goconfig.SaveConfigFile(conf, "conf.ini")
	if err != nil {
		fmt.Printf("err%-4v%s\n", RUNLINEHEAD, err.Error())
	} else {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "设置IP成功")
	}
}

// 设置端口号
func setport(p string) {
	// 检测端口是否合法
	conf, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "setPortError1"+err.Error())
		return
	}

	conf.SetValue("target", "port", p)
	err = goconfig.SaveConfigFile(conf, "conf.ini")
	if err != nil {
		fmt.Printf("err%-4v%s\n", RUNLINEHEAD, err.Error())
	} else {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "设置port成功")
	}
}

// 设置版本号
func setversion(v string) {
	conf, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "setversionError1"+err.Error())
		return
	}

	conf.SetValue("client", "version", v)
	err = goconfig.SaveConfigFile(conf, "conf.ini")
	if err != nil {
		fmt.Printf("err%-4v%s\n", RUNLINEHEAD, err.Error())
	} else {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "设置version成功")
	}
}

func sendFile(conn net.Conn, filepath string) {	
	//打开要传输的文件
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendFileError1"+err.Error())
		return
	}
	defer file.Close()
	buf := make([]byte, 4096)
	//循环读取文件内容，写入远程连接
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "文件读取完毕")
			packet := Packet{Type: FILEEND, Content: []byte("/n")}
			agreePackets, err := json.Marshal(packet)
			if err != nil {
				fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendFileError5"+err.Error())
				return
			}
			_, err = conn.Write(agreementData(agreePackets))
			if err != nil {
				fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendFileError6"+err.Error())
				return
			}			
			conn = nil
			return
		}
		if err != nil {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendFileError2"+err.Error())
			return
		}
		packet := Packet{Type: TRANSFER, Content: []byte(buf[:n])}
		agreePackets, err := json.Marshal(packet)
		if err != nil {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendFileError3"+err.Error())
			return
		}
		_, err = conn.Write(agreementData(agreePackets))
		if err != nil {
			fmt.Printf("%-4v%s\n", RUNLINEHEAD, "sendFileError4"+err.Error())
			return
		}
	}
}
