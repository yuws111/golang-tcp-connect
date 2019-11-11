/*
 监听指定端口,并在握手成功后为客户端发送成功消息
 当有文件传输时能接收并保存，接收完成后发送接收完成消息
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"encoding/json"
)

const (
	CONNECT_CMD = 0x00 // 连接命令
	SEND_CMD = 0x01   // 发送命令
	DISCONNECT_CMD = 0xFF // 关闭命令
	TEST_CMD = 0x02    // 测试
	TRANSFER = 0xEE  // 持续传输
	TITLE    = 0x03   // 先传名称
	FILEEND  = 0x04   // 传输文件结束
	
)

type Packet struct {
	Type byte`json:type`
	Content []byte`json:content`
}

var (
	h bool

	p string
	
	file *os.File
)

func init() {
	flag.BoolVar(&h, "h", false, "帮助")        // -h 命令显示帮助
	flag.StringVar(&p, "p", "9090", "端口设置")  // 默认端口9090
	flag.Usage = usage
}

func main() {
	flag.Parse()
	// 输入-h 时打印出帮助
	if h {
		flag.Usage()
	}

	fmt.Printf("在主机上进行监听：127.0.0.1:%s\n", p)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+p)
	hasErrorExit(err)
	// 创建监听socket
	listener, err := net.ListenTCP("tcp", tcpAddr)
	hasErrorExit(err)
	defer listener.Close()
	for {
		// 等待客户端握手
		conn, err := listener.Accept()
		fmt.Printf("收到来自客户端%s的请求\n", conn.RemoteAddr().String())
		hasErrorExit(err)
		// 握手成功后在协程执行本次客户端操作
		go receiveClient(conn)
	}
}

func receiveClient(conn net.Conn) {
	defer fmt.Println("连接已关闭")
	defer conn.Close()
	state := 0x00
	length := uint16(0)
	crc16 := uint16(0)
	cursor := uint16(0)
	var receverBuf []byte
	bufferRead := bufio.NewReader(conn)
	// 发送端数据格式为 |0xff|0xff|size|len|data|crc32|0xff|0xfe
	for {
		recveByte, err := bufferRead.ReadByte()		
		if err != nil {
			if err == io.EOF {
				fmt.Printf("客户端%s 已经关闭了连接\n", conn.RemoteAddr().String())
			}
			return
		}
		switch state {
		case 0x00:
			if recveByte == 0xFF {
				state = 0x01
				receverBuf = nil
				length = 0
				crc16 = 0
			} else {
				state = 0x00
			}
			break
		case 0x01:
			if recveByte == 0xFF {
				state = 0x02
			} else {
				state = 0x00
			}
			break
		case 0x02:
			length += uint16(recveByte) * 256
			state = 0x03
			break
		case 0x03:
			length += uint16(recveByte)
			receverBuf = make([]byte, length)
			cursor = 0
			state = 0x04
			break
		case 0x04: // 读取数据的内容
			receverBuf[cursor] = recveByte
			cursor++
			if cursor == length {
				state = 0x05
			}
			break
		case 0x05:
			crc16 += uint16(recveByte) * 256
			state = 0x06
			break
		case 0x06:
			crc16 += uint16(recveByte)
			state = 0x07
			break
		case 0x07:
			if recveByte == 0xFF {
				state = 0x08
			} else {
				state = 0x00
			}
			break
		case 0x08:
			if recveByte == 0xFE {
				if crc32.ChecksumIEEE(receverBuf) >> 16 & 0xFFFF == uint32(crc16) {					
					var packet Packet
					json.Unmarshal(receverBuf,&packet)	
					fmt.Println("packet.type:",packet.Type)
					if packet.Type == DISCONNECT_CMD {
						fmt.Println("运行命令:",packet.Type)
						return
					}else if packet.Type == SEND_CMD {
						fmt.Println("准备接收数据....")
						conn.Write([]byte("title"))
					}else if packet.Type == TITLE {
						conn.Write([]byte("ok"))						
						createFile(string(packet.Content))
					}else if packet.Type == FILEEND {
						closeFile()
					}else if packet.Type == TRANSFER{						
						receiveFile(packet.Content)
					}else if packet.Type == CONNECT_CMD {
						fmt.Println("初次创建连接,准备接收数据")
					}else if packet.Type == TEST_CMD {
						back := "测试" + conn.RemoteAddr().String() + " connect to " + conn.LocalAddr().String() + " :" + p + "成功"
						conn.Write([]byte(back))
						return
					}
				}
				state = 0x00
			}
		}
	}	
}

func closeFile(){
	fmt.Println("文件传输完成，关闭")
	if file != nil {
		file.Close()
		file = nil
	}
}

func createFile(filename string) error {
	fmt.Println("正在创建文件",filename)
	file = nil
	var err error
	file, err = os.Create(filename)
	if err != nil {
		fmt.Println("创建文件出错",err)
		return err
	}
	return nil
}


func receiveFile(content []byte) {
	/**
	  读取文件名，向文件发送者返回OK
	*/
	if file == nil {
		fmt.Println("文件不存在")
		return
	}
	fmt.Println("正在接收文件，长度为:",len(content))
	file.Write(content)
}

func usage() {
	fmt.Fprintf(os.Stderr, `server version:server/1.0.0
使用方法: server [-h] [-p port]
选项： 
`)
	flag.PrintDefaults()
}

func hasErrorExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误:%s", err.Error())
		os.Exit(1)
	}
}
