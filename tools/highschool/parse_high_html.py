#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import urllib2
import re
import HTMLParser


html_parser = HTMLParser.HTMLParser()

f = file("4201.html")
lines = []
for line in f.readlines():
    lines.append(line)

body = "".join(lines)
p = re.compile('SchoolComponent.tihuan\(\'(.*?)\'\)">(.*?)</a>')
l = p.findall(body)

qus = [] #qu_4201
quNames = {} #qu_4201->汉口

for e in l:
    if len(e) != 2:
        print "invalid html"
        exit(1)

    qu = html_parser.unescape(e[1]).encode('utf-8')
    #city_qu_420102  江汉区
    quNames[e[0]] = qu
    qus.append(e[0])

p1 = re.compile('SchoolComponent.cl_school\(event,(.*?)>(.*?)</a></li>')
l1 = p1.findall(body)

schools = [] #schoolNames
for e1 in l1:
    schoolName = html_parser.unescape(e1[1]).encode('utf-8')
    schools.append(schoolName)
    #print schoolName

#for e in schools:
#    print e
#exit(1)
#print schools

quLineNumber = {} #qu的行号

idx = 0
for line in lines:
    for e in qus:
        if line.find(e) > 0:
            quLineNumber[e] = idx
            break
    idx += 1

#print quLineNumber

quSchoolCnt = {}
curTotal = 0
for i in range(len(qus)):
    qu = qus[i]

    if i < len(qus) -1:
        cnt = quLineNumber[qus[i+1]]-quLineNumber[qus[i]]-1
        curTotal += cnt
        quSchoolCnt[qu] = cnt

quSchoolCnt[qus[len(qus) -1]] = len(schools) - curTotal

quSchoolIdx = {}
curTotal = 0

for qu in qus:
    cnt = quSchoolCnt[qu]
    quSchoolIdx[qu] = (curTotal, curTotal + cnt -1)

    print qu, curTotal, curTotal + cnt -1

    curTotal += cnt

quSchools = {}
for qu in qus:
    s, e = quSchoolIdx[qu]

    quSchools[qu] = schools[s:e]


for qu in qus:
    #print qu, quSchools[qu]

    for school in quSchools[qu]:
        print qu, quNames[qu], school

#for qu in quSchoolCnt:
#    quName = quNames[qu]
#    cnt = quSchoolCnt[qu]

#    print quName, cnt
