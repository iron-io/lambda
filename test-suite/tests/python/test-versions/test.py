from __future__ import print_function
import sys


def run(event, context):
    print ("python version {major}.{minor}"
        .format(major=sys.version_info[0], minor=sys.version_info[1]))

    modules = [
        'boto3',
        'botocore',
        {'key':'bsddb', 'packages': ['bsddb', 'bsddb3' ]},
        'urlgrabber'
    ]
    for entry in modules:
        moduleName = getModuleName(entry)
        module = getModule(entry)
        print ("{moduleName} version {version}"
            .format(moduleName=moduleName, version=module.__version__))


def getModuleName(entry):
    return (isinstance(entry, dict) and entry['key']) or entry


def getModule(entry):
    if isinstance(entry, dict):
        for name in entry['packages']:
            try:
                return getModuleByName(name)
            except:
                pass
        raise Exception('no module %s found' % entry['key'])
    else:
        return getModuleByName(entry)


def getModuleByName(name):
    return sys.modules.get(name, None) or __import__(name)
