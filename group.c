#include "_cgo_export.h"

#include <stdlib.h>

GHashTable *new_chat_components(int64_t code) {
  GHashTable *t = g_hash_table_new_full(g_str_hash, g_str_equal, NULL, g_free);
  char *v = g_strdup_printf("%lld", code);
  
  g_hash_table_insert(t, "code", v);
  return t;
}

int64_t get_chat_group_code(GHashTable *data) {
  const char *str = g_hash_table_lookup(data, "code");
  if (!str) {
    return -1;
  }

  return atoll(str);
}
