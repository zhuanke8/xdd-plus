import json
import re
import os
path=os.path.join(os.path.dirname(__file__) +'/../../../conf/config.yaml'
)
def read_conf():
	with open(path,'r') as f:
		f=f.read()
		ff=re.findall(r'(?<=Jdcurl:).+',f)[0].replace(' ','')
		return ff
#print(read_conf())