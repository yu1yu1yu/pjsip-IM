import argparse
import pjsua2 as pj

class Account(pj.Account):
    def __init__(self):
        pj.Account.__init__(self)
    
    def onRegState(self, prm):
        ai = self.getInfo()
        
        # print('Registration status=',prm.code,'(',prm.reason,')')
        print(f"Account State: {ai.regStatusText}, Code: {prm.code}, Reason: {prm.reason}")

    # 收到对方发来的 MESSAGE
    def onInstantMessage(self, prm):
        print(f"\n收到消息：{prm.fromUri} -> {prm.msgBody}", flush=True)
        print(f"内容类型：{prm.contentType}", flush=True)

    # 发送结果也由 Account 接收，状态码字段是 code
    def onInstantMessageStatus(self, prm):
        print(f"\n发送结果：{prm.code} {prm.reason}，接收者：{prm.toUri}", flush=True)
        if prm.code == 200:
            print("对端已接受消息（不代表已读）", flush=True)


class MyBuddy(pj.Buddy):
    def __init__(self, acc: Account, target_uri: str):
        pj.Buddy.__init__(self)
        self.acc = acc
        buddy_cfg = pj.BuddyConfig()
        buddy_cfg.uri = target_uri
        buddy_cfg.subscribe = False
        self.create(acc, buddy_cfg)

    def send_text(self, text: str):
        prm = pj.SendInstantMessageParam()
        prm.contentType = "text/plain"
        prm.content = text
        self.sendInstantMessage(prm)

def create_sip_account(ep: pj.Endpoint, sip_uri: str) -> Account:
    """
    创建本地账号，不注册
    :param ep: endpoint实例
    :param sip_uri: 身份URI，例如 sip:alice@127.0.0.1:5060
    """
    acc_cfg = pj.AccountConfig()
    # 设置SIP身份
    acc_cfg.idUri = sip_uri

    # ========== 关键：关闭自动注册，不配置注册服务器 ==========
    acc_cfg.regConfig.registerOnAdd = False
    # 不设置 registrarUri，不向任何SIP服务器注册
    acc_cfg.regConfig.registrarUri = ""

    acc = Account()
    acc.create(acc_cfg, True)
    return acc

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--port',type=int,default=5060)
    parser.add_argument('--username',type=str,default='alice')
    parser.add_argument('--to', help='目标 SIP URI；不填写时仅接收')
    parser.add_argument('--text', default='你好，这是第一条 SIP 消息')
    args = parser.parse_args()
    print (args)
    if not 1<=args.port<=65535:
        print('Invalid port number')
        return
    
    ep=pj.Endpoint()
    created = False
    sip_account = None
    buddy = None
    try:
        ep.libCreate()
        created=True
        
        # configure endpoint
        ep_cfg=pj.EpConfig()
        ep_cfg.uaConfig.threadCnt=0
        ep_cfg.uaConfig.mainThreadOnly=True
        ep_cfg.logConfig.level=4
        ep_cfg.logConfig.consoleLevel=4
        ep.libInit(ep_cfg)
        
        #create UDP
        transport_cfg=pj.TransportConfig()
        transport_cfg.port=args.port
        transport_cfg.boundAddress='127.0.0.1'
        ep.transportCreate(pj.PJSIP_TRANSPORT_UDP,transport_cfg)
        ep.audDevManager().setNullDev()
        
        #start library
        ep.libStart()
        print('Started SIP endpoint')
        print('Press Ctrl-C to stop..')
        
        #create account
        sip_account = create_sip_account(ep, f'sip:{args.username}@127.0.0.1:{args.port}')
        print(f'账号已就绪：sip:{args.username}@127.0.0.1:{args.port}', flush=True)
        if args.to:
            buddy = MyBuddy(sip_account, args.to)
            buddy.send_text(args.text)
            print(f'已提交发送请求：{args.text}', flush=True)

        while True:
            ep.libHandleEvents(50)
        
    except KeyboardInterrupt:
        print('\nStopping..')
    
    except pj.Error as e:
        print('\nPJSIP error:', e.info())
        raise
    finally:
        # CPython 中先释放 Buddy，再关闭账号，最后销毁库
        buddy = None
        if sip_account is not None:
            sip_account.shutdown()
            sip_account = None
        if created:
            ep.libDestroy()
    
if __name__ == '__main__':
    main()