import sys
#sys.path.append('/root/jd/mon_jd/bot_plugins/service')
import bot_plugins.service.nolanjdc as nolanjdc
import bot_plugins.service.read_conf as conff
from nonebot.command import CommandSession
from nonebot.experimental.plugin import on_command
import time, re


@on_command('登陆', aliases=('登录', 'login'))
#@on_command('ping', permission=lambda sender: sender.is_superuser)
async def ssss(session: CommandSession):
	global url
	url=conff.read_conf()
	#if len(s)!=0:
		#await session.send('有人正在和我说话，你等5分钟再来吧！')
	#else:
	ipone = session.current_arg_text.strip()
	if not ipone:
		await session.send('请在5分钟内结束这次对话！')
		time.sleep(1)
		ipone = await session.aget(prompt='嗨！ 请输入手机号码:')
		phone = re.compile('^(13(0|1|2|3|7|8|9|6|5|4)|16(0|1|2|3|7|8|9|6|5|4)|17(0|8|5|6|2|3|7)|18(0|1|2|3|4|5|6|7|8|9)|15(1|2|3|5|6|7|8|9)|19(0|1|3|8|9))\d{8}$')
		if re.match(phone, ipone):
				await session.send('请等待。。。。')
				send_code = nolanjdc.sendsms(ipone)
				time.sleep(8)
				if '安全验证' in send_code:
					#ipone1 = await session.aget(prompt='嗨！ 再输入手机号码:')
					n=1
					while n<=3:
						await session.send(f'正在进行第{n}次滑块验证中。。。。')
						succ=nolanjdc.AutoCaptcha(ipone)

						time.sleep(3)
						await session.send(f'{succ}')
						if succ != True:
							time.sleep(1)
							if n==3:
								await session.send(f'第{n}次滑块验证失败,退出程序。去{url}登陆吧！')
								break
							await session.send(f'第{n}次滑块验证失败，进入下一次验证！')
							n += 1
							continue
						else:
							code = await session.aget(prompt='嗯！验证码该发我了:')
							if len(code) != 6:
								await session.send('丢！你数数你验证码多少位！再见')
								break
							else:
								if code:
									qq = await session.aget(prompt='嗯！QQ该发我了:')
									msg1 = nolanjdc.VerifyCode(ipone,qq,code)
									if '成功' in msg1:
										await session.send('恭喜上车成功!')
									else:
										
										if '验证码输入错误' in msg1:
											code1 = await session.aget(prompt='验证码错误，请重新输入验证码:')
											msg1 = nolanjdc.VerifyCode(ipone,qq,code1)
											await session.send(f'{msg1}')
										else:
											await session.send(f'{msg1}')
									break
				else:
					await session.send('安全验证没有了，请联系管理员！')
		else:
			await session.send('你发的什么东西？我感觉不是手机号码.......')
			#await session.send('也可能是号码段没有加进数据库，如果确定号码无误请联系管理员添加！')
			await session.send('拜拜了您！')
