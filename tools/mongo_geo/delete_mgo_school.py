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
import sys

user="mongouser"
password="yeejaymgo501"
host="10.66.205.15:27017"
uri = "mongodb://%s:%s@%s"% (user, password, host)

if __name__ == "__main__":

    if len(sys.argv) != 2:
        print "use %s schoolId"%(sys.argv[0])
        exit(1)

    schoolId = sys.argv[1]

    print "delete schoolId %s"%(schoolId)

    mgoclient = MongoClient(uri)
    c = mgoclient.yplay.schools

    c.remove({"schoolId":schoolId},{"justOne":1})
