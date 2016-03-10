from __future__ import print_function

def run(event, context):
    for attr in dir(event):
       if hasattr( event, attr ):
           print( "event.%s is an attribute" % (attr))
           if callable(getattr(event, attr)) :
               print( "event.%s is callable" % (attr))
