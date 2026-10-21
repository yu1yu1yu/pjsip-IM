#include "../pjsip/bridge.h"
#include <pjsua-lib/pjsua.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#define QUEUE_SIZE 32
static pjgo_event queue[QUEUE_SIZE];
static unsigned head, count;
static int overflow, created;
static pjsua_acc_id account = PJSUA_INVALID_ID;

static void copy_text(char *dest, unsigned capacity, const pj_str_t *src)
{
    unsigned len = src && src->slen > 0 ? (unsigned)src->slen : 0;
    if (len >= capacity) {
        overflow = 1;
        len = capacity - 1;
    }
    if (len) memcpy(dest, src->ptr, len);
    dest[len] = 0;
}

static pjgo_event *new_event(int kind)
{
    pjgo_event *event;
    if (count == QUEUE_SIZE) {
        overflow = 1;
        return NULL;
    }
    event = &queue[(head + count++) % QUEUE_SIZE];
    memset(event, 0, sizeof(*event));
    event->kind = kind;
    return event;
}

static void on_message(pjsua_call_id call, const pj_str_t *from,
                       const pj_str_t *to, const pj_str_t *contact,
                       const pj_str_t *mime, const pj_str_t *body,
                       pjsip_rx_data *rdata, pjsua_acc_id acc_id)
{
    pjgo_event *event = new_event(1);
    pjsip_uri *uri;
    pjsip_sip_uri *sip_uri;
    PJ_UNUSED_ARG(call);
    PJ_UNUSED_ARG(to);
    PJ_UNUSED_ARG(contact);
    PJ_UNUSED_ARG(acc_id);
    if (!event) return;
    copy_text(event->peer, sizeof(event->peer), from);
    /* From is an identity and may omit the listening port. For this
     * loopback-only transport, reply to the actual UDP source instead. */
    uri = pjsip_uri_get_uri(rdata->msg_info.from->uri);
    if (PJSIP_URI_SCHEME_IS_SIP(uri)) {
        sip_uri = (pjsip_sip_uri *)uri;
        if (sip_uri->user.slen <= 256)
            pj_ansi_snprintf(event->reply_to, sizeof(event->reply_to),
                             "sip:%.*s@127.0.0.1:%d",
                             (int)sip_uri->user.slen, sip_uri->user.ptr,
                             rdata->pkt_info.src_port);
    }
    copy_text(event->content_type, sizeof(event->content_type), mime);
    copy_text(event->body, sizeof(event->body), body);
    event->body_len = body && body->slen > 0 ?
        (unsigned)body->slen : 0;
    if (event->body_len >= sizeof(event->body))
        event->body_len = sizeof(event->body) - 1;
}

static void on_status(pjsua_call_id call, const pj_str_t *to,
                      const pj_str_t *body, void *data,
                      pjsip_status_code code, const pj_str_t *reason)
{
    pjgo_event *event = new_event(2);
    PJ_UNUSED_ARG(call);
    PJ_UNUSED_ARG(body);
    if (!event) return;
    event->id = (unsigned)(uintptr_t)data;
    event->code = code;
    copy_text(event->peer, sizeof(event->peer), to);
    copy_text(event->body, sizeof(event->body), reason);
    event->body_len = (unsigned)strlen(event->body);
}

int pjgo_start(unsigned port, const char *username)
{
    pjsua_config config;
    pjsua_logging_config log;
    pjsua_media_config media;
    pjsua_transport_config transport;
    pjsua_transport_id tid;
    pjsua_acc_config acc;
    pj_status_t status;
    char identity[256];
    const char *log_level = getenv("PJGO_LOG_LEVEL");
    if (created) return PJ_EEXISTS;
    head = count = 0;
    overflow = 0;
    status = pjsua_create();
    if (status != PJ_SUCCESS) return status;
    created = 1;
    pjsua_config_default(&config);
    config.thread_cnt = 0;
    config.cb.on_pager2 = on_message;
    config.cb.on_pager_status = on_status;
    pjsua_logging_config_default(&log);
    log.level = log.console_level = 1;
    if (log_level && strlen(log_level) == 1 &&
        log_level[0] >= '0' && log_level[0] <= '5')
        log.level = log.console_level = (unsigned)(log_level[0] - '0');
    pjsua_media_config_default(&media);
    media.has_ioqueue = PJ_FALSE;
    media.thread_cnt = 0;
    status = pjsua_init(&config, &log, &media);
    if (status != PJ_SUCCESS) goto fail;
    pjsip_cfg()->endpt.disable_tcp_switch = PJ_TRUE;
    pjsua_transport_config_default(&transport);
    transport.port = port;
    transport.bound_addr = pj_str("127.0.0.1");
    transport.public_addr = pj_str("127.0.0.1");
    status = pjsua_transport_create(PJSIP_TRANSPORT_UDP, &transport, &tid);
    if (status != PJ_SUCCESS) goto fail;
    pjsua_acc_config_default(&acc);
    pj_ansi_snprintf(identity, sizeof(identity), "sip:%s@127.0.0.1:%u",
                     username, port);
    acc.id = pj_str(identity);
    acc.transport_id = tid;
    acc.register_on_acc_add = PJ_FALSE;
    status = pjsua_acc_add(&acc, PJ_TRUE, &account);
    if (status != PJ_SUCCESS) goto fail;
    status = pjsua_start();
    if (status != PJ_SUCCESS) goto fail;
    return PJ_SUCCESS;
fail:
    pjgo_stop();
    return status;
}

int pjgo_send(const char *uri, const char *body, unsigned len, unsigned id)
{
    pj_str_t dest = pj_str((char *)uri);
    pj_str_t mime = pj_str("text/plain");
    pj_str_t text;
    text.ptr = (char *)body;
    text.slen = len;
    return pjsua_im_send(account, &dest, &mime, &text, NULL,
                         (void *)(uintptr_t)id);
}

int pjgo_poll(unsigned timeout_ms)
{
    int result = pjsua_handle_events(timeout_ms);
    if (overflow) return -1;
    return result < 0 ? result : 0;
}

int pjgo_pop(pjgo_event *event)
{
    if (!count) return 0;
    *event = queue[head];
    head = (head + 1) % QUEUE_SIZE;
    --count;
    return 1;
}

void pjgo_stop(void)
{
    if (created) pjsua_destroy();
    created = 0;
    account = PJSUA_INVALID_ID;
}

const char *pjgo_version(void) { return pj_get_version(); }
void pjgo_error(int status, char *buffer, unsigned size)
{
    pj_strerror(status, buffer, size);
}
