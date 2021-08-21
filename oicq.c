#include "_cgo_export.h"

#include <assert.h>
#include <stdio.h>

static void initPlugin(PurplePlugin *plugin) { printf("oicq: Initializing\n"); }

static const char *listIcon(PurpleAccount *account, PurpleBuddy *buddy) {
  return "oicq";
}

static GList *statusTypes(PurpleAccount *account) {
  GList *types = NULL;
  PurpleStatusType *status;

  static const char *status_text[] = {[PURPLE_STATUS_AVAILABLE] = "Available",
                                      [PURPLE_STATUS_OFFLINE] = "Offline"};
  static int statuses[] = {PURPLE_STATUS_AVAILABLE, PURPLE_STATUS_OFFLINE};
  for (int i = 0; i < sizeof(statuses) / sizeof(*statuses); i++) {
    status = purple_status_type_new_full(
        statuses[i], NULL, status_text[statuses[i]], TRUE, TRUE, FALSE);
    types = g_list_append(types, status);
  }

  return types;
}

static GList *chatInfo(PurpleConnection *connection) {
  GList *list = NULL;
  struct proto_chat_entry *e;

  e = g_new0(struct proto_chat_entry, 1);
  *e = (struct proto_chat_entry){
      .label = "Group Code", .identifier = "code", .required = TRUE};
  list = g_list_append(list, e);

  return list;
}

static char *getChatName(GHashTable *data) {
  char *code = g_hash_table_lookup(data, "code");
  printf("getChatName: code=%s\n", code);
  return g_strdup(code);
}

void set_buddy_status(PurpleBuddy *buddy, PurpleStatusPrimitive status) {
  purple_prpl_got_user_status(purple_buddy_get_account(buddy),
                              purple_buddy_get_name(buddy),
                              purple_primitive_get_id_from_type(status), NULL);
}

static PurplePluginProtocolInfo proto_info = {
    OPT_PROTO_IM_IMAGE, /* options */
    NULL,               /* user_splits */
    NULL,               /* protocol_options */
    {
        /* icon_spec */
        "png,jpg,gif",             /* format */
        0,                         /* min_width */
        0,                         /* min_height */
        128,                       /* max_width */
        128,                       /* max_height */
        10000,                     /* max_filesize */
        PURPLE_ICON_SCALE_DISPLAY, /* scale_rules */
    },
    listIcon,                         /* list_icon */
    NULL,                             /* list_emblem */
    NULL,                             /* status_text */
    NULL,                             /* tooltip_text */
    statusTypes,                      /* status_types */
    NULL,                             /* blist_node_menu */
    chatInfo,                         /* chat_info */
    NULL,                             /* chat_info_defaults */
    login,                            /* login */
    closeConnection,                            /* close */
    sendIM,                           /* send_im */
    NULL,                             /* set_info */
    NULL,                             /* send_typing */
    NULL,                             /* get_info */
    NULL,                             /* set_status */
    NULL,                             /* set_idle */
    NULL,                             /* change_passwd */
    NULL,                             /* add_buddy */
    NULL,                             /* add_buddies */
    NULL,                             /* remove_buddy */
    NULL,                             /* remove_buddies */
    NULL,                             /* add_permit */
    NULL,                             /* add_deny */
    NULL,                             /* rem_permit */
    NULL,                             /* rem_deny */
    NULL,                             /* set_permit_deny */
    joinChat,                         /* join_chat */
    NULL,                             /* reject_chat */
    getChatName,                      /* get_chat_name */
    NULL,                             /* chat_invite */
    NULL,                             /* chat_leave */
    NULL,                             /* chat_whisper */
    chatSend,                         /* chat_send */
    NULL,                             /* keepalive */
    NULL,                             /* register_user */
    NULL,                             /* get_cb_info, deprecated */
    NULL,                             /* get_cb_away, deprecated */
    NULL,                             /* alias_buddy */
    NULL,                             /* group_buddy */
    NULL,                             /* rename_group */
    NULL,                             /* buddy_free */
    NULL,                             /* convo_closed */
    NULL,                             /* normalize */
    NULL,                             /* set_buddy_icon */
    NULL,                             /* remove_group */
    NULL,                             /* get_cb_real_name */
    NULL,                             /* set_chat_topic */
    NULL,                             /* find_blist_chat */
    NULL,                             /* roomlist_get_list */
    NULL,                             /* roomlist_cancel */
    NULL,                             /* roomlist_expand_category */
    NULL,                             /* can_receive_file */
    NULL,                             /* send_file */
    NULL,                             /* new_xfer */
    NULL,                             /* offline_message */
    NULL,                             /* whiteboard_prpl_ops */
    NULL,                             /* send_raw */
    NULL,                             /* roomlist_room_serialize */
    NULL,                             /* unregister_user */
    NULL,                             /* send_attention */
    NULL,                             /* get_attention_types */
    sizeof(PurplePluginProtocolInfo), /* struct_size */
    NULL,                             /* get_account_text_table */
    NULL,                             /* initiate_media */
    NULL,                             /* get_media_caps */
    NULL,                             /* get_moods */
    NULL,                             /* set_public_alias */
    NULL,                             /* get_public_alias */
    NULL,                             /* add_buddy_with_invite */
    NULL,                             /* add_buddies_with_invite */
    getCBAlias,                             /* get_cb_alias */
    NULL,                             /* chat_can_receive_file */
    NULL                              /* chat_send_file */
};

static PurplePluginInfo info = {PURPLE_PLUGIN_MAGIC,
                                PURPLE_MAJOR_VERSION,
                                PURPLE_MINOR_VERSION,
                                PURPLE_PLUGIN_PROTOCOL,
                                NULL,
                                0,
                                NULL,
                                PURPLE_PRIORITY_DEFAULT,
                                "prpl-amnore-oicq",
                                "OICQ",
                                "0.0.1",
                                "Support for Tencent QQ(Android) protocol",
                                "Support for Tencent QQ(Android) protocol",
                                "amnore",
                                "https://github.com/amnore/libpurple-oicq",
                                NULL,        /* load */
                                NULL,        /* unload */
                                NULL,        /* destroy */
                                NULL,        /* ui_info */
                                &proto_info, /* extra_info */
                                NULL,        /* prefs_info */
                                NULL,        /* actions */
                                NULL,        /* padding... */
                                NULL,
                                NULL,
                                NULL};

PURPLE_INIT_PLUGIN(oicq, initPlugin, info)
