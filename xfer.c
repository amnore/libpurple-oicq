#include "_cgo_export.h"

PurpleXfer *new_xfer_receive(PurpleAccount *account, const char *who,
                             const char *url, const char *filename,
                             size_t size) {
  PurpleXfer *xfer = purple_xfer_new(account, PURPLE_XFER_RECEIVE, who);
  if (!xfer) {
    return NULL;
  }

  purple_xfer_set_filename(xfer, filename);
  purple_xfer_set_size(xfer, size);
  return xfer;
}
