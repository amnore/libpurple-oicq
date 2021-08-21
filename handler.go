package main

/*
   #cgo pkg-config: purple
   #include <purple.h>
*/
import "C"
import (
	"fmt"
	"html"
	"strings"
	"sync"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/gotk3/gotk3/glib"
)

var (
	messageToSelfId sync.Map = sync.Map{}
)

func registerHandlers(account *C.PurpleAccount) {
	cli := clients[account]

	cli.OnPrivateMessage(
		func(q *client.QQClient, pm *message.PrivateMessage) {
			onPrivateMessage(account, q, pm)
		},
	)
	cli.OnSelfPrivateMessage(
		func(q *client.QQClient, pm *message.PrivateMessage) {
			onPrivateMessage(account, q, pm)
		},
	)

	cli.OnReceivedOfflineFile(
		func(q *client.QQClient, ofe *client.OfflineFileEvent) {
			onReceivedOfflineFile(account, q, ofe)
		},
	)

	cli.OnSelfGroupMessage(
		func(q *client.QQClient, gm *message.GroupMessage) {
			onGroupMessage(account, q, gm)
		},
	)

	cli.OnGroupMessage(
		func(q *client.QQClient, gm *message.GroupMessage) {
			onGroupMessage(account, q, gm)
		},
	)
}

func onPrivateMessage(
	account *C.PurpleAccount,
	client *client.QQClient,
	msg *message.PrivateMessage,
) {
	if msg.Sender.Uin == msg.Target && msg.Self == msg.Target {
		_, ok := messageToSelfId.LoadOrStore(msg.Id, struct{}{})
		if ok {
			return
		}
	}

	str, err := parseMessageIntoHtml(msg.Elements)
	if err != nil {
		debugError("onPrivateMessage: %s\n", err.Error())
		return
	}

	debugInfo("onPrivateMessage: sender=%d, recv=%d, msg=\"%s\"\n",
		msg.Sender.Uin, msg.Target, str)

	glib.IdleAdd(func() {
		var other *C.char
		if msg.Self == msg.Sender.Uin {
			other = uidToCString(msg.Target)
		} else {
			other = uidToCString(msg.Sender.Uin)
		}

		cstr := C.CString(str)
		defer freeCString(other, cstr)

		C.serv_got_im(
			C.purple_account_get_connection(account),
			other,
			cstr,
			C.PURPLE_MESSAGE_RECV|C.PURPLE_MESSAGE_IMAGES,
			(C.time_t)(msg.Time),
		)
	})
}

func onGroupMessage(
	account *C.PurpleAccount,
	client *client.QQClient,
	msg *message.GroupMessage,
) {
	str, err := parseMessageIntoHtml(msg.Elements)
	if err != nil {
		debugError("onGroupMessage: %s\n", err.Error())
		return
	}

	id, err := getPurpleChatId(client, msg.GroupCode)
	if err != nil {
		debugError("onGroupMessage: %s\n", err.Error())
		return
	}

	debugInfo("onGroupMessage: sender=%d, msg=\"%s\"\n",
		msg.Sender.Uin, str)
	glib.IdleAdd(func() {
		sender := uidToCString(msg.Sender.Uin)
		cstr := C.CString(str)
		defer freeCString(sender, cstr)

		C.serv_got_chat_in(
			C.purple_account_get_connection(account),
			id,
			sender,
			0,
			cstr,
			C.time_t(msg.Time),
		)
	})
}

func onReceivedOfflineFile(
	account *C.PurpleAccount,
	client *client.QQClient,
	msg *client.OfflineFileEvent,
) {
	debugInfo("onReceivedOfflineFile: %v\n", *msg)

	sender := uidToCString(msg.Sender)
	cstr := C.CString(fmt.Sprintf(
		"<a href=\"%s\">%s(%s)</a>",
		strings.ReplaceAll(msg.DownloadUrl, "\"", "\\\""),
		html.EscapeString(msg.FileName),
		sizeToString(msg.FileSize),
	))
	defer freeCString(sender, cstr)

	conv := C.purple_find_conversation_with_account(
		C.PURPLE_CONV_TYPE_IM, sender, account)
	if conv == nil {
		conv = C.purple_conversation_new(
			C.PURPLE_CONV_TYPE_IM, account, sender)
	}

	C.serv_got_im(
		C.purple_account_get_connection(account),
		sender,
		cstr,
		C.PURPLE_MESSAGE_RECV,
		C.time(nil),
	)
}

func sizeToString(size int64) string {
	unitSizes := []int64{1024 * 1024 * 1024, 1024 * 1024, 1024, 1}
	unitNames := []string{"GB", "MB", "KB", "B"}

	i := 0
	for size/unitSizes[i] == 0 && i < len(unitSizes)-1 {
		i++
	}

	return fmt.Sprintf("%g%s",
		float64(size)/float64(unitSizes[i]), unitNames[i])
}
