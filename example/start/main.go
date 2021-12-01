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
	const room int64 = 573893

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

	ifError := make(chan error)
	go func() {
		// 进入房间
		// room: room id(真实ID，短号需自行转换)
		// key: 用户标识，可留空
		// uid: 用户UID，可随机生成
		if err := l.Enter(ctx, room, "", 12345678); err != nil {
			log.Println("Error Encountered: ", err)
			log.Println("Room Disconnected")
			ifError <- err
			return
		}
	}()

	go rev(ctx, l)

	// 15s的演示
	after := time.After(15 * time.Second)
	select {
	case <-after:
		fmt.Println("I want to stop")
		// 关闭ws连接与相关协程
		stop()
		break
	case err := <-ifError:
		fmt.Println("I don't want to stop, but I encountered an error: ", err)
		break
	}
	// 五秒時間讓他關閉其他 goroutine
	<-time.After(5 * time.Second)
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
	switch msg.(type) {
	// 心跳回应直播间人气值
	case *live.MsgHeartbeatReply:
		log.Printf("hot: %d\n", msg.(*live.MsgHeartbeatReply).GetHot())
	// 弹幕消息
	case *live.MsgDanmaku:
		dm, err := msg.(*live.MsgDanmaku).Parse()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("弹幕: %s (%d:%s) 【%s】| %d\n", dm.Content, dm.MID, dm.Uname, dm.MedalName, dm.Time)
	// 礼物消息
	case *live.MsgSendGift:
		g, err := msg.(*live.MsgSendGift).Parse()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%s: %s %d个%s\n", g.Action, g.Uname, g.Num, g.GiftName)
	// 直播间粉丝数变化消息
	case *live.MsgFansUpdate:
		f, err := msg.(*live.MsgFansUpdate).Parse()
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
