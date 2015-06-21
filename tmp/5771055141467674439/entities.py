#! /usr/bin/env python3
# -*- coding:utf-8 -*-

import sys
import json
import datetime
from urllib import parse, request

key = "2ef2491bc08f6c00eba4c413a19089d3"


def analyze(text, entity_type):
    query = parse.urlencode({"text": text}, encoding="utf-8", errors="replace")
    params = query.encode("utf-8", "replace")
    req = request.Request("http://api.syllabs.com/v0/entities", params)
    req.add_header("API-Key", key)
    r = request.urlopen(req)
    resp = r.read()
    data = json.loads(resp.decode("utf-8", "replace"))
    entities = data["response"]["entities"]
    return entities.get(entity_type, [])


def display(entities):
    entities.sort(key=lambda ent: int(ent["count"]), reverse=True)
    for entity in entities:
        print("%s: %s" % (entity["text"], entity["count"]))


def main():
    start_time = datetime.datetime.now()
    texts = ""
    for line in sys.stdin:
        texts += line

    entities = analyze(texts, "Person")

    display(entities)
    elsapsed_time = datetime.datetime.now() - start_time
    # print(elsapsed_time)


if __name__ == '__main__':
    main()
