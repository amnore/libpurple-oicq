package main

/*
   #cgo pkg-config: purple glib-2.0
   #include <purple.h>
   #include <glib.h>
   #include <stdint.h>

   GHashTable *new_chat_components(int64_t code);
   int64_t get_chat_group_code(GHashTable *data);
*/
import "C"
import (
	"fmt"

	"github.com/Mrs4s/MiraiGo/client"
)

func newChatComponents(code int64) *C.GHashTable {
	return C.new_chat_components(C.int64_t(code))
}

func getPurpleChatId(client *client.QQClient, code int64) (C.int, error) {
        for i, group := range client.GroupList {
                if group.Code == code {
                        return C.int(i), nil
                }
        }

        return -1, fmt.Errorf("group %d not joined", code)
}

func findGroupMember(group *client.GroupInfo, uid int64) *client.GroupMemberInfo {
        for _, member := range group.Members {
                if member.Uin == uid {
                        return member
                }
        }

        return nil
}

func getChatGroupCode(data *C.GHashTable) int64 {
	return int64(C.get_chat_group_code(data));
}
