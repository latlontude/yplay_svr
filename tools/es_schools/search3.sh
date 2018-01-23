#curl -XPOST 'http://localhost:9200/yeejay/_search?pretty' -d '{"query":{"bool":{"must":[{"term":{"title":"妈妈"}}],"must_not":[],"should":[]}},"from":0,"size":250,"sort":[],"aggs":{}}' 
curl -XPOST 'http://localhost:9200/yplay/_search?pretty' -d '
{
    "query":{
        "bool":{
            "must": [
                {"match":  {"schoolName":"实验"}}
            ]
         }
     },
    "from":0,
    "size":10,
    "sort":[],
    "aggs":{}
}' 


