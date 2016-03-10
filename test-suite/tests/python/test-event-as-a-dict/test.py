from __future__ import print_function


def run(event, context):
    print ("event type = ", type(event))
    print ("event = ", event)
    print ("event['a'] = ", event['a'])
    print ("event['b'] = ", event['b'])
    print ("event['null'] = ", event['null'])
    print ("event['obj'] = ", event['obj'])
    print ("event['arr'] = ", event['arr'])
