#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import urllib2
import re
import json

#http://s.xnimg.cn/a85689/nx/core/ui/renren/school.js
f = file("school.js.org")

provinces = {
        "1":"北京",
        "6":"天津",
        "21":"河北",
        "23":"山西",
        "26":"内蒙古",
        "5":"辽宁",
        "4":"吉林",
        "3":"黑龙江",
        "2":"上海",
        "8":"江苏",
        "9":"浙江",
        "7":"安徽",
        "17":"福建",
        "22":"江西",
        "16":"山东",
        "18":"河南",
        "11":"湖北",
        "13":"湖南",
        "12":"广东",
        "25":"广西",
        "30":"海南",
        "19":"重庆",
        "15":"四川",
        "24":"贵州",
        "20":"云南",
        "31":"西藏",
        "10":"陕西",
        "14":"甘肃",
        "28":"青海",
        "27":"宁夏",
        "29":"新疆"}

lines = []
for line in f.readlines():
    lines.append(line)

body = "".join(lines)

o = json.loads(body, strict=False)

#print o
#exit(1)

for e in o["provs"]:
    provid = str(e["id"])
    for s in e["univs"]:
        school = s["name"].encode("utf-8")

        if provid in provinces:
            provname = provinces[provid] 
        else:
            print "provid %s not found"%(provid)
            exit(1)

        print provname, school
