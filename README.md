<img src="./img/logo.png" alt="logo" width="130" height="130" align="left" />

<h1>BiliGO-LIVE</h1>

> BiliBili Live WebSocket Protocol SDK in Golang

<br/>

![](https://img.shields.io/github/go-mod/go-version/iyear/biligo-live?style=flat-square)
![](https://img.shields.io/badge/license-GPL-lightgrey.svg?style=flat-square)
![](https://img.shields.io/github/v/release/iyear/biligo-live?color=red&style=flat-square)
![](https://img.shields.io/github/last-commit/iyear/biligo-live?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/iyear/biligo-live.svg)](https://pkg.go.dev/github.com/iyear/biligo-live)

## 简介

**v0版本不保证对外函数、结构的不变性，请勿大规模用于生产环境**

哔哩哔哩直播 `WebSocket` 协议的 `Golang` 封装

### 特性
- 良好的设计，自定义程度高
- 代码、结构体注释完善，开箱即用
- 功能简单，封装程度高
### 说明

- 该项目永远不会编写直接涉及滥用的接口
- 该项目仅供学习，请勿用于商业用途。任何使用该项目造成的后果由开发者自行承担

### 参考

## 快速开始
### 安装

```shell
go get -u github.com/iyear/biligo-live
```

```go
import "github.com/iyear/biligo-live"
```

### 使用
<details>
<summary>查看代码</summary>

```go
package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/iyear/biligo-live"
	"log"
	"time"
)

// 同 README.md 的快速开始

func main() {
	const room int64 = 48743

	// 获取一个Live实例
	// debug: debug模式，输出一些额外的信息
	// heartbeat: 心跳包发送间隔。不发送心跳包，70 秒之后会断开连接，通常每 30 秒发送 1 次
	// cache: Rev channel 的缓存
	// recover: panic recover后的操作函数
	l := live.NewLive(true, 30*time.Second, 0, func(err error) {
		log.Println("panic:", err)
		// do something...
	})

	// 连接ws服务器
	// dialer: ws dialer
	// host: bilibili live ws host
	if err := l.Conn(websocket.DefaultDialer, live.WsDefaultHost); err != nil {
		log.Fatal(err)
		return
	}

	ctx, stop := context.WithCancel(context.Background())

	go func() {
		// 进入房间
		// room: room id(真实ID，短号需自行转换)
		// key: 用户标识，可留空
		// uid: 用户UID，可随机生成
		if err := l.Enter(ctx, room, "", 12345678); err != nil {
			log.Fatal(err)
			return
		}
	}()

	go rev(ctx, l)

	// 15s的演示
	after := time.NewTimer(15 * time.Second)
	defer after.Stop()
	<-after.C
	fmt.Println("I want to stop")
	// 关闭ws连接与相关协程
	stop()
	// 为了使安全退出效果可见，进行阻塞，真实场景中可以移除
	select {}
}
func rev(ctx context.Context, l *live.Live) {
	for {
		select {
		case tp := <-l.Rev:
			if tp.Error != nil {
				// do something...
				log.Println(tp.Error)
				continue
			}
			handle(tp.Msg)
		case <-ctx.Done():
			log.Println("rev func stopped")
			return
		}
	}
}
func handle(msg live.Msg) {
	// 使用 msg.(type) 进行事件跳转和处理，常见事件基本都完成了解析(Parse)功能，不常见的功能有一些实在太难抓取
	// 更多注释和说明等待添加
	switch msg := msg.(type) {
	// 心跳回应直播间人气值
	case *live.MsgHeartbeatReply:
		log.Printf("hot: %d\n", msg.GetHot())
	// 弹幕消息  
	case *live.MsgDanmaku:
		dm, err := msg.Parse()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("弹幕: %s (%d:%s) 【%s】| %d\n", dm.Content, dm.MID, dm.Uname, dm.MedalName, dm.Time)
	// 礼物消息 
	case *live.MsgSendGift:
		g, err := msg.Parse()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%s: %s %d个%s\n", g.Action, g.Uname, g.Num, g.GiftName)
	// 直播间粉丝数变化消息 
	case *live.MsgFansUpdate:
		f, err := msg.Parse()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("room: %d,fans: %d,fansClub: %d\n", f.RoomID, f.Fans, f.FansClub)
	// case:......

	// General 表示live未实现的CMD命令，请自行处理raw数据。也可以提issue更新这个CMD
	case *live.MsgGeneral:
		fmt.Println("unknown msg type|raw:", string(msg.Raw()))
	}
}

```
</details>

## LICENSE

GPLv3