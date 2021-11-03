package live

// ops
const (
	wsOpHeartbeat        = 2 // 心跳
	wsOpHeartbeatReply   = 3 // 心跳回应
	wsOpMessage          = 5 // 弹幕消息等
	wsOpEnterRoom        = 7 // 请求进入房间
	wsOpEnterRoomSuccess = 8 // 进房回应
)

// Header
const (
	wsPackHeaderTotalLen = 16 // 头部字节大小
	wsPackageLen         = 4
	wsHeaderLen          = 2
	wsVerLen             = 2
	wsOpLen              = 4
	wsSequenceLen        = 4
)

// ws header default
const (
	wsHeaderDefaultSequence = 1
)

// version protocol
const (
	wsVerPlain  = 0
	wsVerInt    = 1
	wsVerZlib   = 2
	wsVerBrotli = 3
)

// cmd
const (
	cmdAttention                 = "ATTENTION"                     // 用户关注【答谢关注】
	cmdShare                     = "SHARE"                         // 用户分享直播间
	cmdSpecialAttention          = "SPECIAL_ATTENTION"             // 特别关注直播间
	cmdSysMsg                    = "SYS_MSG"                       //
	cmdPreparing                 = "PREPARING"                     // 下播
	cmdLive                      = "LIVE"                          // 开播
	cmdRoomChange                = "ROOM_CHANGE"                   // 房间信息改变
	cmdRoomRank                  = "ROOM_RANK"                     // 排名改变
	cmdRoomLimit                 = "ROOM_LIMIT"                    //
	cmdBlock                     = "BLOCK"                         //
	cmdPkPre                     = "PK_PRE"                        //
	cmdPkEnd                     = "PK_END"                        // PK判断胜负
	cmdPkSettle                  = "PK_SETTLE"                     //
	cmdSysGift                   = "SYS_GIFT"                      //
	cmdHotRankSettlement         = "HOT_RANK_SETTLEMENT"           // 荣登热门榜topX
	cmdHotRank                   = "HOT_RANK"                      // 热门榜xx榜topX
	cmdOnlineRankTop3            = "ONLINE_RANK_TOP3"              // 高能榜TOP3改变
	cmdActivityRedPacket         = "ACTIVITY_RED_PACKET"           //
	cmdPkMicEnd                  = "PK_MIC_END"                    //
	cmdStopLiveRoomList          = "STOP_LIVE_ROOM_LIST"           // 刚刚停止了直播的直播间
	cmdPlayTag                   = "PLAY_TAG"                      //
	cmdGuardMsg                  = "GUARD_MSG"                     // 舰长消息
	cmdPlayProgressBar           = "PLAY_PROGRESS_BAR"             //
	cmdHotRoomNotify             = "HOT_ROOM_NOTIFY"               //
	cmdRefresh                   = "REFRESH"                       //
	cmdRound                     = "ROUND"                         //
	cmdDanmaku                   = "DANMU_MSG"                     // 弹幕消息
	cmdWelcomeGuard              = "WELCOME_GUARD"                 //
	cmdEntryEffect               = "ENTRY_EFFECT"                  // 舰长、高能榜、老爷进入【欢迎舰长】
	cmdWelcome                   = "WELCOME"                       // 欢迎进入房间(似乎已废弃)
	cmdSuperChatMessageJPN       = "SUPER_CHAT_MESSAGE_JPN"        // 醒目留言日文翻译
	cmdSuperChatMessage          = "SUPER_CHAT_MESSAGE"            // 醒目留言
	cmdSuperChatMessageDelete    = "SUPER_CHAT_MESSAGE_DELETE"     // 删除醒目留言 (似乎有时候并不会发,同时结束时间在 CmdSuperChatMessage 可以获取)
	cmdLiveInteractiveGame       = "LIVE_INTERACTIVE_GAME"         //
	cmdSendGift                  = "SEND_GIFT"                     // 投喂礼物
	cmdRoomBlockMsg              = "ROOM_BLOCK_MSG"                // 用户被禁言
	cmdComboSend                 = "COMBO_SEND"                    // 连击礼物
	cmdAnchorLotStart            = "ANCHOR_LOT_START"              // 天选之人开始完整信息
	cmdAnchorLotEnd              = "ANCHOR_LOT_END"                // 天选之人获奖id
	cmdAnchorLotAward            = "ANCHOR_LOT_AWARD"              // 天选结果推送
	cmdVoiceJoinRoomCountInfo    = "VOICE_JOIN_ROOM_COUNT_INFO"    // 申请连麦队列变化
	cmdVoiceJoinList             = "VOICE_JOIN_LIST"               // 连麦申请、取消连麦申请
	cmdVoiceJoinStatus           = "VOICE_JOIN_STATUS"             // 开始连麦、结束连麦
	cmdCutOff                    = "CUT_OFF"                       // 被超管切断
	cmdSpecialGift               = "SPECIAL_GIFT"                  // 节奏风暴
	cmdNewGuardCount             = "NEW_GUARD_COUNT"               // 船员数量改变事件
	cmdRoomAdmins                = "ROOM_ADMINS"                   // 房管数量改变
	cmdGuardBuy                  = "GUARD_BUY"                     // 用户上舰长
	cmdUserToastMsg              = "USER_TOAST_MSG"                // 上船附带的通知
	cmdNoticeMsg                 = "NOTICE_MSG"                    // 广播消息(别的直播间投递高价礼物对所有直播间发起的广播)
	cmdAnchorLotCheckStatus      = "ANCHOR_LOT_CHECKSTATUS"        // 天选时刻前的审核
	cmdActivityBannerUpdateV2    = "ACTIVITY_BANNER_UPDATE_V2"     //
	cmdRoomRealTimeMessageUpdate = "ROOM_REAL_TIME_MESSAGE_UPDATE" // 粉丝数量改变
	cmdInteractWord              = "INTERACT_WORD"                 // 用户进入直播间
	cmdOnlineRankCount           = "ONLINE_RANK_COUNT"             // 高能榜数量更新
	cmdOnlineRankV2              = "ONLINE_RANK_V2"                // 高能榜数据
	cmdPkBattlePre               = "PK_BATTLE_PRE"                 // 大乱斗准备，10秒后开始
	cmdPkBattleSettle            = "PK_BATTLE_SETTLE"              //
	cmdHotRankChanged            = "HOT_RANK_CHANGED"              // 热门榜改变
	cmdPkBattleStart             = "PK_BATTLE_START"               // 大乱斗开始
	cmdPkBattleProcess           = "PK_BATTLE_PROCESS"             // 大乱斗双方送礼
	cmdPkEnding                  = "PK_ENDING"                     // 大乱斗尾声，最后几秒
	cmdPkBattleEnd               = "PK_BATTLE_END"                 // 大乱斗结束
	cmdPkBattleSettleUser        = "PK_BATTLE_SETTLE_USER"         //
	cmdPkBattleSettleV2          = "PK_BATTLE_SETTLE_V2"           //
	cmdPkLotteryStart            = "PK_LOTTERY_START"              // 大乱斗胜利后的抽奖
	cmdPkBestUname               = "PK_BEST_UNAME"                 // PK最佳助攻
	cmdCallOnOpposite            = "CALL_ON_OPPOSITE"              // 本直播间的观众跑去对面串门
	cmdAttentionOpposite         = "ATTENTION_OPPOSITE"            // 本直播间观众关注了对面主播
	cmdShareOpposite             = "SHARE_OPPOSITE"                // 本直播间观众分享了对面直播间
	cmdAttentionOnOpposite       = "ATTENTION_ON_OPPOSITE"         // 对面观众关注了本直播间
	cmdPkMatchInfo               = "PK_MATCH_INFO"                 // 获取对面直播间信息
	cmdPkMatchOnlineGuard        = "PK_MATCH_ONLINE_GUARD"         // 获取对面直播间舰长在线人数
	cmdPkWinningStreak           = "PK_WINNING_STREAK"             // 大乱斗连胜事件
	cmdPkDanmuMsg                = "PK_DANMU_MSG"                  // 对面的弹幕消息
	cmdPkSendGift                = "PK_SEND_GIFT"                  // 对面的礼物消息
	cmdPkInteractWord            = "PK_INTERACT_WORD"              // 对面的用户进入
	cmdPkAttention               = "PK_ATTENTION"                  // 对面新增关注
	cmdPkShare                   = "PK_SHARE"                      // 对面有人分享直播间
)

const (
	WsDefaultHost = "wss://broadcastlv.chat.bilibili.com/sub"
)
