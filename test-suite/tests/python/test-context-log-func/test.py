from __future__ import print_function
import inspect


def run(event, context):
    print ("context.log args   = ", inspect.getargspec(context.log))
    logResult = context.log('a sample message\n')
    print ("context.log result = ", logResult)
