#include "_cgo_export.h"

static void send_ticket(gpointer user_data, const char *ticket) {
  sendString((char *)ticket);
}

static void send_null(gpointer user_data) { sendString(NULL); }

static void send_true() { sendBool(true); }

static void send_false() { sendBool(false); }

void get_slider_ticket(PurpleAccount *account, const char *url) {
  PurpleConnection *conn = purple_account_get_connection(account);

  purple_notify_uri(conn, url);
  purple_request_input(conn, "Captcha Required",
                       "Please paste captcha ticket here", NULL, NULL, FALSE,
                       FALSE, NULL, "OK", G_CALLBACK(send_ticket), "Cancel",
                       G_CALLBACK(send_null), account, NULL, NULL, NULL);
}

void wait_for_qrcode(PurpleAccount *account, const char *url) {
  PurpleConnection *conn = purple_account_get_connection(account);

  purple_notify_uri(conn, url);
  purple_request_ok_cancel(conn, "Please Scan QR Code",
                           "Please scan the QR code displayed in your browser",
                           NULL, 0, account, NULL, NULL, NULL, send_true,
                           send_false);
}
