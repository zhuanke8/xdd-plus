import requests
import bot_plugins.service.read_conf as conff


def sendsms(phone):
    url=conff.read_conf()
    url=f'{url}/api/SendSMS'
    data={
        'Phone':phone,
        'qlkey':0
    }
    result=requests.post(url,json=data)
    return result.json()['message']
    
def AutoCaptcha(phone):
    url=conff.read_conf()
    url=f'{url}/api/AutoCaptcha'
    data={
        'Phone':phone
    }
    result=requests.post(url,json=data,timeout=360)
    return result.json()['success']

def VerifyCode(phone,qq,code):
    url=conff.read_conf()
    url=f'{url}/api/VerifyCode'
    qq=str(qq)
    data={
        'Phone':phone,
        'QQ':qq,
        'qlkey':0,
        'Code':code
    }
    result=requests.post(url,json=data)
    return result.json()['message']
