#!/bin/python
# -*- coding:utf-8 -*- 

import json
import os
import MySQLdb
import sys
import re
import urllib2

if __name__ == "__main__":

    #for i in range(NUM_PROCESS):
    inst = MySQLdb.connect(db='yplay',host='10.66.190.26',port=7706,user='root',passwd='frankshi@0928#')
    cursor = inst.cursor()
    cursor.execute('set names utf8')
    cursor.execute('select schoolId, schoolType, schoolName, country, province, city, latitude, longitude, status, ts from schools')

    schools = {}
    rsp = cursor.fetchall()
    for e in rsp:
        schoolId, schoolType, schoolName, country, province, city, latitude, longitude, status, ts = e
        schools[int(schoolId)] = (schoolId, schoolType, schoolName, country, province, city, latitude, longitude, status, ts)

    for schoolId in schools:
        v = schools[schoolId]
        schoolId, schoolType, schoolName, country, province, city, latitude, longitude, status, ts = v

        d = dict(schoolId=schoolId, schoolType=schoolType, school=schoolName, country=country, province=province, city=city, latitude=latitude, longitude=longitude, status = status, ts= ts)
        strd = json.dumps(d)

        url="http://localhost:9200/yplay/schools/%d"%(schoolId)
        request = urllib2.Request(url, strd)
        request.add_header('Content-Type', 'application/json')
        request.get_method = lambda:'PUT'
        request = urllib2.urlopen(request)
        res =  request.read()
        print res
