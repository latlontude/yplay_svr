#!/bin/python
# -*- coding:utf-8 -*- 

import os
import os.path
import MySQLdb
import time
import re

provinces = {
        "11":"北京",
        "12":"天津",
        "13":"河北",
        "14":"山西",
        "15":"内蒙古",
        "21":"辽宁",
        "22":"吉林",
        "23":"黑龙江",
        "31":"上海",
        "32":"江苏",
        "33":"浙江",
        "34":"安徽",
        "35":"福建",
        "36":"江西",
        "37":"山东",
        "41":"河南",
        "42":"湖北",
        "43":"湖南",
        "44":"广东",
        "45":"广西",
        "46":"海南",
        "50":"重庆",
        "51":"四川",
        "52":"贵州",
        "53":"云南",
        "54":"西藏",
        "61":"陕西",
        "62":"甘肃",
        "63":"青海",
        "64":"宁夏",
        "65":"新疆"}

def get_province_by_code(code):

    code2 = code[0:2]

    if code2 not in provinces:
        return ""

    return provinces[code2]

def main():

    inst = MySQLdb.connect(db='yplay',host='10.66.190.26',port=7706,user='root',passwd='frankshi@0928#')
    cursor = inst.cursor()
    cursor.execute('set names utf8')

    f = file("result.lst")

    for line in f.readlines():
        a = re.split("\s+", line.strip())

        if len(a) != 3 and len(a) != 4:
            print "invalid line %s"%(line)
            exit(1)

        if len(a) == 4:
            code, city, district, school = a
        else:
            code, city, school = a
            district = ""

        province = get_province_by_code(code)
        if len(province) == 0:
            print "invalid province code(%s)"%(code)
            exit(1)

        print province, city, district, school

        cursor.execute('insert into schoolsOrgs values(%s,%s,%s,%s,%s,%s,%s,%s,%s)',(0,2,school, "中国",province, city, district, 0.0, 0.0))
        inst.commit()

if __name__ == "__main__":
    main()

