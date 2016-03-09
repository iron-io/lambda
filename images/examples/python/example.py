import json


def run(event, context):
    print (json.dumps(event.__dict__, sort_keys=True, indent=4, separators=(',', ': ')))

