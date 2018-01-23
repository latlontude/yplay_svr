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

cursor.execute('select schoolId, schoolType, schoolName, province from schoolsOrgs where schoolType = 3 and latitude = 0')

baiduAk="YYR9I0YW5xSMrm2miKYLt7fxuxfGtSrT"

for e in cursor.fetchall():
    schoolId, schoolType, school, province = e

    if schoolType == 1 or schoolType == 2:
        tag = "中学"

    if schoolType == 3:
        tag = "高等院校"

    print tag, schoolId, schoolType, school, province

    qEnc = quote("%s"%(school))
    tagEnc = quote("%s"%(tag))
    provinceEnc = quote("%s"%(province))

    queryUrl = 'http://api.map.baidu.com/place/v2/search?query=%s&tag=%s&region=%s&output=json&ak=%s&coord_type=1'%(qEnc, tagEnc, provinceEnc, baiduAk)

    rsp  = urllib2.urlopen(queryUrl)
    body = rsp.read()

    o = json.loads(body)

    if o["status"] != 0:
        print "baidu ret error, schoolId %s, school %s, errmsg %s"%(schoolId, school, o["message"])
        #break
        continue

    if "results" not in o:
        print "baidu ret results empty, schoolId %s, school %s, queryUrl %s"%(schoolId, school, queryUrl)
        #break
        continue

    if len(o["results"]) == 0:
        print "baidu ret empty, schoolId %s, school %s"%(schoolId, school)
        #break
        continue

    if "location" not in o["results"][0]:
        print "baidu ret location empty, schoolId %s, school %s, queryUrl %s"%(schoolId, school, queryUrl)

        if "name" in o["results"][0]:
            city = o["results"][0]["name"]
            queryByCity(schoolId, school, schoolType, city)
        #break
        continue

    loc = o["results"][0]["location"]

    x = loc["lat"]
    y = loc["lng"]

    cursor2.execute('update schoolsOrgs set latitude = %s, longitude = %s where schoolId = %s',(x, y, schoolId))
    inst.commit()

'''
def queryByCity(schoolId, school, schoolType, city):

    if schoolType == 1 or schoolType == 2:
        tag = "中学"

    if schoolType == 3:
        tag = "高等院校"

    qEnc = quote("%s"%(school))
    tagEnc = quote("%s"%(tag))
    cityEnc = quote("%s"%(city))

    queryUrl = 'http://api.map.baidu.com/place/v2/search?query=%s&tag=%s&region=%s&output=json&ak=%s&coord_type=1'%(qEnc, tagEnc, cityEnc, baiduAk)

    rsp  = urllib2.urlopen(queryUrl)
    body = rsp.read()

    o = json.loads(body)

    if o["status"] != 0:
        print "baidu ret error, schoolId %s, school %s, errmsg %s"%(schoolId, school, o["message"])
        #break
        return

    if "results" not in o:
        print "baidu ret results empty, schoolId %s, school %s, queryUrl %s"%(schoolId, school, queryUrl)
        #break
        return

    if len(o["results"]) == 0:
        print "baidu ret empty, schoolId %s, school %s"%(schoolId, school)
        #break
        return

    if "location" not in o["results"][0]:
        print "baidu ret location empty, schoolId %s, school %s, queryUrl %s"%(schoolId, school, queryUrl)
        #break
        return

    loc = o["results"][0]["location"]

    x = loc["lat"]
    y = loc["lng"]

    cursor2.execute('update schoolsOrgs set latitude = %s, longitude = %s where schoolId = %s',(x, y, schoolId))
    inst.commit()
'''
