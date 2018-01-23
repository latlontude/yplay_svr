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
from pymongo import MongoClient

user="mongouser"
password="yeejaymgo501"
host="10.66.205.15:27017"
uri = "mongodb://%s:%s@%s"% (user, password, host)

mgoclient = MongoClient(uri)
c = mgoclient.yplay.schools

inst = MySQLdb.connect(db='yplay',host='10.66.190.26',port=7706,user='root',passwd='frankshi@0928#')
cursor = inst.cursor()
cursor.execute('set names utf8')

cursor.execute('select schoolId, schoolType, schoolName, province, city, latitude, longitude  from schoolsOrgs where latitude > 0 and schoolType = 3 and province = "广东"')

for e in cursor.fetchall():
    schoolId, schoolType, schoolName, province, city, lat, lon = e
    print schoolId, schoolType, schoolName, province, city, lat, lon

    doc={"_id":schoolId, "type":schoolType, "name":schoolName, "province":province, "city":city, "loc":{"type":"Point", "coordinates":[lon,lat]}}
    c.insert(doc)
