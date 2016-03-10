from __future__ import print_function
import sys
import boto3


def run(event, context):
    print ("python version {major}.{minor}"
        .format(major=sys.version_info[0], minor=sys.version_info[1]))
    print ("boto3 version {version}"
        .format(version=boto3.__version__))
