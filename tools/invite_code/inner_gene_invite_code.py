#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import re
import redis
import random

r = redis.Redis(host="10.66.137.165", port = 6379, password="yeejay501")

random.seed(int(time.time()))

nums={}
for i in range(1,20):
    num = random.randint(100001,999999)

    nums[num] = 1

for num, v in nums.items():

    key="22_%d"%(num)
    val = r.get(key)

    if val is not None:
        continue

    print num
    r.set(key,"1",ex=3600*24*7)
