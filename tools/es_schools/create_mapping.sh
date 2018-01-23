#!/bin/bash

#curl -XDELETE "http://127.0.0.1:9200/yeejay/book/_mapping" 
#curl -XDELETE "http://127.0.0.1:9200/yeejay/page/_mapping" 
#curl -XDELETE "http://localhost:9200/yeejay"

#curl -XPUT "http://localhost:9200/yplay"

curl -XDELETE "http://127.0.0.1:9200/yplay/schools/_mapping" 
curl -XPUT "http://127.0.0.1:9200/yplay/_mapping/schools?pretty" -d '
{
    "properties":{
        "schoolId":{
            "type": "integer"
         },
	     "schoolType":{
             "type": "integer"
	     },
	     "school":{
              "type": "string",
              "analyzer":"ik_max_word"
	     },
         "country":{
             "type": "string",
             "index":"not_analyzed"
         },
         "province":{
             "type": "string",
             "index":"not_analyzed"
         },
         "city":{
             "type": "string",
             "index":"not_analyzed"
	     },
	     "latitude":{
             "type": "double"
         },
         "longitude":{
             "type": "double"
	     },
	     "status":{
             "type": "integer"
	     },
	     "ts":{
             "type": "integer"
	     }
    }
}'
