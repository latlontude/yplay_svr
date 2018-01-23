#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import re
import random

inst = MySQLdb.connect(db='yplay',host='10.66.190.26',port=7706,user='root',passwd='frankshi@0928#')
cursor = inst.cursor()
cursor.execute('set names utf8')

cursor.execute('select qtext, qiconUrl, status, ts from questions2')

questions = []
for e in cursor.fetchall():
    qtext, qiconUrl, status, ts = e

    questions.append((qtext, qiconUrl, status, ts))

random.shuffle(questions)

for e in questions:
    qtext, qiconUrl, status, ts = e

    print qtext, qiconUrl, status, ts
    cursor.execute('insert into questions values(%s,%s,%s,%s,%s)',(0, qtext, qiconUrl, status, ts))
    inst.commit()
