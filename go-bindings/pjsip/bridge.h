#ifndef PJGO_BRIDGE_H
#define PJGO_BRIDGE_H
#ifdef _WIN32
# ifdef PJGO_BUILD
#  define PJGO_API __declspec(dllexport)
# else
#  define PJGO_API __declspec(dllimport)
# endif
#else
# define PJGO_API
#endif

/* Endpoint calls must run on one OS thread; one endpoint per process. */
typedef struct pjgo_event {
    int kind; /* 1: message, 2: delivery result */
    int code;
    unsigned id;
    unsigned body_len;
    char peer[2048];
    char reply_to[2048];
    char content_type[256];
    char body[65536];
} pjgo_event;
PJGO_API int pjgo_start(unsigned port, const char *username);
PJGO_API int pjgo_send(const char *uri, const char *body, unsigned len,
                       unsigned id);
PJGO_API int pjgo_poll(unsigned timeout_ms);
PJGO_API int pjgo_pop(pjgo_event *event);
PJGO_API void pjgo_stop(void);
PJGO_API const char *pjgo_version(void);
PJGO_API void pjgo_error(int status, char *buffer, unsigned size);
#endif
