"""Validate native loading, lifecycle and codec access without network or microphone."""
import json
import platform
import sys
import pjsua2 as pj

ep = pj.Endpoint()
created = False
try:
    ep.libCreate()
    created = True
    config = pj.EpConfig()
    config.uaConfig.threadCnt = 0
    config.uaConfig.mainThreadOnly = True
    config.logConfig.level = 2
    config.logConfig.consoleLevel = 2
    ep.libInit(config)
    ep.audDevManager().setNullDev()
    ep.libStart()
    ep.libHandleEvents(10)
    codecs = [codec.codecId for codec in ep.codecEnum2()]
    assert any(c.startswith('PCMU/') for c in codecs), codecs
    assert hasattr(pj.Buddy, 'sendInstantMessage')
    assert hasattr(pj.Account, 'onInstantMessage')
    print(json.dumps({'result': 'PASS', 'python': sys.version.split()[0],
                      'architecture': platform.machine(), 'module': pj.__file__,
                      'pjsip': ep.libVersion().full, 'codecs': codecs}, indent=2))
finally:
    if created:
        ep.libDestroy()
