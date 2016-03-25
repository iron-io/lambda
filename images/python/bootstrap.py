from __future__ import print_function
import sys
import imp
import os
import json
import string
import logging.config
import time
import uuid


debugging = False

debugging and print ('loading...')


class Context(object):
    function_name = None
    function_version = None
    memory_limit_in_mb = None
    aws_request_id = None

    invoked_function_arn = None
    log_group_name = None
    log_stream_name = None
    identity = None
    client_context = None

    def __init__(self):
        self.function_name = getAWS_LAMBDA_FUNCTION_NAME()
        self.function_version = getAWS_LAMBDA_FUNCTION_VERSION()
        self.aws_request_id = getREQUEST_ID()
        self.memory_limit_in_mb = int(getTASK_MAXRAM() / 1024 / 1024)

    def get_remaining_time_in_millis(self):
        remaining = plannedEnd - time.time()
        if remaining < 0:
            remaining = 0
        return remaining * 1000

    def log(self, msg):
        print (msg, end='')
        return


class DynaCallerError(Exception):
    pass


class DynaCaller(object):

    def __init__(self, module, name):
        self.moduleName = module
        self.funcName = name
        self.module = None

    def locateFunc(self):
        loaders = [self.locateModuleInMountFolder, self.locateModuleDefault]

        for i, loader in enumerate(loaders):
            if not (self.module is None):
                break;
            try:
                self.module = loader()
            except Exception as e:
                if i == len(loaders) - 1:
                    raise

        if self.module is None:
            raise DynaCallerError("Failed to locate a module")

        self.func = getattr(self.module, self.funcName)
        if self.func is None:
            raise DynaCallerError("Failed to locate a function inside module")

    def locateModuleDefault(self):
        return __import__(self.moduleName)

    def locateModuleInMountFolder(self):
        mountModuleLocation = '/mnt/' + self.moduleName + '.py'
        if not os.path.isfile(mountModuleLocation):
            return None
        return imp.load_source(self.moduleName, mountModuleLocation)

    def call(self, payload, context):
        return self.func(payload, context)


class UTCFormatter(logging.Formatter):
    converter = time.gmtime

    def __init__(self, fmt=None, datefmt=None):
        super(UTCFormatter, self).__init__(fmt, datefmt)

    def formatTime(self, record, datefmt=None):
        ct = self.converter(record.created)
        if datefmt:
            s = time.strftime(datefmt, ct)
        else:
            t = time.strftime("%Y-%m-%dT%H:%M:%S", ct)
            s = "%s.%03dZ" % (t, record.msecs)
        return s


def stopWithError(msg):
    print ("ERROR:", msg, file=sys.stderr)
    raise SystemExit(1)


def getPAYLOAD_FILE():
    return os.environ.get('PAYLOAD_FILE')


def getTASK_TIMEOUT():
    return os.environ.get('TASK_TIMEOUT') or 3600


def getTASK_MAXRAM():
    # IronWorker uses MAXMEM, Hybrid uses MAXRAM.
    maxmemFlag = os.environ.get('TASK_MAXRAM') or '300m'
    suffix = maxmemFlag[-1:]
    theNumber = int(maxmemFlag[:-1])
    factor = 1024
    valueInBytes = {
        'b': theNumber,
        'k': theNumber * factor,
        'm': theNumber * factor * factor,
        'g': theNumber * factor * factor * factor,
        }.get(suffix, theNumber)
    return valueInBytes


def getAWS_LAMBDA_FUNCTION_NAME():
    return os.environ.get('AWS_LAMBDA_FUNCTION_NAME')


def getAWS_LAMBDA_FUNCTION_VERSION():
    return os.environ.get('AWS_LAMBDA_FUNCTION_VERSION')


def getREQUEST_ID():
    return os.environ.get('TASK_ID') or uuid.uuid4()


def getHandlerName():
    if len(sys.argv) > 1:
        return sys.argv[1]
    return None


def configLogging(context):

    # RequestIdFilter is used to add request_id field value into log line. More details could be found on the following links:
    #  https://docs.python.org/2/howto/logging-cookbook.html#filters-contextual
    #  https://docs.python.org/2/howto/logging-cookbook.html#an-example-dictionary-based-configuration

    class RequestIdFilter(logging.Filter):
        def filter(self, record):
            record.request_id = context.aws_request_id
            return True

    loggingConfig = {
        'version': 1,
        'disable_existing_loggers': True,
        'formatters': {
            'standard': {
                '()': UTCFormatter,
                'format': '[%(levelname)s]\t%(asctime)s\t%(request_id)s\t%(message)s'
            },
        },
        'handlers': {
            'default': {
                'class': 'logging.StreamHandler',
                'formatter': 'standard',
                'stream': 'ext://sys.stdout'
            },
        },
        'filters':{
            'request_id' :{
                '()' : RequestIdFilter
            }
        },
        'loggers': {
            '': {
                'handlers': ['default'],
                'filters': ['request_id'],
                'propagate': True
            },
        }
    }
    logging.config.dictConfig(loggingConfig)

plannedEnd = time.time() + getTASK_TIMEOUT()

debugging and print ('os.environ      = ', os.environ)
debugging and print ('/mnt content    = ', os.listdir("/mnt"))
debugging and print ('pwd dir content = ',
    os.listdir(os.path.dirname(os.path.realpath(__file__))))

context = Context()
debugging and print ('context created')

configLogging(context)
debugging and print ('config loaded')

payloadFileName = getPAYLOAD_FILE()
debugging and print ('PAYLOAD_FILE = ', payloadFileName)

handlerName = getHandlerName()
debugging and print ('handlerName = ', handlerName)

if handlerName is None:
    stopWithError("handlerName arg is not specified")
if payloadFileName is None:
    stopWithError("PAYLOAD_FILE variable is not specified")

if not os.path.isfile(payloadFileName):
    stopWithError("No payload present")

handlerParts = string.rsplit(handlerName, ".", 2)

if len(handlerParts) < 2:
    stopWithError("handlerName arg should be specified " +
        "in format 'moduleName.functionName'")

moduleName = handlerParts[0]
funcName = handlerParts[1]

if moduleName is None:
    stopWithError("Module name is not defined")
if funcName is None:
    stopWithError("Function name is not defined")

try:
    with file(payloadFileName) as f:
        payload = f.read()
except:
    stopWithError("Failed to read {payloadFileName}".format(payloadFileName=payloadFileName))

debugging and print ('payload loaded')

try:
    payload = json.loads(payload)
except:
    debugging and print ('payload is ') and print (payload)
    stopWithError('Payload is not JSON')

debugging and print ('payload parsed as JSON')

caller = DynaCaller(moduleName, funcName)

try:
    caller.locateFunc()
except Exception as e:
    print (e, file=sys.stderr)
    stopWithError("Failed to locate {module}.{func}"
        .format(module=moduleName, func=funcName))

debugging and print ('handler found')

try:
    result = caller.call(payload, context)
    #FIXME where to put result in async mode?
except Exception as e:
    stopWithError(e)

debugging and print ('done')
