#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import urllib2
import re
import json
from urllib import quote

inst = MySQLdb.connect(db='yplay',host='10.66.190.26',port=7706,user='root',passwd='frankshi@0928#')
cursor = inst.cursor()
cursor2 = inst.cursor()
cursor.execute('set names utf8')
cursor2.execute('set names utf8')

cursor.execute('select schoolId, schoolType, schoolName, province, city, latitude, longitude  from schoolsOrgs where latitude > 0 and schoolType = 3 and province = "广东"')

ts = int(time.time())
for e in cursor.fetchall():
    schoolId, schoolType, schoolName, province, city, lat, lon = e
    print schoolId, schoolType, schoolName, province, city, lat, lon

    cursor2.execute('insert ignore into schools values(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)',(schoolId, schoolType, schoolName,'中国',province,city,lat,lon,0,ts))
    inst.commit()
