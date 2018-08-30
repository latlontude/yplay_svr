

curl -XDELETE "http://127.0.0.1:9200/interlocution/questions/_mapping"
curl -XPUT "http://127.0.0.1:9200/interlocution/_mapping/questions?pretty" -d'
{
    "properties": {
        "qid": {
            "type": "integer"
        },
        "qContent": {
            "type": "string",
            "analyzer":"ik_max_word"
        }
    }
}
'


curl -XDELETE "http://127.0.0.1:9200/interlocution/answers/_mapping"
curl -XPUT "http://127.0.0.1:9200/interlocution/_mapping/answers?pretty" -d'
{
    "properties": {
        "qid" : {
            "type":"integer"
        },
        "answerId": {
            "type": "integer"
        },
        "answerContent": {
            "type": "string",
            "analyzer":"ik_max_word"
        }
    }
}
'

curl -XDELETE "http://127.0.0.1:9200/interlocution/labels/_mapping"
curl -XPUT "http://127.0.0.1:9200/interlocution/_mapping/labels?pretty" -d'
{
    "properties": {
        "labelId" : {
            "type":"integer"
        },
        "labelName": {
            "type": "string",
            "analyzer":"ik_max_word"
        }
    }
}
'


curl -XGET "http://127.0.0.1:9200/interlocution/_mapping?pretty" -d'
{
  "interlocution" : {
    "mappings" : { }
  }
}'