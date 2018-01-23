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
cursor.execute('set names utf8')
cursor2 = inst.cursor()

cursor.execute('select schoolId, schoolName, province from schoolsOrgs where schoolType = 3 and latitude = 0')

geos = {}

for e in cursor.fetchall():
    schoolId, school, province = e

    print schoolId, school, province

    q = "%s%s"%(province, school)
    qe = quote("%s%s"%(province, school))

    rsp  = urllib2.urlopen('http://apis.map.qq.com/jsapi?qt=geoc&addr=%s&key=FBOBZ-VODWU-C7SVF-B2BDI-UK3JE-YBFUS&output=jsonp&pf=jsapi&ref=jsapi&cb=qq.maps._svcb2.geocoder0'%(qe))
    body = rsp.read()

    p = re.compile('qq.maps._svcb2.geocoder0&&qq.maps._svcb2.geocoder0\((.*)\)', re.S)

    l = p.findall(body.decode("gb2312").encode("utf-8"))

    for e in l:
        try:
            r = e

            o = json.loads(r, strict=False)
            code = o["info"]["error"]
            if code != 0:
                print "errno ",o["info"]["error"],  schoolId, school, province
                continue

            x = o["detail"]["pointx"]
            y = o["detail"]["pointy"]

            print schoolId, school, x, y
            geos[schoolId] = (x,y)
            cursor2.execute('update schoolsOrgs set latitude = %s, longitude = %s where schoolId = %s',(y, x, schoolId))
            inst.commit()

        except Exception,e:
            print "error schoolId %s,%s,%s,%s"%(schoolId, q, e.message, body)
            exit(1)
            time.sleep(10)

#for schoolId in geos:
#    x, y = geos[schoolId]

    #cursor.execute('update schoolsOrgs set latitude = %s, longitude = %s where schoolId = %s',(y, x, schoolId))
    #inst.commit()
