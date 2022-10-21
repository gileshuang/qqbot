package qbmsg

// doc: https://docs.go-cqhttp.org/event/

type EventSender struct {
	Card     string `json:"card"`
	Title    string `json:"Title"`
	NickName string `json:"nickname"`
	UserId   int64  `json:"user_id"`
	Age      int32  `json:"Age"`
	Area     string `json:"Area"`
	Level    string `json:"level"`
	Role     string `json:"role"`
	Sex      string `json:"sex"`
}

type Event struct {
	// 通用数据
	Time     int64  `json:"time"`
	SelfId   int64  `json:"self_id"`
	PortType string `json:"post_type"`
	// 消息上报 post_type: message
	SubType     string      `json:"sub_type"`
	MessageId   int32       `json:"message_id"`
	UserId      int64       `json:"user_id"`
	GroupId     int64       `json:"group_id"`
	Message     string      `json:"message"`
	RawMessage  string      `json:"raw_message"`
	MessageType string      `json:"message_type"`
	Font        int64       `json:"font"`
	Sender      EventSender `json:"sender"`
	// 消息上报-私聊消息 message_type: private
	// 请求上报 post_type: request
	RequestType string `json:"request_type"`
	// 通知上报 post_type: notice
	NoticeType string `json:"notoce_type"`
	// 通知上报 post_type: meta_event
	MetaEventType string `json:"meta_event_type"`
	Interval      int    `json:"interval"`
}
