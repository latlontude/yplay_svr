#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import re

inst = MySQLdb.connect(db='yplay',host='10.66.190.26',port=7706,user='root',passwd='frankshi@0928#')
cursor = inst.cursor()
cursor.execute('set names utf8')


f = file("q.lst")

for line in f.readlines():
    txt   = line.strip()
    res = re.split(";", txt)

    if len(res) <= 2:
        print "invalid line %s"%(txt)
        exit(1)

    qid = res[0]

    for idx in range(len(res)):

        #类型编号 跟icon有关
        if idx == 0:
            qid = res[idx]
            continue

        #类型名称
        if idx == 1:
            continue

        qtext = res[idx]

        if len(qtext.strip()) == 0:
            continue

        qurl = "%d.png"%(int(qid))
        print qid, qtext, qurl
        #continue

        status = 0
        ts = int(time.time())

        cursor.execute('insert into questions values(%s,%s,%s,%s,%s)',(0,qtext,qurl, status, ts))
        inst.commit()

