/*
1，选择运行进入界面
2，实现帮助查询，当输入help 或者 ？ 时返回命令列表
3，set 设置目标ip和端口ip:port
4，test 测试连接是否成功
5，connect 建立连接
6，send 发送文件
*/

package main

import (
	"flag"
	"fmt"
	"github.com/Unknwon/goconfig"
	"os"
)

var (
	h bool

	v, V bool

	i bool
	p bool
)

func init() {
	flag.BoolVar(&h, "h", false, "帮助")

	flag.BoolVar(&v, "v", false, "显示版本号")
	flag.BoolVar(&V, "V", false, "显示版本号") // 显示客户端当前版本

	flag.BoolVar(&i, "i", false, "目标ip地址") // 显示目标ip地址
	flag.BoolVar(&p, "p", false, "目标访问端口") // 显示目标端口

	flag.Usage = usage
}

func main() {
	flag.Parse()

	if h {
		flag.Usage()
	}

	conf, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		panic(err)
	}
	section, err := conf.GetSection("target")
	if err != nil {
		panic(err)
	}
	if i {
		fmt.Printf("目标机IP地址：%v\n", section["ip"])
	}
	if p {
		fmt.Printf("开放端口：%v\n", section["port"])
	}

	Run_help()

}

func usage() {
	fmt.Fprintf(os.Stderr, `send version:send/1.0.0
使用方法: send [-hvV] [-i ip] [-p port]
选项： 
`)
	flag.PrintDefaults()
}
