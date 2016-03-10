from __future__ import print_function
import time


def run(event, context):
    for attr in dir(context):
       if hasattr( context, attr ):
           print( "context.%s is an attribute" % (attr))
           if callable(getattr(context, attr)) :
               print( "context.%s is callable" % (attr))
