package main

/*
   #cgo pkg-config: purple glib-2.0
   #include <glib.h>
   #include <purple.h>

   void set_buddy_status(PurpleBuddy *buddy, PurpleStatusPrimitive status);

   typedef const char * conststring;
*/
import "C"

import (
	"fmt"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/gotk3/gotk3/glib"
)

var clients map[*C.PurpleAccount]*client.QQClient = make(map[*C.PurpleAccount]*client.QQClient)

//export login
func login(account *C.PurpleAccount) {
	debugInfo("login\n")

	conn := C.purple_account_get_connection(account)
	msg := C.CString("Connecting")
	defer freeCString(msg)

	C.purple_connection_update_progress(conn, msg, 0, 2)
	C.purple_connection_set_state(conn, C.PURPLE_CONNECTING)

	go func() {
		err := doLogin(account)
		if err == nil {
			glib.IdleAdd(func() {
				msg := C.CString("Connected")
				defer freeCString(msg)

				C.purple_connection_update_progress(conn, msg, 1, 2)
				C.purple_connection_set_state(conn, C.PURPLE_CONNECTED)
			})

			registerHandlers(account)

			err := getContacts(account)
			if err != nil {
				debugError("getContacts: %s\n", err.Error())
			}

			return
		}

		msg := fmt.Sprintf("Login failed: %s", err.Error())
		debugError("%s\n", msg)
		glib.IdleAdd(func() {
			cmsg := C.CString(msg)
			defer freeCString(cmsg)

			C.purple_connection_error_reason(
				conn,
				C.PURPLE_CONNECTION_ERROR_AUTHENTICATION_FAILED,
				cmsg,
			)
		})
	}()
}

func doLogin(account *C.PurpleAccount) error {
	password := C.GoString(C.purple_account_get_password(account))
	username, err := parseUidWithError(C.purple_account_get_username(account))
	if err != nil {
		return err
	}

	cli := client.NewClient(username, password)
	clients[account] = cli
	resp, err := cli.Login()
	for {
		switch {
		case err != nil:
			return err

		case resp.Success:
			return nil
		}

		switch resp.Error {
		case client.SliderNeededError:
			debugInfo("Captcha needed: %s\n", resp.VerifyUrl)

			var ticket string
			ticket, err = getSliderTicket(account, resp.VerifyUrl)
			if err != nil {
				return err
			}

			resp, err = cli.SubmitTicket(ticket)

		case client.SMSOrVerifyNeededError:
			debugInfo("SMS or QR code needed, using QR code: %s\n",
				resp.VerifyUrl)

			err = waitForQRCode(account, resp.VerifyUrl)
			if err != nil {
				return err
			}

			resp, err = cli.Login()

		case client.NeedCaptcha, client.UnsafeDeviceError,
			client.SMSNeededError, client.TooManySMSRequestError,
			client.OtherLoginError, client.UnknownLoginError:
			return fmt.Errorf("Verification needed: %d", resp.Error)
		}
	}
}

func getContacts(account *C.PurpleAccount) error {
	debugInfo("Get contacts\n")

	cli := clients[account]
	err := cli.ReloadFriendList()
	if err != nil {
		return err
	}

	err = cli.ReloadGroupList()
	if err != nil {
		return err
	}

	glib.IdleAdd(func() {
		groupname := C.CString("OICQ")
		defer freeCString(groupname)
		group := C.purple_find_group(groupname)
		if group == nil {
			group = C.purple_group_new(groupname)
			C.purple_blist_add_group(group, nil)
		}

		for _, friend := range cli.FriendList {
			id := uidToCString(friend.Uin)
			defer freeCString(id)

			buddy := C.purple_find_buddy(account, id)
			if buddy == nil {
				nickname := C.CString(friend.Nickname)
				defer freeCString(nickname)

				buddy = C.purple_buddy_new(account, id, nickname)
				C.purple_blist_add_buddy(buddy, nil, group, nil)
			}

			C.set_buddy_status(buddy, C.PURPLE_STATUS_AVAILABLE)
			go getBuddyIcon(buddy)
		}

		for _, qgroup := range cli.GroupList {
			id := uidToCString(qgroup.Code)
			defer freeCString(id)

			chat := C.purple_blist_find_chat(account, id)
			if chat == nil {
				name := C.CString(qgroup.Name)
				defer freeCString(name)

				chat = C.purple_chat_new(account, name,
					newChatComponents(qgroup.Code))
				C.purple_blist_add_chat(chat, group, nil)
			}
			go getChatIcon(chat)
		}
	})

	return nil
}

func getBuddyIcon(buddy *C.PurpleBuddy) error {
	uid := parseUid(C.purple_buddy_get_name(buddy))
	debugInfo("getBuddyIcon: %d\n", uid)

	return fetchAndSetIcon(
		&buddy.node,
		fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=640", uid),
	)
}

func getChatIcon(chat *C.PurpleChat) error {
	code := getChatGroupCode(C.purple_chat_get_components(chat))
	debugInfo("getChatIcon: %d\n", code)

	return fetchAndSetIcon(
		&chat.node,
		fmt.Sprintf("https://p.qlogo.cn/gh/%d/%d/640", code, code),
	)
}

func fetchAndSetIcon(node *C.PurpleBlistNode, url string) error {
	icon, ts, err := fetchIcon(url, getIconTimestamp(node))
	if err != nil {
		return err
	}

	if icon == nil {
		debugInfo("fetchAndSetIcon: url %s not modified\n", url)
		return nil
	}

	setIconForNode(node, icon, ts)
	return nil
}

func setIconForNode(node *C.PurpleBlistNode, icon []byte, lastModified string) {
	glib.IdleAdd(func() {
		k := C.CString("icon_last_modified")
		v := C.CString(lastModified)
		defer freeCString(k, v)

		C.purple_blist_node_set_string(node, k, v)
		C.purple_buddy_icons_node_set_custom_icon(
			node, (*C.uchar)(C.CBytes(icon)), C.size_t(len(icon)))
	})
}

func getIconTimestamp(node *C.PurpleBlistNode) *string {
	k := C.CString("icon_last_modified")
	defer freeCString(k)

	str := C.purple_blist_node_get_string(node, k)
	if str == nil {
		return nil
	}

	s := C.GoString(str)
	return &s
}

func fetchIcon(url string, modifiedSince *string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	if modifiedSince != nil {
		req.Header.Add("If-Modified-Since", *modifiedSince)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	ts := resp.Header.Get("Last-Modified")
	if resp.StatusCode == 304 {
		return nil, ts, nil
	}
	data, err := readAll(resp.Body, resp.ContentLength)
	return data, ts, err
}

//export joinChat
func joinChat(connection *C.PurpleConnection, data *C.GHashTable) {
	code := getChatGroupCode(data)
	cli := clients[C.purple_connection_get_account(connection)]
	id, err := getPurpleChatId(cli, code)
	if err != nil {
		debugError("joinChat: %s\n", err.Error())
		return
	}

	group := cli.GroupList[id]
	codeStr := C.CString(strconv.FormatInt(code, 10))
	defer freeCString(codeStr)
	debugInfo("joinChat: %s(%d)\n", group.Name, code)
	conv := C.serv_got_joined_chat(connection, id, codeStr)

	debugInfo("members: %v\n", group.Members)

	users := &glib.List{}
	flags := &glib.List{}
	for _, member := range group.Members {
		users = users.Append(uintptr(unsafe.Pointer(
			C.CString(strconv.FormatInt(member.Uin, 10)))))

		var flag C.PurpleConvChatBuddyFlags
		switch member.Permission {
		case client.Administrator:
			flag = C.PURPLE_CBFLAGS_OP
		case client.Owner:
			flag = C.PURPLE_CBFLAGS_FOUNDER
		}

		flags = flags.Append(uintptr(flag))
	}

	C.purple_conv_chat_add_users(
		C.purple_conversation_get_chat_data(conv),
		(*C.GList)(unsafe.Pointer(users.Native())),
		nil,
		(*C.GList)(unsafe.Pointer(flags.Native())),
		C.FALSE,
	)
}

//export chatSend
func chatSend(connection *C.PurpleConnection, id C.int, msg C.conststring,
	flags C.PurpleMessageFlags) C.int {
	str := C.GoString(msg)
	cli := clients[C.purple_connection_get_account(connection)]
	code := cli.GroupList[id].Code
	debugInfo("chatSend: %s\n", str)

	elems, err := parseHtmlIntoMessage(
		MessagingContext{cli, true, code}, str)
	if err != nil {
		debugError("chatSend: %s\n", err.Error())
		return -1
	}

	cli.SendGroupMessage(code, &message.SendingMessage{Elements: elems})
	return 1
}

//export closeConnection
func closeConnection(connection *C.PurpleConnection) {
	debugInfo("close\n")

	account := C.purple_connection_get_account(connection)

	clients[account].Release()
	delete(clients, account)
}

//export getCBAlias
func getCBAlias(connection *C.PurpleConnection, id C.int, who C.conststring) *C.char {
	cli := clients[C.purple_connection_get_account(connection)]
	member := findGroupMember(cli.GroupList[id], parseUid(who))

	if member == nil {
		return C.strdup(who)
	}
	return C.CString(member.Nickname)
}

//export sendIM
func sendIM(
	connection *C.PurpleConnection,
	who C.conststring,
	msg C.conststring,
	flags C.PurpleMessageFlags,
) C.int {
	account := C.purple_connection_get_account(connection)
	cli := clients[account]
	uid := parseUid(who)
	msgstr := C.GoString(msg)
	elems, err := parseHtmlIntoMessage(
		MessagingContext{cli, false, uid},
		msgstr,
	)
	if err != nil {
		debugError("sendIM: %s\n", err.Error())
		return -1
	}

	debugInfo("sendIM: uid=%d, flags=%d, msg=\"%s\"\n",
		uid, flags, msgstr)

	cli.SendPrivateMessage(uid,
		&message.SendingMessage{Elements: elems})
	return 0
}

func main() {}
