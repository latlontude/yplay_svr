#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import re
import redis

r = redis.Redis(host="10.66.137.165", port = 6379, password="yeejay501")

cursor = 0

while True:

    ncursor, keys = r.scan(cursor=cursor, count = 100)

    print ncursor, keys
    #break

    for key in keys:
        print key

    if int(ncursor) == 0:
        print "ncursor is zero"
        break

    cursor = int(ncursor)

