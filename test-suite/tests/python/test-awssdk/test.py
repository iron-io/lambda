from __future__ import print_function
import boto3
import uuid


def run(event, context):
    print ("Test start")
    s3 = boto3.resource('s3')

    print ("Creating object ...")
    obj = s3.Object(
        bucket_name=event['bucket'],
        key='myKey-' + str(uuid.uuid4())
        )

    print ("Putting object value ...")
    obj.put(Body='Hello!')

    print ("Deleting object ...")
    obj.delete()
    print ("Accessing S3 has been succeeded")
