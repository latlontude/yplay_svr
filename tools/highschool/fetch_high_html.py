#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import urllib2
import re
import json

f = file("../citys/citys.lst")

citys = {}

for line in f.readlines():
    l = line.strip()

    r = l.split(":")
    if len(r) != 2:
        print "invalid line %s"%(line)
        exit(1)

    cityId   = r[0]
    cityName = r[1]

    if r[0].startswith("1101"):
        cityId = "1101"
        cityName = "北京"
    elif r[0].startswith("1102"):
        cityId = "1102"
        cityName = "北京"
    elif r[0].startswith("1201"):
        cityId = "1201"
        cityName = "天津"
    elif r[0].startswith("1202"):
        cityId = "1202"
        cityName = "天津"
    elif r[0].startswith("3101"):
        cityId = "3101"
        cityName = "上海"
    elif r[0].startswith("3102"):
        cityId = "3102"
        cityName = "上海"
    elif r[0].startswith("5001"):
        cityId = "5001"
        cityName = "重庆"
    elif r[0].startswith("5002"):
        cityId = "5002"
        cityName = "重庆"

    citys[cityId] = cityName

for cid in citys:
    #print cid, citys[cid]

    path = "%s.html"%(cid)
    if os.path.exists(path):
        continue

    url = "http://support.renren.com/highschool/%s.html"%(cid)
    os.system('wget %s'%(url))
