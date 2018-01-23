#!/bin/python
# -*- coding:utf-8 -*- 

import json
import os
import MySQLdb
import sys
import re
import urllib2

if __name__ == "__main__":

    if len(sys.argv) != 2:
        print "use %s schoolId"%(sys.argv[0])
        exit(1)

    schoolId = sys.argv[1]

    print "delete schoolId %s"%(schoolId)

    url="http://localhost:9200/yplay/schools/%s"%(schoolId)
    request = urllib2.Request(url)
    #request.add_header('Content-Type', 'application/json')
    request.get_method = lambda:'DELETE'
    request = urllib2.urlopen(request)
    res =  request.read()
    print res
