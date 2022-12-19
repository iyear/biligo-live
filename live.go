package live

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type Live struct {
	ws      *websocket.Conn
	debug   bool
	logger  *log.Logger
	entered chan struct{}
	hb      time.Duration
	recover func(error)
	Rev     chan *Transport
}

// NewLive 创建一个新的直播连接
func NewLive(debug bool, heartbeat time.Duration, cache int, recover func(error)) *Live {
	return &Live{
		ws:      nil,
		debug:   debug,
		logger:  log.New(os.Stdout, "Live ", log.LstdFlags|log.Lshortfile),
		hb:      heartbeat,
		entered: make(chan struct{}),
		recover: recover,
		Rev:     make(chan *Transport, cache),
	}
}

// Conn ws连接bilibili弹幕服务器
func (l *Live) Conn(dialer *websocket.Dialer, host string) error {
	return l.ConnWithHeader(dialer, host, nil)
}

// ConnWithHeader ws连接bilibili弹幕服务器 (带header)
func (l *Live) ConnWithHeader(dialer *websocket.Dialer, host string, header http.Header) error {
	w, _, err := dialer.Dial(host, header)
	if err != nil {
		return err
	}
	l.ws = w
	return nil
}

// TODO 错误重试机制

// Enter 进入房间。 Conn 后五秒内必须进入房间，否则服务器主动断开连接
func (l *Live) Enter(ctx context.Context, room int64, key string, uid int64) error {
	enter := map[string]interface{}{
		"platform": "web",
		"protover": 2,
		"roomid":   room,
		"uid":      uid,
		"type":     2,
		"key":      key,
	}
	body, err := json.Marshal(enter)
	if err != nil {
		return err
	}
	if err = l.ws.WriteMessage(websocket.BinaryMessage, encode(wsVerPlain, wsOpEnterRoom, body)); err != nil {
		return err
	}

	hbCtx, hbCancel := context.WithCancel(ctx)
	revCtx, revCancel := context.WithCancel(ctx)
	ifError := make(chan error, 1)
	go l.revWithError(revCtx, ifError)

	go func() {
		select {
		case <-l.entered:
		case <-hbCtx.Done():
			return
		}
		l.heartbeat(hbCtx, l.hb)
	}()

	defer func() {
		hbCancel()
		revCancel()
		err = l.ws.Close()
	}()

	select {
	// 外部停止ws
	case <-ctx.Done():
		l.info("websocket conn stopped")
		break
	// 內部接收 Websocket 訊息錯誤
	case err = <-ifError:
		l.error("websocket conn stopped with an error: %s", err)
		break
	}

	return err
}
func (l *Live) report() {
	if r := recover(); r != nil {
		var e error
		switch r.(type) {
		case error:
			e = r.(error)
		case string:
			e = fmt.Errorf("%s", r.(string))
		}

		if l.recover != nil {
			l.recover(e)
		}
		l.error("panic: %s", e)
	}
}
func (l *Live) heartbeat(ctx context.Context, t time.Duration) {
	hb := func(live *Live) {
		err := live.ws.WriteMessage(websocket.BinaryMessage, encode(wsVerPlain, wsOpHeartbeat, nil))
		if err != nil {
			live.push(ctx, nil, fmt.Errorf("failed to send hearbeat: %s", err))
		}
	}

	// 开头先执行一次
	hb(l)
	ticker := time.NewTicker(t)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			l.info("heartbeat stopped")
			return
		case <-ticker.C:
			hb(l)
		}
	}
}

// revWithError 接收訊息並捕捉錯誤
func (l *Live) revWithError(ctx context.Context, ifError chan<- error) {
	msgCtx, msgCancel := context.WithCancel(ctx)
	defer l.info("receiving stopped")
	defer msgCancel()
	defer close(ifError)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if t, msg, err := l.ws.ReadMessage(); t == websocket.BinaryMessage && err == nil && len(msg) > 16 {
				go l.handle(msgCtx, msg)
			} else if err != nil {
				ifError <- err
				return
			}
		}
	}
}

func (l *Live) handle(ctx context.Context, b []byte) {
	defer l.report()
	ver, op, body := decode(b)
	switch op {
	case wsOpEnterRoomSuccess:
		l.info("enter room success: %s", string(body))
		l.entered <- struct{}{}
	case wsOpHeartbeatReply:
		l.info("heartbeat reply: %d", binary.BigEndian.Uint32(body))
		l.push(ctx, &MsgHeartbeatReply{base: base{raw: body}}, nil)
	case wsOpMessage:
		// 压缩版本重新解包再调用，直到 ver==0
		switch ver {
		case wsVerZlib:
			de, err := zlibDe(body)
			if err != nil {
				l.push(ctx, nil, fmt.Errorf("failed to decode zlib msg: %s", err))
				return
			}
			l.handles(ctx, l.split(de))
		case wsVerBrotli:
			de, err := brotliDe(body)
			if err != nil {
				l.push(ctx, nil, fmt.Errorf("failed to decode brotli msg: %s", err))
				return
			}
			l.handles(ctx, l.split(de))
		case wsVerPlain:
			l.handlePlain(ctx, body)
		}
	}
}

// split 压缩过的body需要拆包
func (l *Live) split(b []byte) [][]byte {
	var packs [][]byte
	for i, size := uint32(0), uint32(0); i < uint32(len(b)); i += size {
		size = binary.BigEndian.Uint32(b[i : i+4])
		packs = append(packs, b[i:i+size])
	}
	return packs
}
func (l *Live) handles(ctx context.Context, bs [][]byte) {
	for _, b := range bs {
		go l.handle(ctx, b)
	}
}
func (l *Live) handlePlain(ctx context.Context, body []byte) {
	var cmd struct {
		CMD string `json:"cmd"`
	}
	if err := json.Unmarshal(body, &cmd); err != nil {
		l.push(ctx, nil, fmt.Errorf("failed to unmarshal plain msg: %s", err))
		return
	}
	m := l.switchCmd(cmd.CMD, body)
	l.push(ctx, m, nil)
}
func (l *Live) switchCmd(cmd string, body []byte) Msg {
	var m Msg
	b := base{raw: body}
	switch cmd {
	case cmdDanmaku:
		m = &MsgDanmaku{base: b}
	case cmdSendGift:
		m = &MsgSendGift{base: b}
	case cmdComboSend:
		m = &MsgComboSend{base: b}
	case cmdRoomRealTimeMessageUpdate:
		m = &MsgFansUpdate{base: b}
	case cmdOnlineRankCount:
		m = &MsgOnlineRankCount{base: b}
	case cmdSuperChatMessage:
		m = &MsgSuperChatMessage{base: b}
	case cmdHotRankSettlement:
		m = &MsgHotRankSettlement{base: b}
	case cmdOnlineRankTop3:
		m = &MsgOnlineRankTop3{base: b}
	case cmdRoomBlockMsg:
		m = &MsgRoomBlockMsg{base: b}
	case cmdStopLiveRoomList:
		m = &MsgStopLiveRoomList{base: b}
	case cmdOnlineRankV2:
		m = &MsgOnlineRankV2{base: b}
	case cmdNoticeMsg:
		m = &MsgNoticeMsg{base: b}
	case cmdHotRankChanged:
		m = &MsgHotRankChanged{base: b}
	case cmdGuardBuy:
		m = &MsgGuardBuy{base: b}
	case cmdSuperChatMessageJPN:
		m = &MsgSuperChatMessageJPN{base: b}
	case cmdUserToastMsg:
		m = &MsgUserToastMsg{base: b}
	case cmdSuperChatMessageDelete:
		m = &MsgSuperChatMessageDelete{base: b}
	case cmdAnchorLotStart:
		m = &MsgAnchorLotStart{base: b}
	case cmdAnchorLotCheckStatus:
		m = &MsgAnchorLotCheckStatus{base: b}
	case cmdAnchorLotAward:
		m = &MsgAnchorLotAward{base: b}
	case cmdAnchorLotEnd:
		m = &MsgAnchorLotEnd{base: b}
	case cmdRoomChange:
		m = &MsgRoomChange{base: b}
	case cmdVoiceJoinList:
		m = &MsgVoiceJoinList{base: b}
	case cmdVoiceJoinRoomCountInfo:
		m = &MsgVoiceJoinRoomCountInfo{base: b}
	case cmdAttention:
		m = &MsgAttention{base: b}
	case cmdShare:
		m = &MsgShare{base: b}
	case cmdSpecialAttention:
		m = &MsgSpecialAttention{base: b}
	case cmdSysMsg:
		m = &MsgSysMsg{base: b}
	case cmdPreparing:
		m = &MsgPreparing{base: b}
	case cmdLive:
		m = &MsgLive{base: b}
	case cmdRoomRank:
		m = &MsgRoomRank{base: b}
	case cmdRoomLimit:
		m = &MsgRoomLimit{base: b}
	case cmdBlock:
		m = &MsgBlock{base: b}
	case cmdPkPre:
		m = &MsgPkPre{base: b}
	case cmdPkEnd:
		m = &MsgPkEnd{base: b}
	case cmdPkSettle:
		m = &MsgPkSettle{base: b}
	case cmdSysGift:
		m = &MsgSysGift{base: b}
	case cmdHotRank:
		m = &MsgHotRank{base: b}
	case cmdActivityRedPacket:
		m = &MsgActivityRedPacket{base: b}
	case cmdPkMicEnd:
		m = &MsgPkMicEnd{base: b}
	case cmdPlayTag:
		m = &MsgPlayTag{base: b}
	case cmdGuardMsg:
		m = &MsgGuardMsg{base: b}
	case cmdPlayProgressBar:
		m = &MsgPlayProgressBar{base: b}
	case cmdHotRoomNotify:
		m = &MsgHotRoomNotify{base: b}
	case cmdRefresh:
		m = &MsgRefresh{base: b}
	case cmdRound:
		m = &MsgRound{base: b}
	case cmdWelcomeGuard:
		m = &MsgWelcomeGuard{base: b}
	case cmdEntryEffect:
		m = &MsgEntryEffect{base: b}
	case cmdWelcome:
		m = &MsgWelcome{base: b}
	case cmdLiveInteractiveGame:
		m = &MsgLiveInteractiveGame{base: b}
	case cmdVoiceJoinStatus:
		m = &MsgVoiceJoinStatus{base: b}
	case cmdCutOff:
		m = &MsgCutOff{base: b}
	case cmdSpecialGift:
		m = &MsgSpecialGift{base: b}
	case cmdNewGuardCount:
		m = &MsgNewGuardCount{base: b}
	case cmdRoomAdmins:
		m = &MsgRoomAdmins{base: b}
	case cmdActivityBannerUpdateV2:
		m = &MsgActivityBannerUpdateV2{base: b}
	case cmdInteractWord:
		m = &MsgInteractWord{base: b}
	case cmdPkBattlePre:
		m = &MsgPkBattlePre{base: b}
	case cmdPkBattleSettle:
		m = &MsgPkBattleSettle{base: b}
	case cmdPkBattleStart:
		m = &MsgPkBattleStart{base: b}
	case cmdPkBattleProcess:
		m = &MsgPkBattleProcess{base: b}
	case cmdPkEnding:
		m = &MsgPkEnding{base: b}
	case cmdPkBattleEnd:
		m = &MsgPkBattleEnd{base: b}
	case cmdPkBattleSettleUser:
		m = &MsgPkBattleSettleUser{base: b}
	case cmdPkBattleSettleV2:
		m = &MsgPkBattleSettleV2{base: b}
	case cmdPkLotteryStart:
		m = &MsgPkLotteryStart{base: b}
	case cmdPkBestUname:
		m = &MsgPkBestUname{base: b}
	case cmdCallOnOpposite:
		m = &MsgCallOnOpposite{base: b}
	case cmdAttentionOpposite:
		m = &MsgAttentionOpposite{base: b}
	case cmdShareOpposite:
		m = &MsgShareOpposite{base: b}
	case cmdAttentionOnOpposite:
		m = &MsgAttentionOnOpposite{base: b}
	case cmdPkMatchInfo:
		m = &MsgPkMatchInfo{base: b}
	case cmdPkMatchOnlineGuard:
		m = &MsgPkMatchOnlineGuard{base: b}
	case cmdPkWinningStreak:
		m = &MsgPkWinningStreak{base: b}
	case cmdPkDanmuMsg:
		m = &MsgPkDanmuMsg{base: b}
	case cmdPkSendGift:
		m = &MsgPkSendGift{base: b}
	case cmdPkInteractWord:
		m = &MsgPkInteractWord{base: b}
	case cmdPkAttention:
		m = &MsgPkAttention{base: b}
	case cmdPkShare:
		m = &MsgPkShare{base: b}
	case cmdWatChedChange:
		m = &MsgWatChed{base: b}
	default:
		m = &MsgGeneral{base: b}
	}
	return m
}
func (l *Live) push(ctx context.Context, msg Msg, err error) {
	go func(c context.Context, m Msg, e error) {
		// 五秒超时
		after := time.NewTimer(5 * time.Second)
		defer after.Stop()

		select {
		case <-c.Done():
			l.info("push stopped")
			return
		case <-after.C:
			return
		case l.Rev <- &Transport{Msg: m, Error: e}:
			return
		}
	}(ctx, msg, err)
}
func (l *Live) log(v ...interface{}) {
	if l.debug {
		l.logger.Println(v...)
	}
}
func (l *Live) logf(format string, v ...interface{}) {
	if l.debug {
		l.logger.Printf(format, v...)
	}
}
func (l *Live) info(format string, v ...interface{}) {
	l.logf("[INFO] "+format, v...)
}
func (l *Live) error(format string, v ...interface{}) {
	l.logf("[ERROR] "+format, v)
}
